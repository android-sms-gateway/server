package health

import (
	"context"
	"runtime"
)

type health struct {
}

func NewHealth() *health {
	return &health{}
}

// Name implements HealthProvider.
func (h *health) Name() string {
	return "system"
}

// LiveProbe implements HealthProvider.
func (h *health) LiveProbe(ctx context.Context) (Checks, error) {
	const oneGiB uint64 = 1 << 30

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Basic runtime health checks
	goroutineCheck := CheckDetail{
		Description:   "Number of goroutines",
		ObservedValue: int(runtime.NumGoroutine()),
		ObservedUnit:  "goroutines",
		Status:        StatusPass,
	}

	memoryCheck := CheckDetail{
		Description:   "Memory usage",
		ObservedValue: int(m.Alloc / 1024 / 1024), // MiB
		ObservedUnit:  "MiB",
		Status:        StatusPass,
	}

	// Check for potential memory issues
	if m.Alloc > oneGiB { // 1GB
		memoryCheck.Status = StatusWarn
	}

	// Check for excessive goroutines
	if goroutineCheck.ObservedValue > 1000 {
		goroutineCheck.Status = StatusWarn
	}

	return Checks{"goroutines": goroutineCheck, "memory": memoryCheck}, nil
}

// ReadyProbe implements HealthProvider.
func (h *health) ReadyProbe(ctx context.Context) (Checks, error) {
	return nil, nil
}

// StartedProbe implements HealthProvider.
func (h *health) StartedProbe(ctx context.Context) (Checks, error) {
	return nil, nil
}

var _ HealthProvider = (*health)(nil)
