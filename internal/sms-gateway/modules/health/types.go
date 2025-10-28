package health

import (
	"context"
)

type Status string
type statusLevel int

const (
	StatusPass Status = "pass"
	StatusWarn Status = "warn"
	StatusFail Status = "fail"

	levelPass statusLevel = 0
	levelWarn statusLevel = 1
	levelFail statusLevel = 2
)

var statusLevels = map[statusLevel]Status{
	levelPass: StatusPass,
	levelWarn: StatusWarn,
	levelFail: StatusFail,
}

// Health status of the application.
type CheckResult struct {
	// A map of check names to their respective details.
	Checks Checks
}

// Overall status of the application.
// It can be one of the following values: "pass", "warn", or "fail".
func (c CheckResult) Status() Status {
	// Determine overall status
	level := levelPass
	for _, detail := range c.Checks {
		switch detail.Status {
		case StatusPass:
		case StatusFail:
			level = max(level, levelFail)
		case StatusWarn:
			level = max(level, levelWarn)
		}
	}

	return statusLevels[level]
}

// Details of a health check.
type CheckDetail struct {
	// A human-readable description of the check.
	Description string
	// Unit of measurement for the observed value.
	ObservedUnit string
	// Observed value of the check.
	ObservedValue int
	// Status of the check.
	// It can be one of the following values: "pass", "warn", or "fail".
	Status Status
}

// Map of check names to their respective details.
type Checks map[string]CheckDetail

type HealthProvider interface {
	Name() string

	StartedProbe(ctx context.Context) (Checks, error)
	ReadyProbe(ctx context.Context) (Checks, error)
	LiveProbe(ctx context.Context) (Checks, error)
}
