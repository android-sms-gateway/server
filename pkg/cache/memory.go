package cache

import (
	"context"
	"sync"
	"time"
)

type memoryCache struct {
	items map[string]*memoryItem
	ttl   time.Duration

	mux sync.RWMutex
}

func NewMemory(ttl time.Duration) *memoryCache {
	return &memoryCache{
		items: make(map[string]*memoryItem),
		ttl:   ttl,

		mux: sync.RWMutex{},
	}
}

type memoryItem struct {
	value      []byte
	validUntil time.Time
}

func newItem(value []byte, opts options) *memoryItem {
	item := &memoryItem{
		value:      value,
		validUntil: opts.validUntil,
	}

	return item
}

func (i *memoryItem) isExpired(now time.Time) bool {
	if i == nil {
		return true
	}

	return !i.validUntil.IsZero() && now.After(i.validUntil)
}

// Cleanup implements Cache.
func (m *memoryCache) Cleanup(_ context.Context) error {
	m.cleanup(func() {})

	return nil
}

// Delete implements Cache.
func (m *memoryCache) Delete(_ context.Context, key string) error {
	m.mux.Lock()
	delete(m.items, key)
	m.mux.Unlock()

	return nil
}

// Drain implements Cache.
func (m *memoryCache) Drain(_ context.Context) (map[string][]byte, error) {
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

// Get implements Cache.
func (m *memoryCache) Get(_ context.Context, key string, opts ...GetOption) ([]byte, error) {
	return m.getValue(func() (*memoryItem, bool) {
		if len(opts) == 0 {
			m.mux.RLock()
			item, ok := m.items[key]
			m.mux.RUnlock()

			return item, ok
		}

		o := getOptions{}
		o.apply(opts...)

		m.mux.Lock()
		item, ok := m.items[key]
		if !ok {
			// item not found, nothing to do
		} else if o.delete {
			delete(m.items, key)
		} else if !item.isExpired(time.Now()) {
			if o.validUntil != nil {
				item.validUntil = *o.validUntil
			} else if o.setTTL != nil {
				item.validUntil = time.Now().Add(*o.setTTL)
			} else if o.updateTTL != nil {
				item.validUntil = item.validUntil.Add(*o.updateTTL)
			} else if o.defaultTTL {
				item.validUntil = time.Now().Add(m.ttl)
			}
		}
		m.mux.Unlock()

		return item, ok
	})
}

// GetAndDelete implements Cache.
func (m *memoryCache) GetAndDelete(ctx context.Context, key string) ([]byte, error) {
	return m.Get(ctx, key, AndDelete())
}

// Set implements Cache.
func (m *memoryCache) Set(_ context.Context, key string, value []byte, opts ...Option) error {
	m.mux.Lock()
	m.items[key] = m.newItem(value, opts...)
	m.mux.Unlock()

	return nil
}

// SetOrFail implements Cache.
func (m *memoryCache) SetOrFail(_ context.Context, key string, value []byte, opts ...Option) error {
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

func (m *memoryCache) newItem(value []byte, opts ...Option) *memoryItem {
	o := options{
		validUntil: time.Time{},
	}
	if m.ttl > 0 {
		o.validUntil = time.Now().Add(m.ttl)
	}
	o.apply(opts...)

	return newItem(value, o)
}

func (m *memoryCache) getItem(getter func() (*memoryItem, bool)) (*memoryItem, error) {
	item, ok := getter()

	if !ok {
		return nil, ErrKeyNotFound
	}

	if item.isExpired(time.Now()) {
		return nil, ErrKeyExpired
	}

	return item, nil
}

func (m *memoryCache) getValue(getter func() (*memoryItem, bool)) ([]byte, error) {
	item, err := m.getItem(getter)
	if err != nil {
		return nil, err
	}

	return item.value, nil
}

func (m *memoryCache) cleanup(cb func()) {
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

var _ Cache = (*memoryCache)(nil)
