package messages

import "time"

type Config struct {
	ProcessedLifetime time.Duration
	CacheTTL          time.Duration
}
