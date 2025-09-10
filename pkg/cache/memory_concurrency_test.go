package cache_test

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/android-sms-gateway/server/pkg/cache"
)

func TestMemoryCache_ConcurrentReads(t *testing.T) {
	cache := cache.NewMemory(0)

	ctx := context.Background()
	key := "test-key"
	value := "test-value"

	// Set initial value
	err := cache.Set(ctx, key, value)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	const numGoroutines = 100
	var wg sync.WaitGroup

	// Launch multiple concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			retrieved, err := cache.Get(ctx, key)
			if err != nil {
				t.Errorf("Get failed: %v", err)
				return
			}

			if retrieved != value {
				t.Errorf("Expected %s, got %s", value, retrieved)
			}
		}()
	}

	wg.Wait()
}

func TestMemoryCache_ConcurrentWrites(t *testing.T) {
	cache := cache.NewMemory(0)

	ctx := context.Background()
	const numKeys = 100
	const numGoroutines = 10
	var wg sync.WaitGroup

	// Launch multiple concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < numKeys/numGoroutines; j++ {
				key := "key-" + strconv.Itoa(goroutineID) + "-" + strconv.Itoa(j)
				value := "value-" + strconv.Itoa(goroutineID) + "-" + strconv.Itoa(j)

				err := cache.Set(ctx, key, value)
				if err != nil {
					t.Errorf("Set failed for key %s: %v", key, err)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify all keys were set correctly
	for i := 0; i < numGoroutines; i++ {
		for j := 0; j < numKeys/numGoroutines; j++ {
			key := "key-" + strconv.Itoa(i) + "-" + strconv.Itoa(j)
			expectedValue := "value-" + strconv.Itoa(i) + "-" + strconv.Itoa(j)

			retrieved, err := cache.Get(ctx, key)
			if err != nil {
				t.Errorf("Get failed for key %s: %v", key, err)
				continue
			}

			if retrieved != expectedValue {
				t.Errorf("Expected %s, got %s for key %s", expectedValue, retrieved, key)
			}
		}
	}
}

func TestMemoryCache_ConcurrentReadWrite(t *testing.T) {
	c := cache.NewMemory(0)

	ctx := context.Background()
	const numOperations = 1000
	const numReaders = 8
	const numWriters = 2
	var wg sync.WaitGroup

	// Launch concurrent readers
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < numOperations/numReaders; j++ {
				key := "shared-key"
				_, err := c.Get(ctx, key)
				if err != nil && err != cache.ErrKeyNotFound {
					t.Errorf("Get failed: %v", err)
				}
			}
		}()
	}

	// Launch concurrent writers
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < numOperations/numWriters; j++ {
				key := "shared-key"
				value := "value-" + strconv.Itoa(j)

				err := c.Set(ctx, key, value)
				if err != nil {
					t.Errorf("Set failed: %v", err)
				}
			}
		}()
	}

	wg.Wait()
}

func TestMemoryCache_ConcurrentSetAndGetAndDelete(t *testing.T) {
	cache := cache.NewMemory(0)

	ctx := context.Background()
	const numOperations = 500
	const numGoroutines = 10
	var wg sync.WaitGroup

	// Launch goroutines that perform Set, Get, and Delete operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < numOperations/numGoroutines; j++ {
				key := "key-" + strconv.Itoa(goroutineID) + "-" + strconv.Itoa(j)
				value := "value-" + strconv.Itoa(goroutineID) + "-" + strconv.Itoa(j)

				// Set
				err := cache.Set(ctx, key, value)
				if err != nil {
					t.Errorf("Set failed: %v", err)
					continue
				}

				// Get
				retrieved, err := cache.Get(ctx, key)
				if err != nil {
					t.Errorf("Get failed: %v", err)
					continue
				}

				if retrieved != value {
					t.Errorf("Expected %s, got %s", value, retrieved)
				}

				// Delete
				err = cache.Delete(ctx, key)
				if err != nil {
					t.Errorf("Delete failed: %v", err)
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestMemoryCache_ConcurrentSetOrFail(t *testing.T) {
	c := cache.NewMemory(0)

	ctx := context.Background()
	const numGoroutines = 10
	const attemptsPerGoroutine = 100
	var wg sync.WaitGroup

	// Launch goroutines that try to SetOrFail the same key
	key := "contentious-key"
	value := "initial-value"

	var successCount atomic.Int32
	var existsCount atomic.Int32

	for range numGoroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for range attemptsPerGoroutine {
				err := c.SetOrFail(ctx, key, value)
				switch err {
				case nil:
					successCount.Add(1)
				case cache.ErrKeyExists:
					existsCount.Add(1)
				default:
					t.Errorf("SetOrFail failed: %v", err)
				}
			}
		}()
	}

	wg.Wait()

	// Only one goroutine should succeed, all others should get ErrKeyExists
	if c := successCount.Load(); c != 1 {
		t.Errorf("Expected 1 successful SetOrFail, got %d", c)
	}

	expectedExistsCount := (numGoroutines * attemptsPerGoroutine) - 1
	if c := int(existsCount.Load()); c != expectedExistsCount {
		t.Errorf("Expected %d ErrKeyExists, got %d", expectedExistsCount, c)
	}
}

