//nolint:errcheck
package cache_test

import (
	"context"
	"runtime"
	"strconv"
	"testing"

	"github.com/android-sms-gateway/server/pkg/cache"
)

func TestMemoryCache_MemoryAllocationPattern(t *testing.T) {
	// This test analyzes memory allocation patterns during cache operations
	cache := cache.NewMemory(0)
	ctx := context.Background()

	// Force GC and get baseline memory
	runtime.GC()
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Perform cache operations that trigger allocations
	const numItems = 1000
	for i := range numItems {
		key := "profile-key-" + strconv.Itoa(i)
		value := "profile-value-" + strconv.Itoa(i)

		err := cache.Set(ctx, key, value)
		if err != nil {
			t.Errorf("Set failed: %v", err)
		}

		// Get the value to trigger read path
		_, err = cache.Get(ctx, key)
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}
	}

	// Force GC again and measure memory
	runtime.GC()
	runtime.ReadMemStats(&m2)

	// Calculate memory growth
	allocDiff := m2.TotalAlloc - m1.TotalAlloc
	allocPerItem := float64(allocDiff) / float64(numItems)

	t.Logf("Memory allocation stats:")
	t.Logf("  Total allocated: %d bytes", m2.TotalAlloc)
	t.Logf("  Allocation difference: %d bytes", allocDiff)
	t.Logf("  Allocations per item: %.2f bytes", allocPerItem)
	t.Logf("  Heap objects: %d", m2.HeapObjects)
	t.Logf("  GC cycles: %d", m2.NumGC)

	// Reasonable bounds for memory allocation (these are approximate)
	// Higher threshold due to both Set and Get operations
	if allocPerItem > 300 {
		t.Errorf("Expected less than 300 bytes per item, got %.2f bytes", allocPerItem)
	}
}

func TestMemoryCache_MemoryCleanup(t *testing.T) {
	// This test verifies that memory is properly cleaned up after cache operations
	cache := cache.NewMemory(0)
	ctx := context.Background()

	// Force GC and get baseline memory
	runtime.GC()
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Add many items to cache
	const numItems = 5000
	for i := range numItems {
		key := "cleanup-key-" + strconv.Itoa(i)
		value := "cleanup-value-" + strconv.Itoa(i)

		err := cache.Set(ctx, key, value)
		if err != nil {
			t.Errorf("Set failed: %v", err)
		}
	}

	// Drain the cache to clear all items
	_, err := cache.Drain(ctx)
	if err != nil {
		t.Errorf("Drain failed: %v", err)
	}

	// Force GC and measure memory after cleanup
	runtime.GC()
	runtime.ReadMemStats(&m2)

	// Calculate memory reduction
	allocDiff := m2.TotalAlloc - m1.TotalAlloc

	t.Logf("Memory cleanup stats:")
	t.Logf("  Total allocated: %d bytes", m2.TotalAlloc)
	t.Logf("  Allocation difference: %d bytes", allocDiff)
	t.Logf("  Heap objects: %d", m2.HeapObjects)
	t.Logf("  GC cycles: %d", m2.NumGC)

	// Memory should not grow significantly after cleanup
	// Allow some growth for overhead, but it should be reasonable
	if allocDiff > 2*1024*1024 { // 2MB
		t.Errorf("Expected less than 2MB memory growth after cleanup, got %d bytes", allocDiff)
	}
}

func TestMemoryCache_MemoryPressure(t *testing.T) {
	// This test simulates memory pressure scenarios
	ctx := context.Background()

	// Force GC and get baseline memory
	runtime.GC()
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Simulate memory pressure by creating and destroying many cache instances
	const numCaches = 100
	const itemsPerCache = 50

	for i := 0; i < numCaches; i++ {
		// Create a new cache
		tempCache := cache.NewMemory(0)

		// Add items to cache
		for j := 0; j < itemsPerCache; j++ {
			key := "pressure-key-" + strconv.Itoa(i) + "-" + strconv.Itoa(j)
			value := "pressure-value-" + strconv.Itoa(i) + "-" + strconv.Itoa(j)

			err := tempCache.Set(ctx, key, value)
			if err != nil {
				t.Errorf("Set failed: %v", err)
			}
		}

		// Drain the cache
		_, err := tempCache.Drain(ctx)
		if err != nil {
			t.Errorf("Drain failed: %v", err)
		}
	}

	// Force GC and measure memory after pressure test
	runtime.GC()
	runtime.ReadMemStats(&m2)

	// Calculate memory growth
	allocDiff := m2.TotalAlloc - m1.TotalAlloc

	t.Logf("Memory pressure stats:")
	t.Logf("  Total allocated: %d bytes", m2.TotalAlloc)
	t.Logf("  Allocation difference: %d bytes", allocDiff)
	t.Logf("  Heap objects: %d", m2.HeapObjects)
	t.Logf("  GC cycles: %d", m2.NumGC)

	// Memory growth should be reasonable even under pressure
	// Allow some growth for overhead, but it should be proportional
	// Higher threshold due to cache creation/destruction overhead
	expectedMaxGrowth := uint64(numCaches * itemsPerCache * 300) // 300 bytes per item estimate
	if allocDiff > expectedMaxGrowth {
		t.Errorf("Expected less than %d bytes memory growth under pressure, got %d bytes", expectedMaxGrowth, allocDiff)
	}
}

