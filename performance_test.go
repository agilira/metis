// performance_test.go: Performance tests for Metis strategic caching library
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package metis

import (
	"fmt"
	"testing"
	"time"
)

// BenchmarkWTinyLFU_Performance tests W-TinyLFU performance
func BenchmarkWTinyLFU_Performance(b *testing.B) {
	config := CacheConfig{
		EnableCaching:        true,
		CacheSize:            10000,
		ShardCount:           32,
		EnableCompression:    true,
		AdmissionProbability: 0.5,
		EvictionPolicy:       "wtinylfu",
		TTL:                  10 * time.Minute,
		CleanupInterval:      2 * time.Minute,
	}

	cache := NewStrategicCache(config)
	defer cache.Close()

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key_%d", i)
		cache.Set(key, fmt.Sprintf("value_%d", i))
	}

	b.ResetTimer()
	b.Run("Set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("set_key_%d", i)
			cache.Set(key, fmt.Sprintf("set_value_%d", i))
		}
	})

	b.Run("Get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key_%d", i%1000)
			_, _ = cache.Get(key)
		}
	})

	b.Run("Mixed", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if i%3 == 0 {
				// Set operation
				key := fmt.Sprintf("mixed_key_%d", i)
				cache.Set(key, fmt.Sprintf("mixed_value_%d", i))
			} else {
				// Get operation
				key := fmt.Sprintf("key_%d", i%1000)
				_, _ = cache.Get(key)
			}
		}
	})
}

// BenchmarkLRU_Performance tests LRU performance
func BenchmarkLRU_Performance(b *testing.B) {
	config := CacheConfig{
		EnableCaching:        true,
		CacheSize:            10000,
		ShardCount:           32,
		EnableCompression:    true,
		AdmissionProbability: 0.5,
		EvictionPolicy:       "lru",
		TTL:                  10 * time.Minute,
		CleanupInterval:      2 * time.Minute,
	}

	cache := NewStrategicCache(config)
	defer cache.Close()

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key_%d", i)
		cache.Set(key, fmt.Sprintf("value_%d", i))
	}

	b.ResetTimer()
	b.Run("Set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("set_key_%d", i)
			cache.Set(key, fmt.Sprintf("set_value_%d", i))
		}
	})

	b.Run("Get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key_%d", i%1000)
			_, _ = cache.Get(key)
		}
	})

	b.Run("Mixed", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if i%3 == 0 {
				// Set operation
				key := fmt.Sprintf("mixed_key_%d", i)
				cache.Set(key, fmt.Sprintf("mixed_value_%d", i))
			} else {
				// Get operation
				key := fmt.Sprintf("key_%d", i%1000)
				_, _ = cache.Get(key)
			}
		}
	})
}

// BenchmarkConcurrent_Performance tests concurrent performance
func BenchmarkConcurrent_Performance(b *testing.B) {
	config := CacheConfig{
		EnableCaching:        true,
		CacheSize:            10000,
		ShardCount:           64, // More shards for concurrency
		EnableCompression:    true,
		AdmissionProbability: 0.5,
		EvictionPolicy:       "wtinylfu",
		TTL:                  10 * time.Minute,
		CleanupInterval:      2 * time.Minute,
	}

	cache := NewStrategicCache(config)
	defer cache.Close()

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key_%d", i)
		cache.Set(key, fmt.Sprintf("value_%d", i))
	}

	b.ResetTimer()
	b.Run("Concurrent", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				if i%3 == 0 {
					// Set operation
					key := fmt.Sprintf("concurrent_key_%d", i)
					cache.Set(key, fmt.Sprintf("concurrent_value_%d", i))
				} else {
					// Get operation
					key := fmt.Sprintf("key_%d", i%1000)
					_, _ = cache.Get(key)
				}
				i++
			}
		})
	})
}

// TestPerformanceComparison compares different configurations
func TestPerformanceComparison(t *testing.T) {
	configs := []struct {
		name   string
		config CacheConfig
	}{
		{
			name: "W-TinyLFU High Performance",
			config: CacheConfig{
				EnableCaching:        true,
				CacheSize:            50000,
				ShardCount:           64,
				EnableCompression:    true,
				AdmissionProbability: 0.5,
				EvictionPolicy:       "wtinylfu",
				TTL:                  15 * time.Minute,
				CleanupInterval:      2 * time.Minute,
			},
		},
		{
			name: "LRU Standard",
			config: CacheConfig{
				EnableCaching:        true,
				CacheSize:            10000,
				ShardCount:           32,
				EnableCompression:    true,
				AdmissionProbability: 0.5,
				EvictionPolicy:       "lru",
				TTL:                  10 * time.Minute,
				CleanupInterval:      2 * time.Minute,
			},
		},
		{
			name: "Memory Constrained",
			config: CacheConfig{
				EnableCaching:        true,
				CacheSize:            5000,
				ShardCount:           16,
				EnableCompression:    true,
				AdmissionProbability: 0.3,
				EvictionPolicy:       "wtinylfu",
				TTL:                  5 * time.Minute,
				CleanupInterval:      1 * time.Minute,
			},
		},
	}

	for _, tc := range configs {
		t.Run(tc.name, func(t *testing.T) {
			cache := NewStrategicCache(tc.config)
			defer cache.Close()

			// Performance test
			start := time.Now()
			operations := 10000

			// Pre-populate
			for i := 0; i < 1000; i++ {
				key := fmt.Sprintf("key_%d", i)
				cache.Set(key, fmt.Sprintf("value_%d", i))
			}

			// Mixed operations
			for i := 0; i < operations; i++ {
				if i%3 == 0 {
					key := fmt.Sprintf("perf_key_%d", i)
					cache.Set(key, fmt.Sprintf("perf_value_%d", i))
				} else {
					key := fmt.Sprintf("key_%d", i%1000)
					_, _ = cache.Get(key)
				}
			}

			duration := time.Since(start)
			opsPerSec := float64(operations) / duration.Seconds()
			stats := cache.GetStats()

			t.Logf("Configuration: %s", tc.name)
			t.Logf("Operations: %d", operations)
			t.Logf("Duration: %v", duration)
			t.Logf("Operations/sec: %.2f", opsPerSec)
			t.Logf("Hit Rate: %.2f%%", float64(stats.Hits)/float64(stats.Hits+stats.Misses)*100)
			t.Logf("Cache Size: %d", stats.Size)
			t.Logf("Cache Keys: %d", stats.Keys)
		})
	}
}
