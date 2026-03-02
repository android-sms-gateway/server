package jwt

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jaevor/go-nanoid"
)

const (
	jtiLength = 21
)

type service struct {
	config Config

	tokens *Repository

	metrics *Metrics

	idFactory func() string
}

func New(config Config, tokens *Repository, metrics *Metrics) (Service, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	if tokens == nil {
		return nil, fmt.Errorf("%w: revoked storage is required", ErrInitFailed)
	}

	if metrics == nil {
		return nil, fmt.Errorf("%w: metrics is required", ErrInitFailed)
	}

	idFactory, err := nanoid.Standard(jtiLength)
	if err != nil {
		return nil, fmt.Errorf("can't create id factory: %w", err)
	}

	return &service{
		config: config,

		tokens: tokens,

		metrics: metrics,

		idFactory: idFactory,
	}, nil
}

func (s *service) generatePair(
	userID string,
	scopes []string,
	refreshScope string,
	accessTTL time.Duration,
) (*TokenPairInfo, error) {
	if userID == "" {
		return nil, fmt.Errorf("%w: user id is required", ErrInvalidParams)
	}

	if len(scopes) == 0 {
		return nil, fmt.Errorf("%w: scopes are required", ErrInvalidParams)
	}

	if accessTTL < 0 {
		return nil, fmt.Errorf("%w: access ttl must be positive", ErrInvalidParams)
	}

	if accessTTL == 0 {
		accessTTL = s.config.AccessTTL
	}

	now := time.Now()
	accessClaims := s.newClaims(userID, scopes, now.Add(min(accessTTL, s.config.AccessTTL)))
	refreshClaims := s.newRefreshClaims(userID, refreshScope, scopes, now.Add(s.config.RefreshTTL))

	accessToken, signErr := s.sign(accessClaims)
	if signErr != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", signErr)
	}

	refreshToken, signErr := s.sign(refreshClaims)
	if signErr != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", signErr)
	}

	return &TokenPairInfo{
		Access:  TokenInfo{ID: accessClaims.ID, Token: accessToken, ExpiresAt: accessClaims.ExpiresAt.Time},
		Refresh: TokenInfo{ID: refreshClaims.ID, Token: refreshToken, ExpiresAt: refreshClaims.ExpiresAt.Time},
	}, nil
}

func (s *service) GenerateTokenPair(
	ctx context.Context,
	userID string,
	scopes []string,
	refreshScope string,
	accessTTL time.Duration,
) (*TokenPairInfo, error) {
	var tokenInfo *TokenPairInfo
	var err error

	s.metrics.ObserveIssuance(func() {
		tokenInfo, err = s.generatePair(userID, scopes, refreshScope, accessTTL)
		if err != nil {
			return
		}

		if insertErr := s.tokens.Insert(ctx, newTokenModel(userID, tokenInfo.Access)); insertErr != nil {
			err = fmt.Errorf("failed to insert access token: %w", insertErr)
			return
		}
		if insertErr := s.tokens.Insert(ctx, newTokenModel(userID, tokenInfo.Refresh)); insertErr != nil {
			err = fmt.Errorf("failed to insert refresh token: %w", insertErr)
			return
		}
	})

	if err != nil {
		s.metrics.IncrementTokensIssued(StatusError)
	} else {
		s.metrics.IncrementTokensIssued(StatusSuccess)
	}

	return tokenInfo, err
}

func (s *service) RefreshTokenPair(ctx context.Context, refreshToken string) (*TokenPairInfo, error) {
	parsedToken, parseErr := jwt.ParseWithClaims(
		refreshToken,
		new(RefreshClaims),
		func(_ *jwt.Token) (any, error) {
			return []byte(s.config.Secret), nil
		},
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithIssuer(s.config.Issuer),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse token: %w", parseErr)
	}

	parsedClaims, ok := parsedToken.Claims.(*RefreshClaims)
	if !ok || !parsedToken.Valid {
		return nil, ErrInvalidToken
	}

	if len(parsedClaims.OriginalScopes) == 0 {
		return nil, ErrInvalidToken
	}

	tokenPair, err := s.generatePair(
		parsedClaims.UserID,
		parsedClaims.OriginalScopes,
		parsedClaims.Scopes[0],
		s.config.AccessTTL,
	)
	if err != nil {
		return nil, err
	}

	if rotateErr := s.tokens.RotateRefreshToken(
		ctx,
		parsedClaims.ID,
		newTokenModel(parsedClaims.UserID, tokenPair.Refresh),
		newTokenModel(parsedClaims.UserID, tokenPair.Access),
	); rotateErr != nil {
		return nil, rotateErr
	}

	return tokenPair, nil
}

func (s *service) ParseToken(ctx context.Context, token string) (*Claims, error) {
	var claims *Claims
	var err error

	s.metrics.ObserveValidation(func() {
		parsedToken, parseErr := jwt.ParseWithClaims(
			token,
			new(Claims),
			func(_ *jwt.Token) (any, error) {
				return []byte(s.config.Secret), nil
			},
			jwt.WithExpirationRequired(),
			jwt.WithIssuedAt(),
			jwt.WithIssuer(s.config.Issuer),
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
		)
		if parseErr != nil {
			err = fmt.Errorf("failed to parse token: %w", parseErr)
			return
		}

		parsedClaims, ok := parsedToken.Claims.(*Claims)
		if !ok || !parsedToken.Valid {
			err = ErrInvalidToken
			return
		}

		revoked, parseErr := s.tokens.IsRevoked(ctx, parsedClaims.ID)
		if parseErr != nil {
			err = parseErr
			return
		}
		if revoked {
			err = ErrTokenRevoked
			return
		}

		claims = parsedClaims
	})

	if err != nil {
		s.metrics.IncrementTokensValidated(StatusError)
	} else {
		s.metrics.IncrementTokensValidated(StatusSuccess)
	}

	return claims, err
}

func (s *service) RevokeToken(ctx context.Context, userID, jti string) error {
	var err error

	s.metrics.ObserveRevocation(func() {
		err = s.tokens.Revoke(ctx, jti, userID)
		// TODO: revoke refresh tokens too
	})

	if err != nil {
		s.metrics.IncrementTokensRevoked(StatusError)
	} else {
		s.metrics.IncrementTokensRevoked(StatusSuccess)
	}

	return err
}

func (s *service) newClaims(userID string, scopes []string, expiresAt time.Time) *Claims {
	now := time.Now()
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        s.idFactory(),
			Issuer:    s.config.Issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		UserID: userID,
		Scopes: scopes,
	}
	return claims
}

func (s *service) newRefreshClaims(
	userID string,
	refreshScope string,
	accessScopes []string,
	expiresAt time.Time,
) *RefreshClaims {
	claims := s.newClaims(userID, []string{refreshScope}, expiresAt)

	return &RefreshClaims{
		Claims:         *claims,
		OriginalScopes: accessScopes,
	}
}

func (s *service) sign(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.config.Secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return signedToken, nil
}
