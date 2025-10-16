package cache

import "context"

type Cache interface {
	// Set sets the value for the given key in the cache.
	Set(ctx context.Context, key string, value []byte, opts ...Option) error

	// SetOrFail is like Set, but returns ErrKeyExists if the key already exists.
	SetOrFail(ctx context.Context, key string, value []byte, opts ...Option) error

	// Get gets the value for the given key from the cache.
	//
	// If the key is not found, it returns ErrKeyNotFound.
	// If the key has expired, it returns ErrKeyExpired.
	// Otherwise, it returns the value and nil.
	Get(ctx context.Context, key string, opts ...GetOption) ([]byte, error)

	// GetAndDelete is like Get, but also deletes the key from the cache.
	GetAndDelete(ctx context.Context, key string) ([]byte, error)

	// Delete removes the item associated with the given key from the cache.
	// If the key does not exist, it performs no action and returns nil.
	// The operation is safe for concurrent use.
	Delete(ctx context.Context, key string) error

	// Cleanup removes all expired items from the cache.
	// The operation is safe for concurrent use.
	Cleanup(ctx context.Context) error

	// Drain returns a map of all the non-expired items in the cache.
	// The returned map is a snapshot of the cache at the time of the call.
	// The cache is cleared after the call.
	// The operation is safe for concurrent use.
	Drain(ctx context.Context) (map[string][]byte, error)

	// Close closes the cache.
	// The operation is safe for concurrent use.
	Close() error
}
