package push

import "time"

const (
	maxRetries       = 3
	blacklistTimeout = 15 * time.Minute
)

type RetryOutcome string

const (
	RetryOutcomeRetried     RetryOutcome = "retried"
	RetryOutcomeMaxAttempts RetryOutcome = "max_attempts"
)

type BlacklistOperation string

const (
	BlacklistOperationAdded   BlacklistOperation = "added"
	BlacklistOperationSkipped BlacklistOperation = "skipped"
)
