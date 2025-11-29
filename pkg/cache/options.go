package cache

import "time"

// Option configures per-item cache behavior (e.g., expiry).
//
// Options are used with Set and SetOrFail operations to customize how individual
// cache items are stored, including their expiration time and TTL.
type Option func(*options)

// options holds the configuration for a cache item.
type options struct {
	validUntil time.Time
}

// apply applies the given options to this options struct.
func (o *options) apply(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

// WithTTL sets the TTL (time to live) for an item.
//
// The item will expire after the given duration from the time of insertion.
// A TTL of zero means the item will not expire.
// A negative TTL means the item expires immediately.
//
// Parameters:
//   - ttl: The duration after which the item will expire
//
// Returns:
//   - Option: An option that sets the TTL when passed to Set or SetOrFail
//
// Example:
//
//	// Set a value that expires in 30 minutes
//	err := cache.Set(ctx, "session:abc", []byte("data"), cache.WithTTL(30*time.Minute))
func WithTTL(ttl time.Duration) Option {
	return func(o *options) {
		switch {
		case ttl == 0:
			o.validUntil = time.Time{}
		case ttl < 0:
			o.validUntil = time.Now()
		default:
			o.validUntil = time.Now().Add(ttl)
		}
	}
}

// WithValidUntil sets the exact expiration time for an item.
//
// The item will expire at the given time, regardless of when it was inserted.
// This is useful when you need precise control over when an item expires,
// such as at midnight or the end of a billing period.
//
// Parameters:
//   - validUntil: The exact time when the item should expire
//
// Returns:
//   - Option: An option that sets the expiration time when passed to Set or SetOrFail
//
// Example:
//
//	// Set a value that expires at midnight
//	midnight := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local)
//	midnight = midnight.Add(24 * time.Hour) // Next midnight
//	err := cache.Set(ctx, "daily:report", []byte("data"), cache.WithValidUntil(midnight))
func WithValidUntil(validUntil time.Time) Option {
	return func(o *options) {
		o.validUntil = validUntil
	}
}

// getOptions holds the configuration for Get operations.
type getOptions struct {
	validUntil *time.Time
	setTTL     *time.Duration
	updateTTL  *time.Duration
	defaultTTL bool
	delete     bool
}

// GetOption configures the behavior of Get operations.
//
// GetOptions allow you to perform additional operations during a Get call,
// such as updating the item's TTL, setting a new expiration time, or
// deleting the item after retrieval.
type GetOption func(*getOptions)

// apply applies the given GetOptions to this getOptions struct.
func (o *getOptions) apply(opts ...GetOption) {
	for _, opt := range opts {
		opt(o)
	}
}

// isEmpty returns true if no GetOptions are set.
func (o *getOptions) isEmpty() bool {
	return o.validUntil == nil &&
		o.setTTL == nil &&
		o.updateTTL == nil &&
		!o.defaultTTL &&
		!o.delete
}

// AndSetTTL sets a new TTL for the item during a Get operation.
//
// This option is useful for extending the lifetime of an item when it's accessed,
// implementing a "touch" behavior where frequently accessed items remain in cache longer.
//
// Parameters:
//   - ttl: The new TTL duration to set for the item
//
// Returns:
//   - GetOption: An option that sets the TTL when passed to Get
//
// Example:
//
//	// Get a value and extend its TTL to 30 minutes from now
//	value, err := cache.Get(ctx, "session:abc", cache.AndSetTTL(30*time.Minute))
func AndSetTTL(ttl time.Duration) GetOption {
	return func(o *getOptions) {
		o.setTTL = &ttl
	}
}

// AndUpdateTTL extends the current TTL of an item by the given duration.
//
// Unlike AndSetTTL which sets an absolute TTL from the current time,
// AndUpdateTTL adds the specified duration to the item's existing TTL.
// This is useful for incrementally extending the lifetime of an item.
//
// Parameters:
//   - ttl: The duration to add to the item's current TTL
//
// Returns:
//   - GetOption: An option that extends the TTL when passed to Get
//
// Example:
//
//	// Get a value and extend its TTL by 15 minutes
//	value, err := cache.Get(ctx, "session:abc", cache.AndUpdateTTL(15*time.Minute))
func AndUpdateTTL(ttl time.Duration) GetOption {
	return func(o *getOptions) {
		o.updateTTL = &ttl
	}
}

// AndSetValidUntil sets a new exact expiration time for the item during a Get operation.
//
// This option allows you to set a precise expiration time for an item when it's accessed,
// which can be useful for implementing time-based access patterns.
//
// Parameters:
//   - validUntil: The new exact expiration time for the item
//
// Returns:
//   - GetOption: An option that sets the expiration time when passed to Get
//
// Example:
//
//	// Get a value and set it to expire at midnight
//	midnight := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local)
//	midnight = midnight.Add(24 * time.Hour) // Next midnight
//	value, err := cache.Get(ctx, "daily:report", cache.AndSetValidUntil(midnight))
func AndSetValidUntil(validUntil time.Time) GetOption {
	return func(o *getOptions) {
		o.validUntil = &validUntil
	}
}

// AndDefaultTTL resets the item's TTL to the cache's default TTL during a Get operation.
//
// This option is useful when you want to restore an item to the default expiration
// policy of the cache, regardless of its current TTL.
//
// Returns:
//   - GetOption: An option that resets the TTL to the cache default when passed to Get
//
// Example:
//
//	// Get a value and reset its TTL to the cache default
//	value, err := cache.Get(ctx, "session:abc", cache.AndDefaultTTL())
func AndDefaultTTL() GetOption {
	return func(o *getOptions) {
		o.defaultTTL = true
	}
}

// AndDelete deletes the item from the cache during a Get operation.
//
// This option provides an atomic "get and delete" operation, which is useful
// for implementing queue-like behavior or ensuring that an item is only
// processed once. This is equivalent to calling GetAndDelete.
//
// Returns:
//   - GetOption: An option that deletes the item when passed to Get
//
// Example:
//
//	// Get a value and remove it from the cache
//	value, err := cache.Get(ctx, "queue:item:123", cache.AndDelete())
func AndDelete() GetOption {
	return func(o *getOptions) {
		o.delete = true
	}
}
