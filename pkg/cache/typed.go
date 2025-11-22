package cache

import (
	"context"
	"fmt"
	"reflect"
)

// Item defines the interface that types must implement to be used with Typed cache.
//
// Types that implement this interface can be automatically serialized and
// deserialized when stored in or retrieved from the cache. This allows
// for type-safe caching of complex data structures.
//
// Example implementation:
//
//	type User struct {
//	    ID   string
//	    Name string
//	}
//
//	func (u *User) Marshal() ([]byte, error) {
//	    return json.Marshal(u)
//	}
//
//	func (u *User) Unmarshal(data []byte) error {
//	    return json.Unmarshal(data, u)
//	}
type Item interface {
	// Marshal converts the item to a byte slice for storage in the cache.
	//
	// This method is called when the item is stored in the cache.
	// Common implementations include JSON, protobuf, or other serialization formats.
	//
	// Returns:
	//   - []byte: The serialized representation of the item
	//   - error: An error if serialization fails
	Marshal() ([]byte, error)

	// Unmarshal populates the item from a byte slice retrieved from the cache.
	//
	// This method is called when the item is retrieved from the cache.
	// It should restore the item's state from the serialized data.
	//
	// Parameters:
	//   - data: The serialized representation of the item
	//
	// Returns:
	//   - error: An error if deserialization fails
	Unmarshal(data []byte) error
}

// Typed provides a type-safe wrapper around a Cache implementation.
//
// This generic wrapper allows caching of specific types that implement the Item
// interface, providing compile-time type safety and eliminating the need for
// manual type assertions when working with cached values.
//
// The Typed wrapper handles serialization and deserialization automatically,
// making it easy to cache complex data structures while maintaining type safety.
//
// Example usage:
//
//	// Define a type that implements Item
//	type User struct {
//	    ID   string
//	    Name string
//	}
//
//	func (u *User) Marshal() ([]byte, error) {
//	    return json.Marshal(u)
//	}
//
//	func (u *User) Unmarshal(data []byte) error {
//	    return json.Unmarshal(data, u)
//	}
//
//	// Create a typed cache
//	storage := cache.NewMemory(time.Hour)
//	userCache := cache.NewTyped[*User](storage)
//
//	// Set a typed value
//	user := &User{ID: "123", Name: "Alice"}
//	err := userCache.Set(ctx, "user:123", user)
//
//	// Get a typed value
//	retrieved, err := userCache.Get(ctx, "user:123")
//	// retrieved is of type *User, no type assertion needed
type Typed[T Item] struct {
	storage Cache
}

// NewTyped creates a new typed cache wrapper around the provided storage.
//
// The typed cache uses the underlying storage for all operations but adds
// automatic serialization and deserialization of values that implement the Item interface.
//
// Parameters:
//   - storage: The underlying cache implementation to wrap
//
// Returns:
//   - *Typed[T]: A new typed cache wrapper
//
// Example:
//
//	// Create a typed cache with in-memory storage
//	storage := cache.NewMemory(time.Hour)
//	userCache := cache.NewTyped[*User](storage)
//
//	// Create a typed cache with Redis storage
//	config := cache.RedisConfig{URL: "redis://localhost:6379"}
//	redisCache, err := cache.NewRedis(config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer redisCache.Close()
//	userCache := cache.NewTyped[*User](redisCache)
func NewTyped[T Item](storage Cache) *Typed[T] {
	return &Typed[T]{
		storage: storage,
	}
}

