package sse

import "time"

type configOption func(*Config)

type Config struct {
	keepAlivePeriod time.Duration
}

const defaultKeepAlivePeriod = 15 * time.Second

var defaultConfig = Config{
	keepAlivePeriod: defaultKeepAlivePeriod,
}

func NewConfig(opts ...configOption) Config {
	c := defaultConfig

	for _, opt := range opts {
		opt(&c)
	}

	return c
}

func (c *Config) KeepAlivePeriod() time.Duration {
	return c.keepAlivePeriod
}

func WithKeepAlivePeriod(d time.Duration) configOption {
	if d < 0 {
		d = defaultKeepAlivePeriod
	}

	return func(c *Config) {
		c.keepAlivePeriod = d
	}
}
