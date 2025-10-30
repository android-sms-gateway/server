package messages

import "time"

type Config struct {
	Hashing HashingConfig
	Cleanup CleanupConfig
}

type HashingConfig struct {
	Interval time.Duration
}

type CleanupConfig struct {
	Interval time.Duration
	MaxAge   time.Duration
}
