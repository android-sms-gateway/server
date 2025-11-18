package cache

import "time"

// Option configures per-item cache behavior (e.g., expiry).
type Option func(*options)

type options struct {
	validUntil time.Time
}

func (o *options) apply(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

// WithTTL is an Option that sets the TTL (time to live) for an item, i.e. the
// item will expire after the given duration from the time of insertion.
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

// WithValidUntil is an Option that sets the valid until time for an item, i.e.
// the item will expire at the given time.
func WithValidUntil(validUntil time.Time) Option {
	return func(o *options) {
		o.validUntil = validUntil
	}
}

type getOptions struct {
	validUntil *time.Time
	setTTL     *time.Duration
	updateTTL  *time.Duration
	defaultTTL bool
	delete     bool
}

type GetOption func(*getOptions)

func (o *getOptions) apply(opts ...GetOption) {
	for _, opt := range opts {
		opt(o)
	}
}

func (o *getOptions) isEmpty() bool {
	return o.validUntil == nil &&
		o.setTTL == nil &&
		o.updateTTL == nil &&
		!o.defaultTTL &&
		!o.delete
}

func AndSetTTL(ttl time.Duration) GetOption {
	return func(o *getOptions) {
		o.setTTL = &ttl
	}
}

func AndUpdateTTL(ttl time.Duration) GetOption {
	return func(o *getOptions) {
		o.updateTTL = &ttl
	}
}

func AndSetValidUntil(validUntil time.Time) GetOption {
	return func(o *getOptions) {
		o.validUntil = &validUntil
	}
}

func AndDefaultTTL() GetOption {
	return func(o *getOptions) {
		o.defaultTTL = true
	}
}

func AndDelete() GetOption {
	return func(o *getOptions) {
		o.delete = true
	}
}