// Set stores the typed value for the given key in the cache, overwriting any existing value.
//
// The value will be automatically marshaled to bytes before storage using its
// Marshal method. The value will be stored with the default TTL configured for
// the cache implementation, unless overridden by options.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - key: The key to store the value under
//   - value: The typed value to store
//   - opts: Optional configuration for this specific item (e.g., custom TTL)
//
// Returns:
//   - error: nil on success, or an error if marshaling fails or the cache operation fails
//
// Example:
//
//	// Set with default TTL
//	user := &User{ID: "123", Name: "Alice"}
//	err := userCache.Set(ctx, "user:123", user)
//
//	// Set with custom TTL
//	err := userCache.Set(ctx, "session:abc", user, cache.WithTTL(30*time.Minute))
func (c *Typed[T]) Set(ctx context.Context, key string, value T, opts ...Option) error {
	data, err := value.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if setErr := c.storage.Set(ctx, key, data, opts...); setErr != nil {
		return fmt.Errorf("failed to set value in cache: %w", setErr)
	}

	return nil
}

// SetOrFail stores the typed value for the given key only if the key does not already exist.
//
// This is an atomic operation that prevents race conditions when multiple goroutines
// might try to set the same key simultaneously. The value will be automatically
// marshaled to bytes before storage using its Marshal method.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - key: The key to store the value under
//   - value: The typed value to store
//   - opts: Optional configuration for this specific item (e.g., custom TTL)
//
// Returns:
//   - error: nil on success, ErrKeyExists if the key already exists and is not expired,
//     otherwise an error if marshaling fails or the cache operation fails
//
// Example:
//
//	// Try to set a value only if key doesn't exist
//	user := &User{ID: "123", Name: "Alice"}
//	err := userCache.SetOrFail(ctx, "user:123", user)
//	if errors.Is(err, cache.ErrKeyExists) {
//	    // Key already exists, handle conflict
//	}
func (c *Typed[T]) SetOrFail(ctx context.Context, key string, value T, opts ...Option) error {
	data, err := value.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if setErr := c.storage.SetOrFail(ctx, key, data, opts...); setErr != nil {
		return fmt.Errorf("failed to set value in cache: %w", setErr)
	}

	return nil
}

// Get retrieves the typed value for the given key from the cache.
//
// The behavior depends on the key's existence and expiration state:
//   - If the key exists and has not expired, returns the typed value and nil error
//   - If the key does not exist, returns a zero value and ErrKeyNotFound
//   - If the key exists but has expired, returns a zero value and ErrKeyExpired
//
// The retrieved bytes are automatically unmarshaled to the correct type using
// the Unmarshal method. GetOptions can be used to modify the behavior, such as
// updating TTL, deleting the key after retrieval, or setting a new expiration time.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - key: The key to retrieve
//   - opts: Optional operations to perform during retrieval (e.g., AndDelete, AndSetTTL)
//
// Returns:
//   - T: The cached typed value if found and not expired, otherwise a zero value
//   - error: nil on success, ErrKeyNotFound if key doesn't exist,
//     ErrKeyExpired if key exists but has expired, otherwise an error
//
// Example:
//
//	// Simple get
//	user, err := userCache.Get(ctx, "user:123")
//
//	// Get and extend TTL by 30 minutes
//	user, err := userCache.Get(ctx, "session:abc", cache.AndUpdateTTL(30*time.Minute))
//
//	// Get and delete atomically
//	user, err := userCache.Get(ctx, "temp:xyz", cache.AndDelete())
func (c *Typed[T]) Get(ctx context.Context, key string, opts ...GetOption) (T, error) {
	data, err := c.storage.Get(ctx, key, opts...)
	var zero T
	if err != nil {
		return zero, fmt.Errorf("failed to get value from cache: %w", err)
	}

	value, err := newItem[T]()
	if err != nil {
		return zero, err
	}
	if unmarshalErr := value.Unmarshal(data); unmarshalErr != nil {
		return zero, fmt.Errorf("failed to unmarshal value from cache: %w", unmarshalErr)
	}

	return value, nil
}

