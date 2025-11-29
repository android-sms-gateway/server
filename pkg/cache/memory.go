package cache

import (
	"context"
	"sync"
	"time"
)

// MemoryCache implements an in-memory cache with TTL support.
//
// This implementation stores all data in a Go map protected by a read-write mutex,
// making it safe for concurrent access by multiple goroutines. Items are automatically
// checked for expiration on access, but expired items remain in memory until
// explicitly removed by a Cleanup operation or overwritten.
//
// The memory cache is suitable for:
//   - Single-process applications
//   - Caching small to medium amounts of data
//   - Scenarios where low latency is critical
//   - Temporary caching that doesn't need persistence
//
// For distributed caching or persistence, consider using the Redis implementation instead.
type MemoryCache struct {
	items map[string]*memoryItem
	ttl   time.Duration

	mux sync.RWMutex
}

// NewMemory creates a new in-memory cache with the specified default TTL.
//
// The TTL parameter sets the default time-to-live for items stored in the cache.
// A TTL of zero means items do not expire by default, but individual items
// can still have their own TTL set via options.
//
// Parameters:
//   - ttl: The default TTL for cache items. Zero means no expiration.
//
// Returns:
//   - *MemoryCache: A new in-memory cache instance
//
// Example:
//
//	// Create a cache with 1 hour default TTL
//	cache := cache.NewMemory(time.Hour)
//
//	// Create a cache with no expiration
//	cache := cache.NewMemory(0)
func NewMemory(ttl time.Duration) *MemoryCache {
	return &MemoryCache{
		items: make(map[string]*memoryItem),
		ttl:   ttl,

		mux: sync.RWMutex{},
	}
}

// memoryItem represents a single item in the memory cache.
type memoryItem struct {
	value      []byte
	validUntil time.Time
}

// newMemoryItem creates a new memory item with the specified value and options.
func newMemoryItem(value []byte, opts options) *memoryItem {
	item := &memoryItem{
		value:      value,
		validUntil: opts.validUntil,
	}

	return item
}

// isExpired checks if the item has expired at the given time.
//
// An item is considered expired if:
//   - The item is nil (safety check)
//   - The item has a non-zero validUntil time and the current time is after validUntil
//
// Parameters:
//   - now: The time to check expiration against
//
// Returns:
//   - bool: True if the item has expired, false otherwise
func (i *memoryItem) isExpired(now time.Time) bool {
	if i == nil {
		return true
	}

	return !i.validUntil.IsZero() && now.After(i.validUntil)
}

// Cleanup removes all expired items from the memory cache.
//
// This method scans through all items in the cache and removes any that have expired.
// The operation is performed atomically with a write lock to ensure consistency.
// Note that this is a manual cleanup operation - expired items are also checked
// during normal Get operations, but this method explicitly removes them from memory.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts (currently unused but kept for interface compatibility)
//
// Returns:
//   - error: Always nil for memory cache
//
// Example:
//
//	// Periodically clean up expired items
//	err := cache.Cleanup(ctx)
func (m *MemoryCache) Cleanup(_ context.Context) error {
	m.cleanup(func() {})

	return nil
}

// Delete removes the item associated with the given key from the memory cache.
//
// If the key does not exist, this operation performs no action and returns nil.
// The operation is safe for concurrent use by multiple goroutines.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts (currently unused but kept for interface compatibility)
//   - key: The key to remove from the cache
//
// Returns:
//   - error: Always nil for memory cache
//
// Example:
//
//	// Remove a specific key
//	err := cache.Delete(ctx, "user:123")
func (m *MemoryCache) Delete(_ context.Context, key string) error {
	m.mux.Lock()
	delete(m.items, key)
	m.mux.Unlock()

	return nil
}

