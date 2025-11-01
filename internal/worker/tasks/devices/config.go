package devices

import "time"

type Config struct {
	Cleanup CleanupConfig
}

type CleanupConfig struct {
	Interval time.Duration
	MaxAge   time.Duration
}
