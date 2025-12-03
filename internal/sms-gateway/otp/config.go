package otp

import (
	"fmt"
	"time"
)

// Config configures the OTP service.
type Config struct {
	TTL     time.Duration `yaml:"ttl"     envconfig:"OTP__TTL"`
	Retries int           `yaml:"retries" envconfig:"OTP__RETRIES"`
}

func (c Config) Validate() error {
	if c.TTL <= 0 {
		return fmt.Errorf("%w: TTL must be greater than 0", ErrInvalidConfig)
	}

	if c.Retries <= 0 {
		return fmt.Errorf("%w: retries must be greater than 0", ErrInvalidConfig)
	}

	return nil
}
