package jwt

import (
	"context"
	"time"
)

type disabled struct {
}

func newDisabled() Service {
	return &disabled{}
}

// GenerateToken implements Service.
func (d *disabled) GenerateToken(_ context.Context, _ string, _ []string, _ time.Duration) (*TokenInfo, error) {
	return nil, ErrDisabled
}

// ParseToken implements Service.
func (d *disabled) ParseToken(_ context.Context, _ string) (*Claims, error) {
	return nil, ErrDisabled
}

// RevokeToken implements Service.
func (d *disabled) RevokeToken(_ context.Context, _, _ string) error {
	return ErrDisabled
}
