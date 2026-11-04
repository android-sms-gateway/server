package messages_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/android-sms-gateway/client-go/smsgateway"
	messages "github.com/android-sms-gateway/server/internal/sms-gateway/modules/messages"
	cacheImpl "github.com/go-core-fx/cachefx/cache"
	"github.com/stretchr/testify/require"
)

const (
	cacheTestUserID    = "user-cache"
	cacheTestMessageID = "msg-12345678901234"
)

func newCacheTestState(createdAt time.Time) *messages.MessageState {
	return &messages.MessageState{
		MessageStateInput: messages.MessageStateInput{
			ID:    cacheTestMessageID,
			State: messages.ProcessingStatePending,
			Recipients: []smsgateway.RecipientState{
				{PhoneNumber: "+1234567890", State: smsgateway.ProcessingStatePending},
			},
			States: map[string]time.Time{string(messages.ProcessingStatePending): createdAt},
		},
		MessageStateContent: messages.MessageStateContent{
			MessageContent: messages.MessageContent{
				TextContent: &messages.TextContent{Text: "test message"},
			},
		},
		DeviceID:  "device-12345678901234",
		CreatedAt: createdAt,
	}
}

// TestStateCacheGetIgnoresUnversionedEntry pins the cache key versioning: an entry
// written under the pre-versioning key must not be served, so the caller reloads
// the message from the database instead of returning a payload that predates the
// currently deployed fields (e.g. `scheduleAt`).
func TestStateCacheGetIgnoresUnversionedEntry(t *testing.T) {
	t.Parallel()

	storage := cacheImpl.NewMemory(time.Minute)
	c := messages.NewStateCacheForTest(messages.Config{CacheTTL: time.Minute}, storage)

	createdAt := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	data, err := json.Marshal(newCacheTestState(createdAt))
	require.NoError(t, err)

	legacyKey := cacheTestUserID + ":" + cacheTestMessageID
	require.NoError(t, storage.Set(context.Background(), legacyKey, data))

	state, err := c.Get(context.Background(), cacheTestUserID, cacheTestMessageID)
	require.Nil(t, state)
	require.ErrorIs(t, err, cacheImpl.ErrKeyNotFound)
}

// TestStateCacheRoundTripPreservesScheduleAt pins that a state cached through the
// current key layout returns its scheduled delivery time intact.
func TestStateCacheRoundTripPreservesScheduleAt(t *testing.T) {
	t.Parallel()

	c := messages.NewStateCacheForTest(messages.Config{CacheTTL: time.Minute}, cacheImpl.NewMemory(time.Minute))

	createdAt := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	scheduleAt := createdAt.Add(time.Hour)

	want := newCacheTestState(createdAt)
	want.ScheduleAt = &scheduleAt

	require.NoError(t, c.Set(context.Background(), cacheTestUserID, cacheTestMessageID, want))

	state, err := c.Get(context.Background(), cacheTestUserID, cacheTestMessageID)
	require.NoError(t, err)
	require.NotNil(t, state)
	require.NotNil(t, state.ScheduleAt)
	require.True(t, state.ScheduleAt.Equal(scheduleAt), "scheduleAt = %s, want %s", state.ScheduleAt, scheduleAt)
	require.True(t, state.CreatedAt.Equal(createdAt), "createdAt = %s, want %s", state.CreatedAt, createdAt)
}

// TestStateCacheSetNilCachesNotFound pins the nil sentinel: a nil state is stored as
// an empty value and read back as (nil, nil), which callers translate to "not found".
func TestStateCacheSetNilCachesNotFound(t *testing.T) {
	t.Parallel()

	c := messages.NewStateCacheForTest(messages.Config{CacheTTL: time.Minute}, cacheImpl.NewMemory(time.Minute))

	require.NoError(t, c.Set(context.Background(), cacheTestUserID, cacheTestMessageID, nil))

	state, err := c.Get(context.Background(), cacheTestUserID, cacheTestMessageID)
	require.NoError(t, err)
	require.Nil(t, state)
}

// TestStateCacheDeleteUsesVersionedKey pins that invalidation targets the same
// versioned key as Set, so a stale unversioned entry is never resurrected.
func TestStateCacheDeleteUsesVersionedKey(t *testing.T) {
	t.Parallel()

	storage := cacheImpl.NewMemory(time.Minute)
	c := messages.NewStateCacheForTest(messages.Config{CacheTTL: time.Minute}, storage)

	createdAt := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	require.NoError(t, c.Set(context.Background(), cacheTestUserID, cacheTestMessageID, newCacheTestState(createdAt)))
	require.NoError(t, c.Delete(context.Background(), cacheTestUserID, cacheTestMessageID))

	state, err := c.Get(context.Background(), cacheTestUserID, cacheTestMessageID)
	require.Nil(t, state)
	require.ErrorIs(t, err, cacheImpl.ErrKeyNotFound)
}

// TestStateCacheGetRejectsStalePayload pins the pre-existing zero-CreatedAt guard:
// payloads that decode without a creation time are reported as stale so the caller
// reloads them instead of renewing their expiry.
func TestStateCacheGetRejectsStalePayload(t *testing.T) {
	t.Parallel()

	storage := cacheImpl.NewMemory(time.Minute)
	c := messages.NewStateCacheForTest(messages.Config{CacheTTL: time.Minute}, storage)

	data, err := json.Marshal(newCacheTestState(time.Time{}))
	require.NoError(t, err)

	require.NoError(
		t,
		storage.Set(context.Background(), messages.StateCacheKeyForTest(cacheTestUserID, cacheTestMessageID), data),
	)

	state, err := c.Get(context.Background(), cacheTestUserID, cacheTestMessageID)
	require.Nil(t, state)
	require.ErrorIs(t, err, messages.ErrStaleCacheEntry)
}
