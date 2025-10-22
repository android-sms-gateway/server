//nolint:errcheck
package cache_test

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/android-sms-gateway/server/pkg/cache"
)

// BenchmarkMemoryCache_Set measures the performance of Set operations
func BenchmarkMemoryCache_Set(b *testing.B) {
	cache := cache.NewMemory(0)
	ctx := context.Background()
	key := "benchmark-key"
	value := []byte("benchmark-value")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Set(ctx, key, value)
		}
	})
}

// BenchmarkMemoryCache_Get measures the performance of Get operations
func BenchmarkMemoryCache_Get(b *testing.B) {
	cache := cache.NewMemory(0)
	ctx := context.Background()
	key := "benchmark-key"
	value := "benchmark-value"

	// Pre-populate the cache
	cache.Set(ctx, key, []byte(value))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Get(ctx, key)
		}
	})
}

// BenchmarkMemoryCache_SetAndGet measures the performance of Set followed by Get
func BenchmarkMemoryCache_SetAndGet(b *testing.B) {
	cache := cache.NewMemory(0)
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "key-" + strconv.Itoa(i)
			value := "value-" + strconv.Itoa(i)
			i++

			cache.Set(ctx, key, []byte(value))
			cache.Get(ctx, key)
		}
	})
}

// BenchmarkMemoryCache_SetOrFail measures the performance of SetOrFail operations
func BenchmarkMemoryCache_SetOrFail(b *testing.B) {
	cache := cache.NewMemory(0)
	ctx := context.Background()
	key := "benchmark-key"
	value := []byte("benchmark-value")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.SetOrFail(ctx, key, value)
		}
	})
}

// BenchmarkMemoryCache_GetAndDelete measures the performance of GetAndDelete operations
func BenchmarkMemoryCache_GetAndDelete(b *testing.B) {
	cache := cache.NewMemory(0)
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "key-" + strconv.Itoa(i)
			value := "value-" + strconv.Itoa(i)
			i++

			cache.Set(ctx, key, []byte(value))
			cache.GetAndDelete(ctx, key)
		}
	})
}

// BenchmarkMemoryCache_Delete measures the performance of Delete operations
func BenchmarkMemoryCache_Delete(b *testing.B) {
	cache := cache.NewMemory(0)
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "key-" + strconv.Itoa(i)
			value := "value-" + strconv.Itoa(i)
			i++

			cache.Set(ctx, key, []byte(value))
			cache.Delete(ctx, key)
		}
	})
}

// BenchmarkMemoryCache_Cleanup measures the performance of Cleanup operations
func BenchmarkMemoryCache_Cleanup(b *testing.B) {
	cache := cache.NewMemory(0)
	ctx := context.Background()

	// Pre-populate cache with many items
	for i := 0; i < 1000; i++ {
		key := "item-" + strconv.Itoa(i)
		value := "value-" + strconv.Itoa(i)
		cache.Set(ctx, key, []byte(value))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Cleanup(ctx)
		}
	})
}

// BenchmarkMemoryCache_Drain measures the performance of Drain operations
func BenchmarkMemoryCache_Drain(b *testing.B) {
	cache := cache.NewMemory(0)
	ctx := context.Background()

	// Pre-populate cache with many items
	for i := 0; i < 1000; i++ {
		key := "item-" + strconv.Itoa(i)
		value := "value-" + strconv.Itoa(i)
		cache.Set(ctx, key, []byte(value))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Drain(ctx)
		}
	})
}

// BenchmarkMemoryCache_ConcurrentReads measures performance with different numbers of concurrent readers
func BenchmarkMemoryCache_ConcurrentReads(b *testing.B) {
	cache := cache.NewMemory(0)
	ctx := context.Background()
	key := "benchmark-key"
	value := "benchmark-value"

	// Pre-populate the cache
	cache.Set(ctx, key, []byte(value))

	benchmarks := []struct {
		name       string
		goroutines int
	}{
		{"1 Reader", 1},
		{"4 Readers", 4},
		{"16 Readers", 16},
		{"64 Readers", 64},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.SetParallelism(bm.goroutines)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					cache.Get(ctx, key)
				}
			})
		})
	}
}

// BenchmarkMemoryCache_ConcurrentWrites measures performance with different numbers of concurrent writers
func BenchmarkMemoryCache_ConcurrentWrites(b *testing.B) {
	cache := cache.NewMemory(0)
	ctx := context.Background()

	benchmarks := []struct {
		name       string
		goroutines int
	}{
		{"1 Writer", 1},
		{"4 Writers", 4},
		{"16 Writers", 16},
		{"64 Writers", 64},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.SetParallelism(bm.goroutines)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					key := "key-" + strconv.Itoa(i)
					value := "value-" + strconv.Itoa(i)
					i++

					cache.Set(ctx, key, []byte(value))
				}
			})
		})
	}
}

// BenchmarkMemoryCache_MixedWorkload measures performance with mixed read/write operations
func BenchmarkMemoryCache_MixedWorkload(b *testing.B) {
	cache := cache.NewMemory(0)
	ctx := context.Background()

	benchmarks := []struct {
		name       string
		readRatio  float64
		goroutines int
	}{
		{"Read-Heavy 90/10", 0.9, 16},
		{"Balanced 50/50", 0.5, 16},
		{"Write-Heavy 10/90", 0.1, 16},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.SetParallelism(bm.goroutines)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				r := rand.New(rand.NewSource(time.Now().UnixNano()))
				i := 0

				for pb.Next() {
					if r.Float64() < bm.readRatio {
						// Read operation
						key := "key-" + strconv.Itoa(i%100) // Reuse keys to simulate working set
						cache.Get(ctx, key)
					} else {
						// Write operation
						key := "key-" + strconv.Itoa(i%100)
						value := "value-" + strconv.Itoa(i)
						i++

						cache.Set(ctx, key, []byte(value))
					}
				}
			})
		})
	}
}

