package executor

import (
	"context"
	"time"
)

type PeriodicTask interface {
	Name() string
	Interval() time.Duration
	Run(ctx context.Context) error
}
