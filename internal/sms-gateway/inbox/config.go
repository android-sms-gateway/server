package inbox

import "time"

type Config struct {
	RetentionPeriod time.Duration
	CleanupInterval time.Duration
}
