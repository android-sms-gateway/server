package cache

import "time"

// Option configures per-item cache behavior (e.g., expiry).
type Option func(*options)

type options struct {
	validUntil time.Time
}

func (o *options) apply(opts ...Option) *options {
	for _, opt := range opts {
		opt(o)
	}

	return o
}

// WithTTL is an Option that sets the TTL (time to live) for an item, i.e. the
// item will expire after the given duration from the time of insertion.
func WithTTL(ttl time.Duration) Option {
	return func(o *options) {
		if ttl <= 0 {
			o.validUntil = time.Time{}
		}

		o.validUntil = time.Now().Add(ttl)
	}
}

// WithValidUntil is an Option that sets the valid until time for an item, i.e.
// the item will expire at the given time.
func WithValidUntil(validUntil time.Time) Option {
	return func(o *options) {
		o.validUntil = validUntil
	}
}