// GetAndDelete retrieves the typed value for the given key and atomically deletes it from the cache.
//
// This is equivalent to calling Get with the AndDelete option, but provides a more
// convenient API for the common pattern of reading and removing a value in one operation.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - key: The key to retrieve and delete
//
// Returns:
//   - T: The cached typed value if found and not expired, otherwise a zero value
//   - error: nil on success, ErrKeyNotFound if key doesn't exist,
//     ErrKeyExpired if key exists but has expired, otherwise an error
//
// Example:
//
//	// Atomically get and remove a value
//	user, err := userCache.GetAndDelete(ctx, "queue:item:123")
func (c *Typed[T]) GetAndDelete(ctx context.Context, key string) (T, error) {
	return c.Get(ctx, key, AndDelete())
}

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
//   - error: nil on success, or an error if the cache operation fails
//
// Example:
//
//	// Remove a specific key
//	err := userCache.Delete(ctx, "user:123")
func (c *Typed[T]) Delete(ctx context.Context, key string) error {
	if err := c.storage.Delete(ctx, key); err != nil {
		return fmt.Errorf("failed to delete value from cache: %w", err)
	}

	return nil
}

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
//   - error: nil on success, or an error if the cache operation fails
//
// Example:
//
//	// Periodically clean up expired items
//	err := userCache.Cleanup(ctx)
func (c *Typed[T]) Cleanup(ctx context.Context) error {
	if err := c.storage.Cleanup(ctx); err != nil {
		return fmt.Errorf("failed to cleanup cache: %w", err)
	}

	return nil
}

// Drain returns a map of all non-expired typed items in the cache and clears the cache.
//
// The returned map is a snapshot of the cache at the time of the call.
// After this operation, the cache will be empty. This is useful for cache migration,
// backup, or when shutting down an application. The operation is safe for
// concurrent use by multiple goroutines.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//
// Returns:
//   - map[string]T: A map containing all non-expired key-typed value pairs
//   - error: nil on success, or an error if the cache operation fails
//
// Example:
//
//	// Get all items and clear the cache
//	users, err := userCache.Drain(ctx)
//	for key, user := range users {
//	    log.Printf("Drained: %s = %+v", key, user)
//	}
func (c *Typed[T]) Drain(ctx context.Context) (map[string]T, error) {
	data, err := c.storage.Drain(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to drain cache: %w", err)
	}

	items := make(map[string]T, len(data))
	for key, raw := range data {
		item, newErr := newItem[T]()
		if newErr != nil {
			return nil, newErr
		}
		if unmarshalErr := item.Unmarshal(raw); unmarshalErr != nil {
			return nil, fmt.Errorf("failed to unmarshal value from cache: %w", unmarshalErr)
		}

		items[key] = item
	}

	return items, nil
}

// Close releases any resources held by the cache.
//
// This should be called when the cache is no longer needed. For some implementations
// (like Redis), this may close network connections. The operation is safe for
// concurrent use by multiple goroutines.
//
// Returns:
//   - error: nil on success, or an error if the cache operation fails
//
// Example:
//
//	// Properly close the cache when done
//	defer userCache.Close()
func (c *Typed[T]) Close() error {
	if err := c.storage.Close(); err != nil {
		return fmt.Errorf("failed to close cache: %w", err)
	}

	return nil
}

// newItem creates a new instance of type T using reflection.
//
// This is a helper function that creates a new instance of the generic type T,
// which must be a pointer type that implements the Item interface. It uses
// reflection to instantiate the type and verifies that it implements the interface.
//
// Returns:
//   - T: A new instance of type T
//   - error: An error if T is not a pointer type or cannot be instantiated
//
// Note: This is an internal helper function and is not intended for direct use.
func newItem[T Item]() (T, error) {
	var zero T

	t := reflect.TypeOf((*T)(nil)).Elem()
	if t.Kind() != reflect.Ptr {
		return zero, fmt.Errorf("%w: type %s must be a pointer", ErrFailedToCreateZeroValue, t.String())
	}

	v := reflect.New(t.Elem())
	item, ok := v.Interface().(T)
	if !ok {
		return zero, fmt.Errorf("%w: cannot create value of type %s", ErrFailedToCreateZeroValue, t.String())
	}

	return item, nil
}
