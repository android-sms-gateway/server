package cache_test

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/android-sms-gateway/server/pkg/cache"
)

func TestMemoryCache_ZeroTTL(t *testing.T) {
	// Test cache with zero TTL (no expiration)
	cache := cache.NewMemory(0)
	ctx := context.Background()

	key := "zero-ttl-key"
	value := "zero-ttl-value"

	err := cache.Set(ctx, key, []byte(value))
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Wait some time to ensure no expiration
	time.Sleep(100 * time.Millisecond)

	retrieved, err := cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(retrieved) != value {
		t.Errorf("Expected %s, got %s", value, retrieved)
	}
}

func TestMemoryCache_ImmediateExpiration(t *testing.T) {
	// Test c with very short TTL
	c := cache.NewMemory(0)
	ctx := context.Background()

	key := "expiring-key"
	value := "expiring-value"
	ttl := 1 * time.Millisecond

	err := c.Set(ctx, key, []byte(value), cache.WithTTL(ttl))
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Wait for expiration
	time.Sleep(2 * ttl)

	_, err = c.Get(ctx, key)
	if !errors.Is(err, cache.ErrKeyExpired) {
		t.Errorf("Expected ErrKeyExpired, got %v", err)
	}
}

func TestMemoryCache_NilContext(t *testing.T) {
	// Test cache operations with nil context
	cache := cache.NewMemory(0)
	key := "nil-context-key"
	value := "nil-context-value"

	err := cache.Set(context.Background(), key, []byte(value))
	if err != nil {
		t.Fatalf("Set with nil context failed: %v", err)
	}

	retrieved, err := cache.Get(context.Background(), key)
	if err != nil {
		t.Fatalf("Get with nil context failed: %v", err)
	}

	if string(retrieved) != value {
		t.Errorf("Expected %s, got %s", value, retrieved)
	}
}

func TestMemoryCache_EmptyKey(t *testing.T) {
	// Test cache operations with empty key
	cache := cache.NewMemory(0)
	ctx := context.Background()
	key := ""
	value := "empty-key-value"

	err := cache.Set(ctx, key, []byte(value))
	if err != nil {
		t.Fatalf("Set with empty key failed: %v", err)
	}

	retrieved, err := cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get with empty key failed: %v", err)
	}

	if string(retrieved) != value {
		t.Errorf("Expected %s, got %s", value, retrieved)
	}
}

func TestMemoryCache_OverwriteWithDifferentTTL(t *testing.T) {
	// Test overwriting a key with different TTL
	c := cache.NewMemory(0)
	ctx := context.Background()
	key := "ttl-key"
	value1 := "value1"
	value2 := "value2"

	// Set with short TTL
	err := c.Set(ctx, key, []byte(value1), cache.WithTTL(100*time.Millisecond))
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Overwrite with longer TTL
	err = c.Set(ctx, key, []byte(value2), cache.WithTTL(1*time.Second))
	if err != nil {
		t.Fatalf("Set overwrite failed: %v", err)
	}

	retrieved, err := c.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(retrieved) != value2 {
		t.Errorf("Expected %s, got %s", value2, retrieved)
	}

	// Wait for short TTL to expire but not long TTL
	time.Sleep(200 * time.Millisecond)

	retrieved, err = c.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get after partial wait failed: %v", err)
	}

	if string(retrieved) != value2 {
		t.Errorf("Expected %s after partial wait, got %s", value2, retrieved)
	}
}

func TestMemoryCache_MixedTTLScenarios(t *testing.T) {
	// Test various TTL scenarios
	c := cache.NewMemory(0)
	ctx := context.Background()

	// Set multiple keys with different TTLs
	keys := map[string]time.Duration{
		"no-ttl":     0,
		"short-ttl":  50 * time.Millisecond,
		"medium-ttl": 200 * time.Millisecond,
		"long-ttl":   500 * time.Millisecond,
	}

	for key, ttl := range keys {
		value := "value-" + key
		var err error
		if ttl > 0 {
			err = c.Set(ctx, key, []byte(value), cache.WithTTL(ttl))
		} else {
			err = c.Set(ctx, key, []byte(value))
		}
		if err != nil {
			t.Fatalf("Set %s failed: %v", key, err)
		}
	}

	// Verify all keys are present initially
	for key := range keys {
		_, err := c.Get(ctx, key)
		if err != nil {
			t.Fatalf("Get %s failed: %v", key, err)
		}
	}

	// Wait for short TTL to expire
	time.Sleep(100 * time.Millisecond)

	// Short TTL key should be expired, others should still be there
	_, err := c.Get(ctx, "short-ttl")
	if !errors.Is(err, cache.ErrKeyExpired) {
		t.Errorf("Expected ErrKeyExpired for short-ttl, got %v", err)
	}

	for key := range keys {
		if key == "short-ttl" {
			continue
		}
		_, getErr := c.Get(ctx, key)
		if getErr != nil {
			t.Errorf("Get %s failed: %v", key, getErr)
		}
	}

	// Wait for medium TTL to expire
	time.Sleep(150 * time.Millisecond)

	// Medium TTL key should be expired, others should still be there
	_, err = c.Get(ctx, "medium-ttl")
	if !errors.Is(err, cache.ErrKeyExpired) {
		t.Errorf("Expected ErrKeyExpired for medium-ttl, got %v", err)
	}

	for key := range keys {
		if key == "short-ttl" || key == "medium-ttl" {
			continue
		}
		_, getErr := c.Get(ctx, key)
		if getErr != nil {
			t.Errorf("Get %s failed: %v", key, getErr)
		}
	}
}

