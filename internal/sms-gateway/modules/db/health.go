package db

import (
	"context"
	"database/sql"
	"sync/atomic"

	healthmod "github.com/android-sms-gateway/server/internal/sms-gateway/modules/health"
)

type health struct {
	db *sql.DB

	failedPings atomic.Int64
}

func newHealth(db *sql.DB) *health {
	return &health{
		db: db,
	}
}

// Name implements HealthProvider.
func (h *health) Name() string {
	return "db"
}

// LiveProbe implements HealthProvider.
func (h *health) LiveProbe(ctx context.Context) (healthmod.Checks, error) {
	return nil, nil
}

// ReadyProbe implements HealthProvider.
func (h *health) ReadyProbe(ctx context.Context) (healthmod.Checks, error) {
	pingCheck := healthmod.CheckDetail{
		Description:   "Database ping",
		ObservedUnit:  "failed pings",
		ObservedValue: 0,
		Status:        healthmod.StatusPass,
	}

	if err := h.db.PingContext(ctx); err != nil {
		h.failedPings.Add(1)

		pingCheck.Status = healthmod.StatusFail
	} else {
		h.failedPings.Store(0)
	}

	pingCheck.ObservedValue = int(h.failedPings.Load())

	return healthmod.Checks{"ping": pingCheck}, nil
}

// StartedProbe implements HealthProvider.
func (h *health) StartedProbe(ctx context.Context) (healthmod.Checks, error) {
	return nil, nil
}

var _ healthmod.HealthProvider = (*health)(nil)