// Drain returns a map of all non-expired items in the memory cache and clears the cache.
//
// The returned map is a snapshot of the cache at the time of the call.
// After this operation, the cache will be empty. This is useful for cache migration,
// backup, or when shutting down an application. The operation is performed atomically
// with a write lock to ensure consistency.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts (currently unused but kept for interface compatibility)
//
// Returns:
//   - map[string][]byte: A map containing all non-expired key-value pairs
//   - error: Always nil for memory cache
//
// Example:
//
//	// Get all items and clear the cache
//	items, err := cache.Drain(ctx)
//	for key, value := range items {
//	    log.Printf("Drained: %s = %s", key, string(value))
//	}
func (m *MemoryCache) Drain(_ context.Context) (map[string][]byte, error) {
	var cpy map[string]*memoryItem

	m.cleanup(func() {
		cpy = m.items
		m.items = make(map[string]*memoryItem)
	})

	items := make(map[string][]byte, len(cpy))
	for key, item := range cpy {
		items[key] = item.value
	}

	return items, nil
}

// Get retrieves the value for the given key from the memory cache.
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
//   - ctx: Context for cancellation and timeouts (currently unused but kept for interface compatibility)
//   - key: The key to retrieve
//   - opts: Optional operations to perform during retrieval (e.g., AndDelete, AndSetTTL)
//
// Returns:
//   - []byte: The cached value if found and not expired
//   - error: nil on success, ErrKeyNotFound if key doesn't exist,
//     ErrKeyExpired if key exists but has expired, otherwise an error
//
// Example:
//
//	// Simple get
//	value, err := cache.Get(ctx, "user:123")
//
//	// Get and extend TTL by 30 minutes
//	value, err := cache.Get(ctx, "session:abc", cache.AndUpdateTTL(30*time.Minute))
//
//	// Get and delete atomically
//	value, err := cache.Get(ctx, "temp:xyz", cache.AndDelete())
func (m *MemoryCache) Get(_ context.Context, key string, opts ...GetOption) ([]byte, error) {
	return m.getValue(func() (*memoryItem, bool) {
		if len(opts) == 0 {
			m.mux.RLock()
			item, ok := m.items[key]
			m.mux.RUnlock()

			return item, ok
		}

		o := new(getOptions)
		o.apply(opts...)

		m.mux.Lock()
		defer m.mux.Unlock()
		item, ok := m.items[key]

		if !ok || item.isExpired(time.Now()) {
			return item, ok
		}

		if o.delete {
			delete(m.items, key)
			return item, ok
		}

		switch {
		case o.validUntil != nil:
			item.validUntil = *o.validUntil
		case o.setTTL != nil:
			item.validUntil = time.Now().Add(*o.setTTL)
		case o.updateTTL != nil:
			if item.validUntil.IsZero() {
				item.validUntil = time.Now().Add(*o.updateTTL)
			} else {
				item.validUntil = item.validUntil.Add(*o.updateTTL)
			}
		case o.defaultTTL:
			if m.ttl > 0 {
				item.validUntil = time.Now().Add(m.ttl)
			} else {
				item.validUntil = time.Time{}
			}
		}

		return item, ok
	})
}

// GetAndDelete retrieves the value for the given key and atomically deletes it from the memory cache.
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
//     ErrKeyExpired if key exists but has expired, otherwise an error
//
// Example:
//
//	// Atomically get and remove a value
//	value, err := cache.GetAndDelete(ctx, "queue:item:123")
func (m *MemoryCache) GetAndDelete(ctx context.Context, key string) ([]byte, error) {
	return m.Get(ctx, key, AndDelete())
}

// Set stores the value for the given key in the memory cache, overwriting any existing value.
//
// The value will be stored with the default TTL configured for the cache,
// unless overridden by options. If the key already exists, its value and TTL will be updated.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts (currently unused but kept for interface compatibility)
//   - key: The key to store the value under
//   - value: The value to store as a byte slice
//   - opts: Optional configuration for this specific item (e.g., custom TTL)
//
// Returns:
//   - error: Always nil for memory cache
//
// Example:
//
//	// Set with default TTL
//	err := cache.Set(ctx, "user:123", []byte("user data"))
//
//	// Set with custom TTL
//	err := cache.Set(ctx, "session:abc", []byte("session data"), cache.WithTTL(30*time.Minute))
func (m *MemoryCache) Set(_ context.Context, key string, value []byte, opts ...Option) error {
	m.mux.Lock()
	m.items[key] = m.newItem(value, opts...)
	m.mux.Unlock()

	return nil
}

