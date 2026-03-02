package jwt

import (
	"fmt"
	"time"
)

const (
	minSecretLength = 32
)

type Config struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	Issuer     string
}

func (c Config) Validate() error {
	if c.Secret == "" {
		return fmt.Errorf("%w: secret is required", ErrInvalidConfig)
	}

	if len(c.Secret) < minSecretLength {
		return fmt.Errorf("%w: secret must be at least %d bytes", ErrInvalidConfig, minSecretLength)
	}

	if c.AccessTTL <= 0 {
		return fmt.Errorf("%w: access ttl must be positive", ErrInvalidConfig)
	}

	if c.RefreshTTL <= 0 {
		return fmt.Errorf("%w: refresh ttl must be positive", ErrInvalidConfig)
	}

	if c.RefreshTTL <= c.AccessTTL {
		return fmt.Errorf("%w: refresh ttl must be greater than access ttl", ErrInvalidConfig)
	}

	return nil
}
