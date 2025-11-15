package health

import (
	"context"
	"runtime"
)

type health struct {
}

func newHealth() *health {
	return &health{}
}

// Name implements HealthProvider.
func (h *health) Name() string {
	return "system"
}

// LiveProbe implements HealthProvider.
func (h *health) LiveProbe(_ context.Context) (Checks, error) {
	const oneMiB uint64 = 1 << 20
	const memoryThreshold uint64 = 128 * oneMiB
	const goroutineThreshold = 100

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Basic runtime health checks
	goroutineCheck := CheckDetail{
		Description:   "Number of goroutines",
		ObservedValue: runtime.NumGoroutine(),
		ObservedUnit:  "goroutines",
		Status:        StatusPass,
	}

	memoryCheck := CheckDetail{
		Description:   "Memory usage",
		ObservedValue: int(m.Alloc / oneMiB), //nolint:gosec // not a security issue
		ObservedUnit:  "MiB",
		Status:        StatusPass,
	}

	// Check for potential memory issues
	if m.Alloc > memoryThreshold {
		memoryCheck.Status = StatusWarn
	}

	// Check for excessive goroutines
	if goroutineCheck.ObservedValue > goroutineThreshold {
		goroutineCheck.Status = StatusWarn
	}

	return Checks{"goroutines": goroutineCheck, "memory": memoryCheck}, nil
}

// ReadyProbe implements HealthProvider.
func (h *health) ReadyProbe(_ context.Context) (Checks, error) {
	return nil, nil //nolint:nilnil // empty result
}

// StartedProbe implements HealthProvider.
func (h *health) StartedProbe(_ context.Context) (Checks, error) {
	return nil, nil //nolint:nilnil // empty result
}

var _ Provider = (*health)(nil)
