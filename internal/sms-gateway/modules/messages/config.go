package messages

import "time"

type Config struct {
	HashingInterval time.Duration
	CacheTTL        time.Duration

	Queue QueueConfig
}

type QueueConfig struct {
	MaxPending    int64
	MaxPendingAge time.Duration
	MaxFailed     int
	MaxFailedAge  time.Duration

	StatsRefreshInterval time.Duration
	StatsCacheTTL        time.Duration
}

func (c QueueConfig) IsEmpty() bool {
	return c.MaxPending <= 0 && c.MaxPendingAge <= 0 && c.MaxFailed <= 0
}