func TestMemoryCache_GCStress(t *testing.T) {
	// This test verifies cache behavior under frequent GC cycles
	c := cache.NewMemory(0)
	ctx := context.Background()

	// Add items to cache
	const numItems = 1000
	for i := range numItems {
		key := "gc-key-" + strconv.Itoa(i)
		value := "gc-value-" + strconv.Itoa(i)

		err := c.Set(ctx, key, value)
		if err != nil {
			t.Errorf("Set failed: %v", err)
		}
	}

	// Perform frequent GC operations and verify cache still works
	const numGCs = 10
	for range numGCs {
		// Force GC
		runtime.GC()

		// Verify cache operations still work
		for j := range 100 {
			key := "gc-key-" + strconv.Itoa(j)
			_, err := c.Get(ctx, key)
			if err != nil {
				t.Errorf("Get failed during GC stress test: %v", err)
			}
		}
	}

	// Verify all items are still accessible
	for i := range numItems {
		key := "gc-key-" + strconv.Itoa(i)
		value := "gc-value-" + strconv.Itoa(i)

		retrieved, err := c.Get(ctx, key)
		if err != nil {
			t.Errorf("Get failed after GC stress: %v", err)
		}

		if retrieved != value {
			t.Errorf("Value mismatch after GC stress: expected %s, got %s", value, retrieved)
		}
	}
}

func TestMemoryCache_MemoryLeakDetection(t *testing.T) {
	// This test helps detect memory leaks by creating and destroying many caches
	ctx := context.Background()

	// Force GC and get baseline memory
	runtime.GC()
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Create and destroy many cache instances
	const numCaches = 1000
	for i := range numCaches {
		// Create a new cache
		tempCache := cache.NewMemory(0)

		// Add some items
		for j := range 10 {
			key := "leak-key-" + strconv.Itoa(i) + "-" + strconv.Itoa(j)
			value := "leak-value-" + strconv.Itoa(i) + "-" + strconv.Itoa(j)

			err := tempCache.Set(ctx, key, value)
			if err != nil {
				t.Errorf("Set failed: %v", err)
			}
		}

		// Clear the cache
		tempCache.Drain(ctx)

		// Help GC by clearing reference
		tempCache = nil
	}

	// Force GC and measure memory
	runtime.GC()
	runtime.ReadMemStats(&m2)

	// Calculate memory growth
	// Convert to int64 to avoid unsigned wrap-around when memory decreases
	heapDiff := int64(m2.HeapAlloc) - int64(m1.HeapAlloc)

	t.Logf("Memory leak detection stats:")
	t.Logf("  Initial heap: %d bytes", m1.HeapAlloc)
	t.Logf("  Final heap: %d bytes", m2.HeapAlloc)
	t.Logf("  Heap delta: %d bytes", heapDiff)
	t.Logf("  Heap objects: %d", m2.HeapObjects)
	t.Logf("  GC cycles: %d", m2.NumGC)

	// Only report as leak if memory increased beyond threshold
	if heapDiff > 1*1024*1024 { // 1MB threshold for leak detection
		t.Errorf("Potential memory leak detected: %d bytes retained after cleanup", heapDiff)
	} else if heapDiff < 0 {
		t.Logf("Memory reduced by %d bytes after cleanup", -heapDiff)
	}
}

func BenchmarkMemoryCache_MemoryUsage(b *testing.B) {
	// This benchmark tracks memory usage patterns
	cache := cache.NewMemory(0)
	ctx := context.Background()

	b.ReportAllocs()
	runtime.GC()
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "bench-key-" + strconv.Itoa(i)
			value := "bench-value-" + strconv.Itoa(i)

			// Set and get
			cache.Set(ctx, key, value)
			cache.Get(ctx, key)

			// Delete
			cache.Delete(ctx, key)

			i++
		}
	})
	runtime.ReadMemStats(&m2)
	b.Logf("TotalAlloc per op: %.2f bytes/op", float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N))
}
