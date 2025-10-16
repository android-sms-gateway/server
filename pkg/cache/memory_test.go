package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/android-sms-gateway/server/pkg/cache"
)

func TestMemoryCache_SetAndGet(t *testing.T) {
	cache := cache.NewMemory(0) // No TTL for basic tests

	ctx := context.Background()
	key := "test-key"
	value := "test-value"

	// Test setting a value
	err := cache.Set(ctx, key, []byte(value))
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Test getting the value
	retrieved, err := cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(retrieved) != value {
		t.Errorf("Expected %s, got %s", value, retrieved)
	}
}

func TestMemoryCache_SetAndGetWithTTL(t *testing.T) {
	c := cache.NewMemory(0) // No default TTL

	ctx := context.Background()
	key := "test-key"
	value := "test-value"
	ttl := 2 * time.Hour

	// Test setting a value with TTL
	err := c.Set(ctx, key, []byte(value), cache.WithTTL(ttl))
	if err != nil {
		t.Fatalf("Set with TTL failed: %v", err)
	}

	// Test getting the value
	retrieved, err := c.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(retrieved) != value {
		t.Errorf("Expected %s, got %s", value, retrieved)
	}
}

func TestMemoryCache_SetAndGetWithValidUntil(t *testing.T) {
	c := cache.NewMemory(0) // No default TTL

	ctx := context.Background()
	key := "test-key"
	value := "test-value"
	validUntil := time.Now().Add(2 * time.Hour)

	// Test setting a value with validUntil
	err := c.Set(ctx, key, []byte(value), cache.WithValidUntil(validUntil))
	if err != nil {
		t.Fatalf("Set with validUntil failed: %v", err)
	}

	// Test getting the value
	retrieved, err := c.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(retrieved) != value {
		t.Errorf("Expected %s, got %s", value, retrieved)
	}
}

func TestMemoryCache_SetAndGetWithDefaultTTL(t *testing.T) {
	defaultTTL := 1 * time.Hour
	cache := cache.NewMemory(defaultTTL) // With default TTL

	ctx := context.Background()
	key := "test-key"
	value := "test-value"

	// Test setting a value without explicit TTL (should use default)
	err := cache.Set(ctx, key, []byte(value))
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Test getting the value
	retrieved, err := cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(retrieved) != value {
		t.Errorf("Expected %s, got %s", value, retrieved)
	}
}

func TestMemoryCache_GetNotFound(t *testing.T) {
	c := cache.NewMemory(0)

	ctx := context.Background()
	key := "non-existent-key"

	_, err := c.Get(ctx, key)
	if err != cache.ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}
}

func TestMemoryCache_SetOrFailNewKey(t *testing.T) {
	cache := cache.NewMemory(0)

	ctx := context.Background()
	key := "test-key"
	value := "test-value"

	// Test SetOrFail with new key
	err := cache.SetOrFail(ctx, key, []byte(value))
	if err != nil {
		t.Fatalf("SetOrFail failed: %v", err)
	}

	// Verify the value was set
	retrieved, err := cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(retrieved) != value {
		t.Errorf("Expected %s, got %s", value, retrieved)
	}
}

func TestMemoryCache_SetOrFailExistingKey(t *testing.T) {
	c := cache.NewMemory(0)

	ctx := context.Background()
	key := "test-key"
	value1 := "value1"
	value2 := "value2"

	// Set initial value
	err := c.Set(ctx, key, []byte(value1))
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Try SetOrFail with existing key
	err = c.SetOrFail(ctx, key, []byte(value2))
	if err != cache.ErrKeyExists {
		t.Errorf("Expected ErrKeyExists, got %v", err)
	}

	// Verify original value is still there
	retrieved, err := c.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(retrieved) != value1 {
		t.Errorf("Expected %s, got %s", value1, retrieved)
	}
}

func TestMemoryCache_Delete(t *testing.T) {
	c := cache.NewMemory(0)

	ctx := context.Background()
	key := "test-key"
	value := "test-value"

	// Set a value
	err := c.Set(ctx, key, []byte(value))
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Delete the key
	err = c.Delete(ctx, key)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify the key is gone
	_, err = c.Get(ctx, key)
	if err != cache.ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound after delete, got %v", err)
	}
}

func TestMemoryCache_DeleteNonExistent(t *testing.T) {
	cache := cache.NewMemory(0)

	ctx := context.Background()
	key := "non-existent-key"

	// Delete non-existent key should not fail
	err := cache.Delete(ctx, key)
	if err != nil {
		t.Errorf("Delete of non-existent key failed: %v", err)
	}
}

func TestMemoryCache_GetAndDelete(t *testing.T) {
	c := cache.NewMemory(0)

	ctx := context.Background()
	key := "test-key"
	value := "test-value"

	// Set a value
	err := c.Set(ctx, key, []byte(value))
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get and delete the key
	retrieved, err := c.GetAndDelete(ctx, key)
	if err != nil {
		t.Fatalf("GetAndDelete failed: %v", err)
	}

	if string(retrieved) != value {
		t.Errorf("Expected %s, got %s", value, retrieved)
	}

	// Verify the key is gone
	_, err = c.Get(ctx, key)
	if err != cache.ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound after GetAndDelete, got %v", err)
	}
}

