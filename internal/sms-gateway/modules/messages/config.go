package messages

import "time"

type Config struct {
	HashingInterval time.Duration
	CacheTTL        time.Duration
}
