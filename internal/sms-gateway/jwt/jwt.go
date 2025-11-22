package jwt

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Service interface {
	GenerateToken(ctx context.Context, userID string, scopes []string, ttl time.Duration) (*TokenInfo, error)
	ParseToken(ctx context.Context, token string) (*Claims, error)
	RevokeToken(ctx context.Context, userID, jti string) error
}

type Claims struct {
	jwt.RegisteredClaims

	UserID string   `json:"user_id"`
	Scopes []string `json:"scopes"`
}

type TokenInfo struct {
	ID          string
	AccessToken string
	ExpiresAt   time.Time
}
