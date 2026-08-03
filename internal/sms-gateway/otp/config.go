package otp

import (
	"fmt"
	"time"
)

const minLength = 6

// Config configures the OTP service.
type Config struct {
	Enabled bool          `yaml:"enabled" envconfig:"OTP__ENABLED"`
	Length  int           `yaml:"length"  envconfig:"OTP__LENGTH"`
	TTL     time.Duration `yaml:"ttl"     envconfig:"OTP__TTL"`
	Retries int           `yaml:"retries" envconfig:"OTP__RETRIES"`
}

func (c Config) Validate() error {
	if c.Length < minLength {
		return fmt.Errorf("%w: length must be at least %d", ErrInvalidConfig, minLength)
	}

	if c.TTL <= 0 {
		return fmt.Errorf("%w: TTL must be greater than 0", ErrInvalidConfig)
	}

	if c.Retries <= 0 {
		return fmt.Errorf("%w: retries must be greater than 0", ErrInvalidConfig)
	}

	return nil
}
