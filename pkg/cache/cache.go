// Package cache provides a flexible caching abstraction with multiple backend implementations.
//
// The cache package offers a unified interface for caching operations with support for:
//   - In-memory caching (memoryCache)
//   - Redis-based caching (redisCache)
//   - Type-safe caching through generics (Typed[T])
//
// Features:
//   - Time-to-live (TTL) support for cache entries
//   - Atomic operations for consistency
//   - Concurrent-safe implementations
//   - Configurable expiration policies
//   - Support for custom serialization through the Item interface
//
// Basic Usage:
//
//	// Create an in-memory cache with 1 hour default TTL
//	cache := cache.NewMemory(time.Hour)
//
//	// Set a value
//	err := cache.Set(ctx, "key", []byte("value"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Get a value
//	value, err := cache.Get(ctx, "key")
//	if err != nil {
//	    if errors.Is(err, cache.ErrKeyNotFound) {
//	        // Handle missing key
//	    } else if errors.Is(err, cache.ErrKeyExpired) {
//	        // Handle expired key
//	    } else {
//	        log.Fatal(err)
//	    }
//	}
//
//	// Set with custom TTL
//	err = cache.Set(ctx, "key", []byte("value"), cache.WithTTL(30*time.Minute))
//
//	// Set only if key doesn't exist
//	err = cache.SetOrFail(ctx, "key", []byte("value"))
//	if errors.Is(err, cache.ErrKeyExists) {
//	    // Key already exists
//	}
//
//	// Get and delete in one operation
//	value, err = cache.GetAndDelete(ctx, "key")
//
//	// Remove expired entries
//	err = cache.Cleanup(ctx)
//
//	// Get all items and clear the cache
//	items, err := cache.Drain(ctx)
//
//	// Close the cache when done
//	err = cache.Close()
//
// Using Typed Cache:
//
//	// Define a type that implements the Item interface
//	type MyData struct {
//	    Field1 string
//	    Field2 int
//	}
//
//	func (d *MyData) Marshal() ([]byte, error) {
//	    return json.Marshal(d)
//	}
//
//	func (d *MyData) Unmarshal(data []byte) error {
//	    return json.Unmarshal(data, d)
//	}
//
//	// Create a typed cache
//	storage := cache.NewMemory(time.Hour)
//	typedCache := cache.NewTyped[*MyData](storage)
//
//	// Set typed value
//	data := &MyData{Field1: "test", Field2: 42}
//	err := typedCache.Set(ctx, "key", data)
//
//	// Get typed value
//	retrieved, err := typedCache.Get(ctx, "key")
//
// Using Redis Cache:
//
//	// Create a Redis cache
//	config := cache.RedisConfig{
//	    URL:    "redis://localhost:6379",
//	    Prefix: "myapp:",
//	    TTL:    time.Hour,
//	}
//
//	redisCache, err := cache.NewRedis(config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer redisCache.Close()
//
//	// Use the same interface as memory cache
//	err = redisCache.Set(ctx, "key", []byte("value"))
//	value, err := redisCache.Get(ctx, "key")
package cache

import "context"