// BenchmarkMemoryCache_Scaling measures how performance scales with increasing load
func BenchmarkMemoryCache_Scaling(b *testing.B) {
	cache := cache.NewMemory(0)
	ctx := context.Background()

	benchmarks := []struct {
		name                   string
		operationsPerGoroutine int
		goroutines             int
	}{
		{"Small Load", 10, 1},
		{"Medium Load", 100, 10},
		{"Large Load", 1000, 100},
		{"Very Large Load", 10000, 1000},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// Pre-populate cache
			for i := 0; i < bm.operationsPerGoroutine*bm.goroutines; i++ {
				key := "key-" + strconv.Itoa(i)
				value := "value-" + strconv.Itoa(i)
				cache.Set(ctx, key, []byte(value))
			}

			b.SetParallelism(bm.goroutines)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				localI := 0
				for pb.Next() {
					// Simulate random access
					key := "key-" + strconv.Itoa(localI%(bm.operationsPerGoroutine*bm.goroutines))
					cache.Get(ctx, key)
					localI++
				}
			})
		})
	}
}

// BenchmarkMemoryCache_TTLOverhead measures the performance impact of TTL operations
func BenchmarkMemoryCache_TTLOverhead(b *testing.B) {
	c := cache.NewMemory(0)
	ctx := context.Background()
	key := "benchmark-key"
	value := []byte("benchmark-value")
	ttl := time.Hour

	benchmarks := []struct {
		name    string
		withTTL bool
	}{
		{"Without TTL", false},
		{"With TTL", true},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					if bm.withTTL {
						c.Set(ctx, key, value, cache.WithTTL(ttl))
					} else {
						c.Set(ctx, key, value)
					}
				}
			})
		})
	}
}

// BenchmarkMemoryCache_LargeValues measures performance with large values
func BenchmarkMemoryCache_LargeValues(b *testing.B) {
	cache := cache.NewMemory(0)
	ctx := context.Background()
	key := "benchmark-key"

	sizes := []struct {
		name string
		size int
	}{
		{"1KB", 1 * 1024},
		{"10KB", 10 * 1024},
		{"100KB", 100 * 1024},
		{"1MB", 1024 * 1024},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			value := make([]byte, size.size)
			for i := range value {
				value[i] = byte(i % 256)
			}

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					if err := cache.Set(ctx, key, value); err != nil {
						b.Fatal(err)
					}
					if _, err := cache.Get(ctx, key); err != nil {
						b.Fatal(err)
					}
				}
			})
		})
	}
}

// BenchmarkMemoryCache_MemoryGrowth measures memory allocation patterns
func BenchmarkMemoryCache_MemoryGrowth(b *testing.B) {
	cache := cache.NewMemory(0)
	ctx := context.Background()
	b.ReportAllocs()

	sizes := []int{100, 1000, 10000, 100000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("%d_items", size), func(b *testing.B) {
			b.ResetTimer()

			for b.Loop() {
				// Clear cache
				cache.Drain(ctx)

				// Add new items
				for j := range size {
					key := "key-" + strconv.Itoa(j)
					value := "value-" + strconv.Itoa(j)
					cache.Set(ctx, key, []byte(value))
				}
			}
		})
	}
}

// BenchmarkMemoryCache_RandomAccess measures performance with random key access patterns
func BenchmarkMemoryCache_RandomAccess(b *testing.B) {
	cache := cache.NewMemory(0)
	ctx := context.Background()
	const numKeys = 1000

	// Pre-populate cache with many keys
	for i := 0; i < numKeys; i++ {
		key := "key-" + strconv.Itoa(i)
		value := "value-" + strconv.Itoa(i)
		cache.Set(ctx, key, []byte(value))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))

		for pb.Next() {
			key := "key-" + strconv.Itoa(r.Intn(numKeys))
			cache.Get(ctx, key)
		}
	})
}

// BenchmarkMemoryCache_HotKey measures performance with a frequently accessed key
func BenchmarkMemoryCache_HotKey(b *testing.B) {
	cache := cache.NewMemory(0)
	ctx := context.Background()
	hotKey := "hot-key"
	value := "hot-value"

	// Pre-populate the hot key
	cache.Set(ctx, hotKey, []byte(value))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Get(ctx, hotKey)
		}
	})
}

// BenchmarkMemoryCache_ColdKey measures performance with rarely accessed keys
func BenchmarkMemoryCache_ColdKey(b *testing.B) {
	cache := cache.NewMemory(0)
	ctx := context.Background()
	const numKeys = 10000

	// Pre-populate cache with many keys
	for i := 0; i < numKeys; i++ {
		key := "key-" + strconv.Itoa(i)
		value := "value-" + strconv.Itoa(i)
		cache.Set(ctx, key, []byte(value))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))

		for pb.Next() {
			key := "key-" + strconv.Itoa(r.Intn(numKeys))
			cache.Get(ctx, key)
		}
	})
}