func TestMemoryCache_GetAndDeleteNonExistent(t *testing.T) {
	c := cache.NewMemory(0)

	ctx := context.Background()
	key := "non-existent-key"

	// GetAndDelete non-existent key should return ErrKeyNotFound
	_, err := c.GetAndDelete(ctx, key)
	if err != cache.ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}
}

func TestMemoryCache_Drain(t *testing.T) {
	c := cache.NewMemory(0)

	ctx := context.Background()
	items := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	// Set multiple values
	for key, value := range items {
		err := c.Set(ctx, key, []byte(value))
		if err != nil {
			t.Fatalf("Set failed for %s: %v", key, err)
		}
	}

	// Drain the cache
	drained, err := c.Drain(ctx)
	if err != nil {
		t.Fatalf("Drain failed: %v", err)
	}

	// Verify all items are drained
	if len(drained) != len(items) {
		t.Errorf("Expected %d items, got %d", len(items), len(drained))
	}

	for key, expectedValue := range items {
		actualValue, ok := drained[key]
		if !ok {
			t.Errorf("Expected key %s in drained items", key)
			continue
		}
		if string(actualValue) != expectedValue {
			t.Errorf("Expected %s, got %s for key %s", expectedValue, actualValue, key)
		}
	}

	// Verify cache is now empty
	for key := range items {
		_, err := c.Get(ctx, key)
		if err != cache.ErrKeyNotFound {
			t.Errorf("Expected ErrKeyNotFound for key %s after drain, got %v", key, err)
		}
	}
}

func TestMemoryCache_DrainEmpty(t *testing.T) {
	cache := cache.NewMemory(0)

	ctx := context.Background()

	// Drain empty cache
	drained, err := cache.Drain(ctx)
	if err != nil {
		t.Fatalf("Drain failed: %v", err)
	}

	if len(drained) != 0 {
		t.Errorf("Expected 0 items from empty cache, got %d", len(drained))
	}
}

func TestMemoryCache_Cleanup(t *testing.T) {
	c := cache.NewMemory(0) // No default TTL

	ctx := context.Background()
	key := "test-key"
	value := "test-value"
	shortTTL := 100 * time.Millisecond

	// Set a value with short TTL
	err := c.Set(ctx, key, []byte(value), cache.WithTTL(shortTTL))
	if err != nil {
		t.Fatalf("Set with TTL failed: %v", err)
	}

	// Verify the value is there initially
	_, err = c.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Wait for the item to expire
	time.Sleep(2 * shortTTL)

	// Run cleanup
	err = c.Cleanup(ctx)
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Verify the expired item is gone
	_, err = c.Get(ctx, key)
	if err != cache.ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound after cleanup, got %v", err)
	}
}

func TestMemoryCache_CleanupNoExpired(t *testing.T) {
	cache := cache.NewMemory(0)

	ctx := context.Background()
	key := "test-key"
	value := "test-value"

	// Set a value without TTL
	err := cache.Set(ctx, key, []byte(value))
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Run cleanup on cache with no expired items
	err = cache.Cleanup(ctx)
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Verify the value is still there
	retrieved, err := cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(retrieved) != value {
		t.Errorf("Expected %s, got %s", value, retrieved)
	}
}

func TestMemoryCache_Overwrite(t *testing.T) {
	cache := cache.NewMemory(0)

	ctx := context.Background()
	key := "test-key"
	value1 := "value1"
	value2 := "value2"

	// Set initial value
	err := cache.Set(ctx, key, []byte(value1))
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Overwrite with new value
	err = cache.Set(ctx, key, []byte(value2))
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Verify the new value is there
	retrieved, err := cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(retrieved) != value2 {
		t.Errorf("Expected %s, got %s", value2, retrieved)
	}
}

func TestMemoryCache_EmptyValue(t *testing.T) {
	cache := cache.NewMemory(0)

	ctx := context.Background()
	key := "test-key"
	value := ""

	// Set empty value
	err := cache.Set(ctx, key, []byte(value))
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get the empty value
	retrieved, err := cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(retrieved) != value {
		t.Errorf("Expected empty string, got %s", retrieved)
	}
}

func TestMemoryCache_SpecialCharacters(t *testing.T) {
	cache := cache.NewMemory(0)

	ctx := context.Background()
	key := "test:key/with@special#chars"
	value := "value with special chars: !@#$%^&*()"

	// Set value with special characters
	err := cache.Set(ctx, key, []byte(value))
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get the value
	retrieved, err := cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(retrieved) != value {
		t.Errorf("Expected %s, got %s", value, retrieved)
	}
}

func TestMemoryCache_LargeValue(t *testing.T) {
	cache := cache.NewMemory(0)

	ctx := context.Background()
	key := "large-key"
	value := string(make([]byte, 1024*1024)) // 1MB value

	// Set large value
	err := cache.Set(ctx, key, []byte(value))
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get the large value
	retrieved, err := cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(retrieved) != value {
		t.Errorf("Large value mismatch")
	}
}
