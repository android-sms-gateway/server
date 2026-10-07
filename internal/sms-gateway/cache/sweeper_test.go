package cache_test

import (
	"context"
	"errors"
	"testing"
	"time"

	appcache "github.com/android-sms-gateway/server/internal/sms-gateway/cache"
	"github.com/go-core-fx/cachefx/cache"
	"github.com/go-playground/assert/v2"
)

const testInterval = 10 * time.Millisecond

func newFactory(t *testing.T) appcache.Factory {
	t.Helper()

	factory, err := appcache.NewTestFactory("sms-gateway")
	assert.Equal(t, nil, err)

	return factory
}

func TestSweeper_EvictsExpiredEntries(t *testing.T) {
	factory := newFactory(t)

	storage, err := factory.New("messages")
	assert.Equal(t, nil, err)

	sweeper := appcache.NewTestSweeper(factory, testInterval)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		_ = sweeper.Run(ctx)
	}()

	// Fill the cache with entries that expire immediately.
	for i := range 100 {
		setErr := storage.Set(ctx, "user:ext-"+string(rune('a'+i%26)), []byte("v"), cache.WithTTL(time.Millisecond))
		assert.Equal(t, nil, setErr)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		// Drain returns only non-expired items; an empty result means the
		// sweeper has evicted the expired ones.
		items, drainErr := storage.Drain(ctx)
		assert.Equal(t, nil, drainErr)

		if len(items) == 0 {
			return
		}

		time.Sleep(20 * time.Millisecond)
	}

	t.Fatal("expired entries were not evicted by the sweeper")
}

func TestSweeper_KeepsLiveEntries(t *testing.T) {
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
	factory := newFactory(t)

	first, err := factory.New("messages")
	assert.Equal(t, nil, err)
	second, err := factory.New("messages")
	assert.Equal(t, nil, err)

	assert.Equal(t, nil, first.Set(context.Background(), "only-in-first", []byte("v")))
	_, err = second.Get(context.Background(), "only-in-first")
	assert.Equal(t, true, errors.Is(err, cache.ErrKeyNotFound))
}
