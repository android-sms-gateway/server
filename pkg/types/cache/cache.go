package cache

import (
	"sync"
	"time"
)

type item[T any] struct {
	value      T
	validUntil time.Time
}

func (i *item[T]) isExpired(now time.Time) bool {
	return !i.validUntil.IsZero() && now.After(i.validUntil)
}

type Cache[T any] struct {
	items map[string]*item[T]
	ttl   time.Duration
	empty T

	mux sync.RWMutex
}

func New[T any](cfg Config) *Cache[T] {
	return &Cache[T]{
		items: make(map[string]*item[T]),
		ttl:   cfg.TTL,
	}
}

func (c *Cache[T]) Set(key string, value T) error {
	var validUntil time.Time
	if c.ttl > 0 {
		validUntil = time.Now().Add(c.ttl)
	}

	c.mux.Lock()
	c.items[key] = &item[T]{
		value:      value,
		validUntil: validUntil,
	}
	c.mux.Unlock()

	return nil
}

func (c *Cache[T]) Get(key string) (T, error) {
	c.mux.RLock()
	item, ok := c.items[key]
	c.mux.RUnlock()

	if !ok {
		return c.empty, ErrKeyNotFound
	}

	if item.isExpired(time.Now()) {
		return c.empty, ErrKeyExpired
	}

	return item.value, nil
}

func (c *Cache[T]) Delete(key string) error {
	c.mux.Lock()
	delete(c.items, key)
	c.mux.Unlock()

	return nil
}

func (c *Cache[T]) Drain() map[string]T {
	t := time.Now()

	c.mux.Lock()
	copy := c.items
	c.items = make(map[string]*item[T], len(copy))
	c.mux.Unlock()

	items := make(map[string]T, len(copy))
	for key, item := range copy {
		if item.isExpired(t) {
			continue
		}
		items[key] = item.value
	}

	return items
}

func (c *Cache[T]) Cleanup() {
	t := time.Now()

	c.mux.Lock()
	defer c.mux.Unlock()

	for key, item := range c.items {
		if item.isExpired(t) {
			delete(c.items, key)
		}
	}
}
