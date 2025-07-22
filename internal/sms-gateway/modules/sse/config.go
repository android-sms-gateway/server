package sse

import "time"

const defaultKeepAlivePeriod = 15 * time.Second

type Config struct {
	keepAlivePeriod time.Duration
}

func NewConfig() Config {
	return Config{
		keepAlivePeriod: defaultKeepAlivePeriod,
	}
}

func (c *Config) KeepAlivePeriod() time.Duration {
	return c.keepAlivePeriod
}

func (c *Config) SetKeepAlivePeriod(d time.Duration) *Config {
	if d <= 0 {
		d = defaultKeepAlivePeriod
	}

	c.keepAlivePeriod = d

	return c
}
