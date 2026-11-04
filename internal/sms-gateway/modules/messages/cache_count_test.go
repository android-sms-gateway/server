package messages_test

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	messages "github.com/android-sms-gateway/server/internal/sms-gateway/modules/messages"
	cacheImpl "github.com/go-core-fx/cachefx/cache"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestCountCacheKeyIsStableAndScoped pins the key layout: identical filters
// collide, while every field that changes the counted row set - user, device,
// id, date range and states - gets its own key.
func TestCountCacheKeyIsStableAndScoped(t *testing.T) {
	t.Parallel()

	base := messages.SelectFilter{
		UserID:   "user-key",
		DeviceID: "device-key",
	}
	baseNoDevice := base
	baseNoDevice.DeviceID = "other-device"

	otherUser := base
	otherUser.UserID = "other-user"

	otherExtID := base
	otherExtID.ExtID = "ext-1"

	otherRange := base
	otherRange.StartDate = time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC)

	otherStates := base
	otherStates.State = []messages.ProcessingState{messages.ProcessingStateProcessed}

	key := base.CountCacheKeyForTest()
	require.True(
		t,
		strings.HasPrefix(key, messages.CountCacheKeyVersionForTest),
		"key %q must carry the version prefix",
		key,
	)
	require.Equal(t, key, base.CountCacheKeyForTest(), "identical filters must share one key")
	require.NotEqual(t, key, baseNoDevice.CountCacheKeyForTest())
	require.NotEqual(t, key, otherUser.CountCacheKeyForTest())
	require.NotEqual(t, key, otherExtID.CountCacheKeyForTest())
	require.NotEqual(t, key, otherRange.CountCacheKeyForTest())
	require.NotEqual(t, key, otherStates.CountCacheKeyForTest())
}

// TestCountCacheServesCachedTotalWithinTTL pins that a second call inside the
// TTL is served from the cache instead of repeating the database work.
func TestCountCacheServesCachedTotalWithinTTL(t *testing.T) {
	t.Parallel()

	c := messages.NewCountCacheForTest(
		messages.Config{ListCountCacheTTL: time.Minute},
		cacheImpl.NewMemory(time.Minute),
		zap.NewNop(),
	)

	var calls int
	fetch := func() (int64, error) {
		calls++
		return 1836, nil
	}

	total, hit, err := c.Count("count-cached-key", fetch)
	require.NoError(t, err)
	require.False(t, hit)
	require.Equal(t, int64(1836), total)

	total, hit, err = c.Count("count-cached-key", fetch)
	require.NoError(t, err)
	require.True(t, hit)
	require.Equal(t, int64(1836), total)
	require.Equal(t, 1, calls, "the second call must not reach the database")
}

// TestCountCacheDoesNotRenewTTLOnRead pins the fixed expiry: reads must not
// push the deadline forward, otherwise a continuously polled total would never
// be refreshed and could pin a stale value forever.
func TestCountCacheDoesNotRenewTTLOnRead(t *testing.T) {
	t.Parallel()

	c := messages.NewCountCacheForTest(
		messages.Config{ListCountCacheTTL: 30 * time.Millisecond},
		cacheImpl.NewMemory(time.Minute),
		zap.NewNop(),
	)

	var calls int
	fetch := func() (int64, error) {
		calls++
		return int64(calls), nil
	}

	_, _, err := c.Count("count-expiring-key", fetch)
	require.NoError(t, err)
	require.Equal(t, 1, calls)

	_, _, err = c.Count("count-expiring-key", fetch)
	require.NoError(t, err)
	require.Equal(t, 1, calls, "a read inside the TTL must not count as a miss")

	time.Sleep(60 * time.Millisecond)

	_, hit, err := c.Count("count-expiring-key", fetch)
	require.NoError(t, err)
	require.False(t, hit, "an expired total must not be served")
	require.Equal(t, 2, calls, "the total must be fetched again once its TTL passed")
}

// TestCountCacheCoalescesConcurrentMisses pins the singleflight behaviour: a
// burst of simultaneous polls for one key produces exactly one fetch.
func TestCountCacheCoalescesConcurrentMisses(t *testing.T) {
	t.Parallel()

	c := messages.NewCountCacheForTest(
		messages.Config{ListCountCacheTTL: time.Minute},
		cacheImpl.NewMemory(time.Minute),
		zap.NewNop(),
	)

	var calls atomic.Int64
	fetch := func() (int64, error) {
		if calls.Add(1) > 1 {
			return 0, errors.New("list total was fetched more than once")
		}
		time.Sleep(20 * time.Millisecond)
		return 42, nil
	}

	type result struct {
		total int64
		err   error
	}

	const workers = 16
	results := make(chan result, workers)
	for range workers {
		go func() {
			total, _, fetchErr := c.Count("count-burst-key", fetch)
			results <- result{total: total, err: fetchErr}
		}()
	}

	for range workers {
		got := <-results
		require.NoError(t, got.err)
		require.Equal(t, int64(42), got.total, "every caller must see the coalesced value")
	}
	require.Equal(t, int64(1), calls.Load(), "the burst must trigger a single fetch")
}

// TestCountCacheDoesNotStoreFetchError pins that a failed count is not cached,
// so the next poll retries instead of serving a poisoned entry.
func TestCountCacheDoesNotStoreFetchError(t *testing.T) {
	t.Parallel()

	c := messages.NewCountCacheForTest(
		messages.Config{ListCountCacheTTL: time.Minute},
		cacheImpl.NewMemory(time.Minute),
		zap.NewNop(),
	)

	var calls int
	fetchErr := errors.New("database unavailable")
	fetch := func() (int64, error) {
		calls++
		if calls == 1 {
			return 0, fetchErr
		}
		return 7, nil
	}

	_, _, err := c.Count("count-error-key", fetch)
	require.ErrorIs(t, err, fetchErr)

	total, hit, err := c.Count("count-error-key", fetch)
	require.NoError(t, err)
	require.False(t, hit, "a failed count must not leave an entry behind")
	require.Equal(t, int64(7), total)
	require.Equal(t, 2, calls)
}

// TestCountCacheDisabledWhenTTLOverrides pins that a non-positive TTL stores
// nothing while the coalescing still applies.
func TestCountCacheDisabledWhenTTLOverrides(t *testing.T) {
	t.Parallel()

	storage := cacheImpl.NewMemory(time.Minute)
	c := messages.NewCountCacheForTest(messages.Config{ListCountCacheTTL: 0}, storage, zap.NewNop())

	var calls int
	fetch := func() (int64, error) {
		calls++
		return int64(calls * 10), nil
	}

	total, hit, err := c.Count("count-disabled-key", fetch)
	require.NoError(t, err)
	require.False(t, hit)
	require.Equal(t, int64(10), total)

	total, hit, err = c.Count("count-disabled-key", fetch)
	require.NoError(t, err)
	require.False(t, hit, "a disabled cache must not serve anything")
	require.Equal(t, int64(20), total)
	require.Equal(t, 2, calls)

	_, err = storage.Get(context.Background(), "count-disabled-key")
	require.ErrorIs(t, err, cacheImpl.ErrKeyNotFound, "a disabled cache must not store anything")
}