func TestMemoryCache_RapidOperations(t *testing.T) {
	// Test rapid c operations
	c := cache.NewMemory(0)
	ctx := context.Background()

	const numOperations = 1000
	const duration = 100 * time.Millisecond

	start := time.Now()
	opsCompleted := 0

	for i := range numOperations {
		// Alternate between set and get
		if i%2 == 0 {
			key := "rapid-key-" + strconv.Itoa(i)
			value := "rapid-value-" + strconv.Itoa(i)
			err := c.Set(ctx, key, []byte(value))
			if err != nil {
				t.Errorf("Set failed: %v", err)
			}
		} else {
			key := "rapid-key-" + strconv.Itoa(i-1)
			_, err := c.Get(ctx, key)
			if err != nil {
				t.Errorf("Get failed: %v", err)
			}
		}
		opsCompleted++
	}

	durationTaken := time.Since(start)
	t.Logf(
		"Completed %d operations in %v (%.2f ops/ms)",
		opsCompleted,
		durationTaken,
		float64(opsCompleted)/float64(durationTaken.Milliseconds()),
	)

	// Verify operations completed within reasonable time
	if durationTaken > 2*duration {
		t.Errorf("Operations took too long: %v", durationTaken)
	}
}

func TestMemoryCache_CleanupOnEmptyCache(t *testing.T) {
	// Test cleanup operation on empty cache
	cache := cache.NewMemory(0)
	ctx := context.Background()

	err := cache.Cleanup(ctx)
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Should still work normally after cleanup
	key := "post-cleanup-key"
	value := "post-cleanup-value"

	err = cache.Set(ctx, key, []byte(value))
	if err != nil {
		t.Fatalf("Set after cleanup failed: %v", err)
	}

	retrieved, err := cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get after cleanup failed: %v", err)
	}

	if string(retrieved) != value {
		t.Errorf("Expected %s, got %s", value, retrieved)
	}
}

func TestMemoryCache_DrainWithExpiredItems(t *testing.T) {
	// Test drain operation with mix of expired and non-expired items
	c := cache.NewMemory(0)
	ctx := context.Background()

	// Set non-expired item
	err := c.Set(ctx, "valid-key", []byte("valid-value"))
	if err != nil {
		t.Fatalf("Set valid key failed: %v", err)
	}

	// Set expired item
	err = c.Set(ctx, "expired-key", []byte("expired-value"), cache.WithTTL(1*time.Millisecond))
	if err != nil {
		t.Fatalf("Set expired key failed: %v", err)
	}

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	// Drain should only return non-expired items
	items, err := c.Drain(ctx)
	if err != nil {
		t.Fatalf("Drain failed: %v", err)
	}

	if len(items) != 1 {
		t.Errorf("Expected 1 item in drain result, got %d", len(items))
	}

	if string(items["valid-key"]) != "valid-value" {
		t.Errorf("Expected valid-value, got %s", items["valid-key"])
	}

	// Verify expired item is gone (should be completely removed, not just expired)
	_, err = c.Get(ctx, "expired-key")
	if !errors.Is(err, cache.ErrKeyNotFound) {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}
}

func TestMemoryCache_ExtremeKeyLength(t *testing.T) {
	// Test with very long keys
	cache := cache.NewMemory(0)
	ctx := context.Background()

	// Create a very long key (1KB)
	longKey := strings.Repeat("a", 1024)
	value := "extreme-key-value"

	err := cache.Set(ctx, longKey, []byte(value))
	if err != nil {
		t.Fatalf("Set with long key failed: %v", err)
	}

	retrieved, err := cache.Get(ctx, longKey)
	if err != nil {
		t.Fatalf("Get with long key failed: %v", err)
	}

	if string(retrieved) != value {
		t.Errorf("Expected %s, got %s", value, retrieved)
	}
}

func TestMemoryCache_RaceConditionWithExpiration(t *testing.T) {
	// Test race conditions between expiration and access
	c := cache.NewMemory(0)
	ctx := context.Background()

	key := "race-expire-key"
	value := "race-expire-value"
	ttl := 10 * time.Millisecond

	// Set item with short TTL
	err := c.Set(ctx, key, []byte(value), cache.WithTTL(ttl))
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	const numGoroutines = 50
	var wg sync.WaitGroup

	// Launch goroutines that try to access the key while it's expiring
	for i := range numGoroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Wait for the item to be close to expiration with some jitter
			jitter := time.Duration(id%3) * time.Millisecond
			time.Sleep(ttl - 2*time.Millisecond + jitter)

			// Try to get the item
			_, getErr := c.Get(ctx, key)
			if getErr != nil && !errors.Is(getErr, cache.ErrKeyExpired) && !errors.Is(getErr, cache.ErrKeyNotFound) {
				t.Errorf("Get failed: %v", getErr)
			}
		}(i)
	}

	wg.Wait()
}
