package cache_test

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	appcache "github.com/android-sms-gateway/server/internal/sms-gateway/cache"
	"github.com/go-core-fx/cachefx"
	"github.com/go-core-fx/cachefx/cache"
	"github.com/go-playground/assert/v2"
)

const (
	testInterval = 10 * time.Millisecond
	stubTTL      = time.Millisecond
	waitTimeout  = 2 * time.Second
)

func newFactory(t *testing.T) appcache.Factory {
	t.Helper()

	factory, err := appcache.NewTestFactory("sms-gateway")
	assert.Equal(t, nil, err)

	return factory
}

// TestSweeper_EvictsExpiredEntries checks that the sweeper asks the tracked
// caches to drop their expired entries.
//
// The assertion is made on the number of items the cache still holds, because
// the real cache implementations drop the expired items on their own when they
// are drained, so a drained cache cannot tell a swept cache from an unswept
// one.
func TestSweeper_EvictsExpiredEntries(t *testing.T) {
	t.Parallel()

	factory := appcache.NewTestFactoryWith(&stubFactory{})

	storage, err := factory.New("messages")
	assert.Equal(t, nil, err)

	stub, ok := storage.(*stubCache)
	assert.Equal(t, true, ok)

	// Fill the cache with entries that expire immediately.
	for i := range 100 {
		assert.Equal(t, nil, storage.Set(context.Background(), "user:ext-"+strconv.Itoa(i), []byte("v")))
	}

	assert.Equal(t, 100, stub.entryCount())

	sweeper := appcache.NewTestSweeper(factory, testInterval)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		_ = sweeper.Run(ctx)
	}()

	waitFor(t, func() bool { return stub.entryCount() == 0 })

	assert.Equal(t, true, stub.cleanupCount() > 0)
}

// TestFactory_NamedFactoryCachesAreSwept checks that a cache created through a
// named factory is swept by the factory it was derived from.
func TestFactory_NamedFactoryCachesAreSwept(t *testing.T) {
	t.Parallel()

	factory := appcache.NewTestFactoryWith(&stubFactory{})

	storage, err := factory.WithName("named").New("messages")
	assert.Equal(t, nil, err)

	stub, ok := storage.(*stubCache)
	assert.Equal(t, true, ok)

	assert.Equal(t, nil, storage.Set(context.Background(), "user:ext-1", []byte("v")))

	assert.Equal(t, 1, stub.entryCount())

	sweeper := appcache.NewTestSweeper(factory, testInterval)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		_ = sweeper.Run(ctx)
	}()

	waitFor(t, func() bool { return stub.entryCount() == 0 })
}

func TestSweeper_KeepsLiveEntries(t *testing.T) {
	t.Parallel()

	factory := newFactory(t)

	storage, err := factory.New("messages")
	assert.Equal(t, nil, err)

	sweeper := appcache.NewTestSweeper(factory, testInterval)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		_ = sweeper.Run(ctx)
	}()

	assert.Equal(t, nil, storage.Set(ctx, "live", []byte("v"), cache.WithTTL(time.Hour)))

	time.Sleep(100 * time.Millisecond)

	got, err := storage.Get(ctx, "live")
	assert.Equal(t, nil, err)
	assert.Equal(t, "v", string(got))
}

func TestFactory_TracksCachesPerInstance(t *testing.T) {
	t.Parallel()

	factory := newFactory(t)

	first, err := factory.New("messages")
	assert.Equal(t, nil, err)
	second, err := factory.New("messages")
	assert.Equal(t, nil, err)

	assert.Equal(t, nil, first.Set(context.Background(), "only-in-first", []byte("v")))
	_, err = second.Get(context.Background(), "only-in-first")
	assert.Equal(t, true, errors.Is(err, cache.ErrKeyNotFound))
}

// waitFor waits until the condition is met, and fails the test if it is not met
// in time.
func waitFor(t *testing.T, cond func() bool) {
	t.Helper()

	deadline := time.Now().Add(waitTimeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}

		time.Sleep(20 * time.Millisecond)
	}

	t.Fatal("condition was not met in time")
}

// stubCache is a cache that keeps its expired items until Cleanup removes them.
//
// The cachefx options are opaque outside of the cachefx package, so the
// lifetime of the items is defined by the stub itself.
type stubCache struct {
	mu       sync.Mutex
	items    map[string]stubItem
	cleanups int
}

type stubItem struct {
	value      []byte
	validUntil time.Time
}

func newStubCache() *stubCache {
	return &stubCache{
		mu:    sync.Mutex{},
		items: map[string]stubItem{},
	}
}

func (c *stubCache) Set(_ context.Context, key string, value []byte, _ ...cache.Option) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = stubItem{
		value:      value,
		validUntil: time.Now().Add(stubTTL),
	}

	return nil
}

func (c *stubCache) SetOrFail(_ context.Context, key string, value []byte, _ ...cache.Option) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.items[key]; ok {
		return cache.ErrKeyExists
	}

	c.items[key] = stubItem{
		value:      value,
		validUntil: time.Now().Add(stubTTL),
	}

	return nil
}

func (c *stubCache) Get(_ context.Context, key string, _ ...cache.GetOption) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, ok := c.items[key]
	if !ok || item.expired(time.Now()) {
		return nil, cache.ErrKeyNotFound
	}

	return item.value, nil
}

func (c *stubCache) GetAndDelete(ctx context.Context, key string) ([]byte, error) {
	value, err := c.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if err = c.Delete(ctx, key); err != nil {
		return nil, err
	}

	return value, nil
}

func (c *stubCache) Delete(_ context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)

	return nil
}

func (c *stubCache) Cleanup(_ context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.items {
		if item.expired(now) {
			delete(c.items, key)
		}
	}

	c.cleanups++

	return nil
}

func (c *stubCache) Drain(_ context.Context) (map[string][]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	items := make(map[string][]byte, len(c.items))
	for key, item := range c.items {
		if !item.expired(time.Now()) {
			items[key] = item.value
		}
	}

	c.items = map[string]stubItem{}

	return items, nil
}

func (c *stubCache) Close() error {
	return nil
}

func (c *stubCache) entryCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.items)
}

func (c *stubCache) cleanupCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.cleanups
}

func (i stubItem) expired(now time.Time) bool {
	return !i.validUntil.IsZero() && now.After(i.validUntil)
}

// stubFactory creates stub caches, so a test can observe how the sweeper treats
// the caches it tracks.
type stubFactory struct{}

func (f *stubFactory) New(_ string) (cache.Cache, error) {
	return newStubCache(), nil
}

func (f *stubFactory) WithName(_ string) cachefx.Factory {
	return f
}
