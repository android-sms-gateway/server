package sse

import "time"

type Option func(*Config)

type Config struct {
	keepAlivePeriod time.Duration
}

const defaultKeepAlivePeriod = 15 * time.Second

func DefaultConfig() Config {
	return Config{
		keepAlivePeriod: defaultKeepAlivePeriod,
	}
}

func NewConfig(opts ...Option) Config {
	c := DefaultConfig()

	for _, opt := range opts {
		opt(&c)
	}

	return c
}

func (c *Config) KeepAlivePeriod() time.Duration {
	return c.keepAlivePeriod
}

func WithKeepAlivePeriod(d time.Duration) Option {
	if d < 0 {
		d = defaultKeepAlivePeriod
	}

	return func(c *Config) {
		c.keepAlivePeriod = d
	}
}