func TestMemoryCache_ConcurrentDrain(t *testing.T) {
	c := cache.NewMemory(0)

	ctx := context.Background()
	const numItems = 100
	const numGoroutines = 5
	var wg sync.WaitGroup

	// Pre-populate cache with items
	for i := 0; i < numItems; i++ {
		key := "item-" + strconv.Itoa(i)
		value := "value-" + strconv.Itoa(i)

		err := c.Set(ctx, key, value)
		if err != nil {
			t.Fatalf("Set failed for item %d: %v", i, err)
		}
	}

	// Launch concurrent drain operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			_, err := c.Drain(ctx)
			if err != nil {
				t.Errorf("Drain failed: %v", err)
			}
		}()
	}

	wg.Wait()

	// Cache should be empty after all drain operations
	for i := 0; i < numItems; i++ {
		key := "item-" + strconv.Itoa(i)
		_, err := c.Get(ctx, key)
		if err != cache.ErrKeyNotFound {
			t.Errorf("Expected ErrKeyNotFound for key %s after drain, got %v", key, err)
		}
	}
}

func TestMemoryCache_ConcurrentCleanup(t *testing.T) {
	c := cache.NewMemory(0)

	ctx := context.Background()
	const numItems = 50
	const numGoroutines = 5
	var wg sync.WaitGroup

	// Pre-populate cache with items that will expire quickly
	for i := 0; i < numItems; i++ {
		key := "item-" + strconv.Itoa(i)
		value := "value-" + strconv.Itoa(i)

		err := c.Set(ctx, key, value, cache.WithTTL(10*time.Millisecond))
		if err != nil {
			t.Fatalf("Set failed for item %d: %v", i, err)
		}
	}

	// Launch concurrent cleanup operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			err := c.Cleanup(ctx)
			if err != nil {
				t.Errorf("Cleanup failed: %v", err)
			}
		}()
	}

	wg.Wait()

	// Wait for items to actually expire
	time.Sleep(50 * time.Millisecond)

	// Run one final cleanup to ensure all expired items are removed
	err := c.Cleanup(ctx)
	if err != nil {
		t.Errorf("Final cleanup failed: %v", err)
	}

	// All items should be expired and removed
	for i := 0; i < numItems; i++ {
		key := "item-" + strconv.Itoa(i)
		_, err := c.Get(ctx, key)
		if err != cache.ErrKeyExpired && err != cache.ErrKeyNotFound {
			t.Errorf("Expected ErrKeyExpired or ErrKeyNotFound for key %s, got %v", key, err)
		}
	}
}

func TestMemoryCache_ConcurrentGetAndDelete(t *testing.T) {
	c := cache.NewMemory(0)

	ctx := context.Background()
	const numGoroutines = 10
	const attemptsPerGoroutine = 50
	var wg sync.WaitGroup

	// Pre-populate cache with items
	for i := 0; i < numGoroutines*attemptsPerGoroutine; i++ {
		key := "item-" + strconv.Itoa(i)
		value := "value-" + strconv.Itoa(i)

		err := c.Set(ctx, key, value)
		if err != nil {
			t.Fatalf("Set failed for item %d: %v", i, err)
		}
	}

	// Launch goroutines that perform GetAndDelete operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < attemptsPerGoroutine; j++ {
				key := "item-" + strconv.Itoa(goroutineID*attemptsPerGoroutine+j)

				_, err := c.GetAndDelete(ctx, key)
				if err != nil && err != cache.ErrKeyNotFound {
					t.Errorf("GetAndDelete failed: %v", err)
				}
			}
		}(i)
	}

	wg.Wait()

	// All items should be deleted
	for i := 0; i < numGoroutines*attemptsPerGoroutine; i++ {
		key := "item-" + strconv.Itoa(i)
		_, err := c.Get(ctx, key)
		if err != cache.ErrKeyNotFound {
			t.Errorf("Expected ErrKeyNotFound for key %s after GetAndDelete, got %v", key, err)
		}
	}
}

func TestMemoryCache_RaceConditionDetection(t *testing.T) {
	// This test is specifically designed to detect race conditions
	// by running many operations concurrently with the race detector enabled

	cache := cache.NewMemory(0)

	ctx := context.Background()
	const duration = 2 * time.Second
	const numGoroutines = 20
	var wg sync.WaitGroup

	start := time.Now()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for time.Since(start) < duration {
				key := "race-key-" + strconv.Itoa(goroutineID)
				value := "race-value-" + strconv.Itoa(goroutineID) + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)

				// Randomly choose operation
				switch time.Now().UnixNano() % 4 {
				case 0:
					cache.Set(ctx, key, value)
				case 1:
					cache.Get(ctx, key)
				case 2:
					cache.Delete(ctx, key)
				case 3:
					cache.GetAndDelete(ctx, key)
				}
			}
		}(i)
	}

	wg.Wait()
}
