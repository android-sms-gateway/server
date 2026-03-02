package jwt

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Service interface {
	GenerateTokenPair(
		ctx context.Context,
		userID string,
		scopes []string,
		refreshScope string,
		accessTTL time.Duration,
	) (*TokenPairInfo, error)
	RefreshTokenPair(ctx context.Context, refreshToken string) (*TokenPairInfo, error)
	ParseToken(ctx context.Context, token string) (*Claims, error)
	RevokeToken(ctx context.Context, userID, jti string) error
}

type Claims struct {
	jwt.RegisteredClaims

	UserID string   `json:"user_id"`
	Scopes []string `json:"scopes"`
}

type RefreshClaims struct {
	Claims

	OriginalScopes []string `json:"orginal_scopes"`
}

type TokenInfo struct {
	ID        string
	Token     string
	ExpiresAt time.Time
}

type TokenPairInfo struct {
	Access  TokenInfo
	Refresh TokenInfo
}
