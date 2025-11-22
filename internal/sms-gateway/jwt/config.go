package jwt

import (
	"fmt"
	"time"
)

const (
	minSecretLength = 32
)

type Config struct {
	Secret string
	TTL    time.Duration
	Issuer string
}

func (c Config) Validate() error {
	if c.Secret == "" {
		return fmt.Errorf("%w: secret is required", ErrInvalidConfig)
	}

	if len(c.Secret) < minSecretLength {
		return fmt.Errorf("%w: secret must be at least %d bytes", ErrInvalidConfig, minSecretLength)
	}

	if c.TTL <= 0 {
		return fmt.Errorf("%w: ttl must be positive", ErrInvalidConfig)
	}

	return nil
}
