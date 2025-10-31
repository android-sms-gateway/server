package locker

import (
	"context"
	"errors"
)

// ErrLockNotAcquired is returned when a lock cannot be acquired within the configured timeout.
var ErrLockNotAcquired = errors.New("lock not acquired")

type Locker interface {
	// AcquireLock attempts to acquire a lock for the given key.
	// Returns ErrLockNotAcquired if the lock cannot be acquired within the configured timeout.
	AcquireLock(ctx context.Context, key string) error
	// ReleaseLock releases a previously acquired lock for the given key.
	// Returns an error if the lock was not held by this instance.
	ReleaseLock(ctx context.Context, key string) error

	// Close releases any held locks.
	Close() error
}
