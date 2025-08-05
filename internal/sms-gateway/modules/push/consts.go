package push

import "time"

const (
	maxRetries       = 3
	blacklistTimeout = 15 * time.Minute
)
