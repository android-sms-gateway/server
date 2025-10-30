package locker

import (
	"context"
	"errors"
)

var ErrLockNotAcquired = errors.New("lock not acquired")

type Locker interface {
	AcquireLock(ctx context.Context, key string) error
	ReleaseLock(ctx context.Context, key string) error
}
