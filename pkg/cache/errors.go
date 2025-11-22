package cache

import "errors"

var (
	// ErrInvalidConfig indicates an invalid configuration.
	ErrInvalidConfig = errors.New("invalid config")
	// ErrKeyNotFound indicates no value exists for the given key.
	//
	// This error is returned by Get operations when the requested key has never been
	// set in the cache or has been explicitly deleted. It is also returned by
	// GetAndDelete when the key does not exist.
	//
	// Example:
	//	_, err := cache.Get(ctx, "nonexistent-key")
	//	if errors.Is(err, cache.ErrKeyNotFound) {
	//	    // Handle missing key
	//	}
	ErrKeyNotFound = errors.New("key not found")

	// ErrKeyExpired indicates a value exists but has expired.
	//
	// This error is returned by Get operations when the requested key exists in the
	// cache but its time-to-live (TTL) has elapsed. Expired items may still exist
	// in the cache until they are explicitly removed by a Cleanup operation or
	// automatically by the cache implementation.
	//
	// Example:
	//	// Set a value with 1 second TTL
	//	cache.Set(ctx, "temp-key", []byte("data"), cache.WithTTL(time.Second))
	//	time.Sleep(2 * time.Second)
	//	_, err := cache.Get(ctx, "temp-key")
	//	if errors.Is(err, cache.ErrKeyExpired) {
	//	    // Handle expired key
	//	}
	ErrKeyExpired = errors.New("key expired")

	// ErrKeyExists indicates a conflicting set when the key already exists.
	//
	// This error is returned by SetOrFail operations when attempting to set a value
	// for a key that already exists in the cache and has not expired. This is useful
	// for implementing atomic "create if not exists" operations and preventing
	// race conditions in concurrent scenarios.
	//
	// Example:
	//	// Try to set a value only if key doesn't exist
	//	err := cache.SetOrFail(ctx, "lock-key", []byte("locked"))
	//	if errors.Is(err, cache.ErrKeyExists) {
	//	    // Key already exists, handle conflict
	//	}
	ErrKeyExists = errors.New("key already exists")

	ErrFailedToCreateZeroValue = errors.New("failed to create zero item")
)