// SetOrFail stores the value for the given key only if the key does not already exist.
//
// This is an atomic operation that prevents race conditions when multiple goroutines
// might try to set the same key simultaneously. If the key exists but has expired,
// it will be overwritten.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts (currently unused but kept for interface compatibility)
//   - key: The key to store the value under
//   - value: The value to store as a byte slice
//   - opts: Optional configuration for this specific item (e.g., custom TTL)
//
// Returns:
//   - error: nil on success, ErrKeyExists if the key already exists and is not expired,
//     otherwise an error
//
// Example:
//
//	// Try to set a value only if key doesn't exist
//	err := cache.SetOrFail(ctx, "lock:resource", []byte("locked"))
//	if errors.Is(err, cache.ErrKeyExists) {
//	    // Key already exists, handle conflict
//	}
func (m *MemoryCache) SetOrFail(_ context.Context, key string, value []byte, opts ...Option) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	if item, ok := m.items[key]; ok {
		if !item.isExpired(time.Now()) {
			return ErrKeyExists
		}
	}

	m.items[key] = m.newItem(value, opts...)
	return nil
}

// newItem creates a new memory item with the specified value and options.
//
// This method applies the cache's default TTL if no explicit expiration is set
// in the options, and then creates a new memoryItem with the combined configuration.
//
// Parameters:
//   - value: The value to store in the item
//   - opts: Optional configuration for the item (e.g., custom TTL)
//
// Returns:
//   - *memoryItem: A new memory item with the specified configuration
func (m *MemoryCache) newItem(value []byte, opts ...Option) *memoryItem {
	o := options{
		validUntil: time.Time{},
	}
	if m.ttl > 0 {
		o.validUntil = time.Now().Add(m.ttl)
	}
	o.apply(opts...)

	return newMemoryItem(value, o)
}

// getItem retrieves a memory item using the provided getter function and checks for expiration.
//
// This is a helper method that wraps the getter function and adds expiration checking.
// It returns ErrKeyNotFound if the item doesn't exist, or ErrKeyExpired if it exists but has expired.
//
// Parameters:
//   - getter: A function that returns a memory item and a boolean indicating if it was found
//
// Returns:
//   - *memoryItem: The memory item if found and not expired
//   - error: nil on success, ErrKeyNotFound if key doesn't exist,
//     ErrKeyExpired if key exists but has expired
func (m *MemoryCache) getItem(getter func() (*memoryItem, bool)) (*memoryItem, error) {
	item, ok := getter()

	if !ok {
		return nil, ErrKeyNotFound
	}

	if item.isExpired(time.Now()) {
		return nil, ErrKeyExpired
	}

	return item, nil
}

// getValue retrieves the value of a memory item using the provided getter function.
//
// This is a helper method that uses getItem to get the memory item and then
// extracts the byte slice value from it.
//
// Parameters:
//   - getter: A function that returns a memory item and a boolean indicating if it was found
//
// Returns:
//   - []byte: The cached value if found and not expired
//   - error: nil on success, ErrKeyNotFound if key doesn't exist,
//     ErrKeyExpired if key exists but has expired
func (m *MemoryCache) getValue(getter func() (*memoryItem, bool)) ([]byte, error) {
	item, err := m.getItem(getter)
	if err != nil {
		return nil, err
	}

	return item.value, nil
}

// cleanup removes all expired items from the memory cache and executes the provided callback.
//
// This method scans through all items in the cache and removes any that have expired.
// The callback function is executed while holding the write lock, allowing for
// atomic operations during cleanup.
//
// Parameters:
//   - cb: A callback function to execute during cleanup (while holding the write lock)
func (m *MemoryCache) cleanup(cb func()) {
	t := time.Now()

	m.mux.Lock()
	for key, item := range m.items {
		if item.isExpired(t) {
			delete(m.items, key)
		}
	}

	cb()
	m.mux.Unlock()
}

// Close releases any resources held by the memory cache.
//
// For the memory cache implementation, this is a no-op since there are no
// external resources to clean up. The method is included for interface compatibility.
//
// Returns:
//   - error: Always nil for memory cache
//
// Example:
//
//	// Properly close the cache when done
//	defer cache.Close()
func (m *MemoryCache) Close() error {
	return nil
}

// Compile-time check to ensure memoryCache implements the Cache interface.
var _ Cache = (*MemoryCache)(nil)