// Cache defines the interface for cache implementations.
//
// All cache operations are context-aware and support cancellation and timeouts.
// Implementations must be safe for concurrent use by multiple goroutines.
type Cache interface {
	// Set stores the value for the given key in the cache, overwriting any existing value.
	//
	// The value will be stored with the default TTL configured for the cache implementation,
	// unless overridden by options. If the key already exists, its value and TTL will be updated.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//   - key: The key to store the value under
	//   - value: The value to store as a byte slice
	//   - opts: Optional configuration for this specific item (e.g., custom TTL)
	//
	// Returns:
	//   - error: nil on success, otherwise an error describing the failure
	//
	// Example:
	//	// Set with default TTL
	//	err := cache.Set(ctx, "user:123", []byte("user data"))
	//
	//	// Set with custom TTL
	//	err := cache.Set(ctx, "session:abc", []byte("session data"), cache.WithTTL(30*time.Minute))
	//
	//	// Set with specific expiration time
	//	expiration := time.Now().Add(2 * time.Hour)
	//	err := cache.Set(ctx, "temp:xyz", []byte("temp data"), cache.WithValidUntil(expiration))
	Set(ctx context.Context, key string, value []byte, opts ...Option) error

	// SetOrFail stores the value for the given key only if the key does not already exist.
	//
	// This is an atomic operation that prevents race conditions when multiple goroutines
	// might try to set the same key simultaneously. If the key exists but has expired,
	// it will be overwritten.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//   - key: The key to store the value under
	//   - value: The value to store as a byte slice
	//   - opts: Optional configuration for this specific item (e.g., custom TTL)
	//
	// Returns:
	//   - error: nil on success, ErrKeyExists if the key already exists and is not expired,
	//            otherwise an error describing the failure
	//
	// Example:
	//	// Try to set a value only if key doesn't exist
	//	err := cache.SetOrFail(ctx, "lock:resource", []byte("locked"))
	//	if errors.Is(err, cache.ErrKeyExists) {
	//	    // Key already exists, handle conflict
	//	}
	SetOrFail(ctx context.Context, key string, value []byte, opts ...Option) error

	// Get retrieves the value for the given key from the cache.
	//
	// The behavior depends on the key's existence and expiration state:
	//   - If the key exists and has not expired, returns the value and nil error
	//   - If the key does not exist, returns nil and ErrKeyNotFound
	//   - If the key exists but has expired, returns nil and ErrKeyExpired
	//
	// GetOptions can be used to modify the behavior, such as updating TTL,
	// deleting the key after retrieval, or setting a new expiration time.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//   - key: The key to retrieve
	//   - opts: Optional operations to perform during retrieval (e.g., AndDelete, AndSetTTL)
	//
	// Returns:
	//   - []byte: The cached value if found and not expired
	//   - error: nil on success, ErrKeyNotFound if key doesn't exist,
	//            ErrKeyExpired if key exists but has expired, otherwise an error
	//
	// Example:
	//	// Simple get
	//	value, err := cache.Get(ctx, "user:123")
	//
	//	// Get and extend TTL by 30 minutes
	//	value, err := cache.Get(ctx, "session:abc", cache.AndUpdateTTL(30*time.Minute))
	//
	//	// Get and delete atomically
	//	value, err := cache.Get(ctx, "temp:xyz", cache.AndDelete())
	Get(ctx context.Context, key string, opts ...GetOption) ([]byte, error)

	// GetAndDelete retrieves the value for the given key and atomically deletes it from the cache.
	//
	// This is equivalent to calling Get with the AndDelete option, but provides a more
	// convenient API for the common pattern of reading and removing a value in one operation.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//   - key: The key to retrieve and delete
	//
	// Returns:
	//   - []byte: The cached value if found and not expired
	//   - error: nil on success, ErrKeyNotFound if key doesn't exist,
	//            ErrKeyExpired if key exists but has expired, otherwise an error
	//
	// Example:
	//	// Atomically get and remove a value
	//	value, err := cache.GetAndDelete(ctx, "queue:item:123")
	GetAndDelete(ctx context.Context, key string) ([]byte, error)

	// Delete removes the item associated with the given key from the cache.
	//
	// If the key does not exist, this operation performs no action and returns nil.
	// The operation is safe for concurrent use by multiple goroutines.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//   - key: The key to remove from the cache
	//
	// Returns:
	//   - error: nil on success, otherwise an error describing the failure
	//
	// Example:
	//	// Remove a specific key
	//	err := cache.Delete(ctx, "user:123")
	Delete(ctx context.Context, key string) error

	// Cleanup removes all expired items from the cache.
	//
	// This operation scans the entire cache and removes any items that have expired.
	// The operation is safe for concurrent use by multiple goroutines.
	// Note that some cache implementations (like Redis) handle expiration automatically
	// and may not require explicit cleanup.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//
	// Returns:
	//   - error: nil on success, otherwise an error describing the failure
	//
	// Example:
	//	// Periodically clean up expired items
	//	err := cache.Cleanup(ctx)
	Cleanup(ctx context.Context) error

	// Drain returns a map of all non-expired items in the cache and clears the cache.
	//
	// The returned map is a snapshot of the cache at the time of the call.
	// After this operation, the cache will be empty. This is useful for cache migration,
	// backup, or when shutting down an application.
	// The operation is safe for concurrent use by multiple goroutines.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//
	// Returns:
	//   - map[string][]byte: A map containing all non-expired key-value pairs
	//   - error: nil on success, otherwise an error describing the failure
	//
	// Example:
	//	// Get all items and clear the cache
	//	items, err := cache.Drain(ctx)
	//	for key, value := range items {
	//	    log.Printf("Drained: %s = %s", key, string(value))
	//	}
	Drain(ctx context.Context) (map[string][]byte, error)

	// Close releases any resources held by the cache.
	//
	// This should be called when the cache is no longer needed. For some implementations
	// (like Redis), this may close network connections. The operation is safe for
	// concurrent use by multiple goroutines.
	//
	// Returns:
	//   - error: nil on success, otherwise an error describing the failure
	//
	// Example:
	//	// Properly close the cache when done
	//	defer cache.Close()
	Close() error
}
