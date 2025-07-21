package sse

import "time"

const defaultKeepAlivePeriod = 15 * time.Second

type Config struct {
	KeepAlivePeriod time.Duration
}

func (c *Config) SetDefaults() {
	if c.KeepAlivePeriod == 0 {
		c.KeepAlivePeriod = defaultKeepAlivePeriod
	}
}
