package jwt

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jaevor/go-nanoid"
)

const jtiLength = 21

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

func (s *service) GenerateToken(ctx context.Context, userID string, scopes []string, ttl time.Duration) (*TokenInfo, error) {
	var tokenInfo *TokenInfo
	var err error

	s.metrics.ObserveIssuance(func() {
		if userID == "" {
			err = fmt.Errorf("%w: user id is required", ErrInvalidParams)
			return
		}

		if len(scopes) == 0 {
			err = fmt.Errorf("%w: scopes are required", ErrInvalidParams)
			return
		}

		if ttl < 0 {
			err = fmt.Errorf("%w: ttl must be non-negative", ErrInvalidParams)
			return
		}

		if ttl == 0 {
			ttl = s.config.TTL
		}

		now := time.Now()
		claims := &Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				ID:        s.idFactory(),
				Issuer:    s.config.Issuer,
				Subject:   userID,
				IssuedAt:  jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(min(ttl, s.config.TTL))),
			},
			UserID: userID,
			Scopes: scopes,
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, signErr := token.SignedString([]byte(s.config.Secret))
		if signErr != nil {
			err = fmt.Errorf("failed to sign token: %w", signErr)
			return
		}

		if storeErr := s.tokens.Insert(ctx, newTokenModel(claims.ID, claims.UserID, claims.ExpiresAt.Time)); storeErr != nil {
			err = fmt.Errorf("failed to insert token: %w", storeErr)
			return
		}

		tokenInfo = &TokenInfo{ID: claims.ID, AccessToken: signedToken, ExpiresAt: claims.ExpiresAt.Time}
	})

	if err != nil {
		s.metrics.IncrementTokensIssued(StatusError)
	} else {
		s.metrics.IncrementTokensIssued(StatusSuccess)
	}

	return tokenInfo, err
}

func (s *service) ParseToken(ctx context.Context, token string) (*Claims, error) {
	var claims *Claims
	var err error

	s.metrics.ObserveValidation(func() {
		parsedToken, parseErr := jwt.ParseWithClaims(
			token,
			new(Claims),
			func(t *jwt.Token) (any, error) {
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
	})

	if err != nil {
		s.metrics.IncrementTokensRevoked(StatusError)
	} else {
		s.metrics.IncrementTokensRevoked(StatusSuccess)
	}

	return err
}
