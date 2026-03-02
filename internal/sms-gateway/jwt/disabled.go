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

// GenerateTokenPair implements Service.
func (d *disabled) GenerateTokenPair(
	_ context.Context,
	_ string,
	_ []string,
	_ string,
	_ time.Duration,
) (*TokenPairInfo, error) {
	return nil, ErrDisabled
}

// RefreshTokenPair implements Service.
func (d *disabled) RefreshTokenPair(_ context.Context, _ string) (*TokenPairInfo, error) {
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
