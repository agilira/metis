// integration_stress_test.go: Integration and Stress Tests for Metis
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package metis

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestExtremeRaceConditions tests race conditions under extreme concurrency
func TestExtremeRaceConditions(t *testing.T) {
	config := CacheConfig{
		EnableCaching:     true,
		CacheSize:         1000,
		ShardCount:        4,
		TTL:               100 * time.Millisecond,
		CleanupInterval:   10 * time.Millisecond,
		EnableCompression: true,
		EvictionPolicy:    "wtinylfu",
		AdmissionPolicy:   "always",
	}
	cache := NewStrategicCache(config)
	defer cache.Close()

	// Test extreme concurrent access with same key
	const numGoroutines = 100
	const operations = 50
	var wg sync.WaitGroup
	var successfulOps int64
	var failedOps int64

	key := "race_condition_key"

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				// Mix operations randomly
				switch j % 4 {
				case 0:
					// Set operation
					value := fmt.Sprintf("worker_%d_op_%d", workerID, j)
					if cache.Set(key, value) {
						atomic.AddInt64(&successfulOps, 1)
					} else {
						atomic.AddInt64(&failedOps, 1)
					}
				case 1:
					// Get operation
					if _, exists := cache.Get(key); exists {
						atomic.AddInt64(&successfulOps, 1)
					}
				case 2:
					// Delete operation
					cache.Delete(key)
					atomic.AddInt64(&successfulOps, 1)
				case 3:
					// Stats operation
					stats := cache.GetStats()
					if stats.Size >= 0 {
						atomic.AddInt64(&successfulOps, 1)
					}
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify cache is still functional after extreme race conditions
	cache.Set("final_race_test", "success")
	value, exists := cache.Get("final_race_test")
	if !exists || value != "success" {
		t.Error("Cache should still be functional after extreme race conditions")
	}

	t.Logf("Race condition test: successful ops=%d, failed ops=%d", successfulOps, failedOps)
}

// TestMemoryLeakPrevention tests that no memory leaks occur during intensive operations
func TestMemoryLeakPrevention(t *testing.T) {
	// Get initial memory stats
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	config := CacheConfig{
		EnableCaching:     true,
		CacheSize:         1000,
		ShardCount:        16,
		TTL:               10 * time.Millisecond,
		CleanupInterval:   5 * time.Millisecond,
		EnableCompression: true,
		EvictionPolicy:    "wtinylfu",
		AdmissionPolicy:   "always",
	}

	// Create and destroy multiple caches
	for iteration := 0; iteration < 10; iteration++ {
		cache := NewStrategicCache(config)

		// Intensive operations
		var wg sync.WaitGroup
		const numWorkers = 10
		const opsPerWorker = 100

		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				for j := 0; j < opsPerWorker; j++ {
					key := fmt.Sprintf("leak_test_%d_%d", workerID, j)
					value := make([]byte, 1024) // 1KB per value
					for k := range value {
						value[k] = byte(k % 256)
					}

					cache.Set(key, value)
					cache.Get(key)
					if j%10 == 0 {
						cache.Delete(key)
					}
				}
			}(i)
		}

		wg.Wait()

		// Wait for cleanup
		time.Sleep(50 * time.Millisecond)

		// Close and cleanup
		cache.Close()

		// Force garbage collection
		runtime.GC()
		runtime.GC() // Call twice to be sure
	}

	// Check final memory stats
	runtime.GC()
	runtime.ReadMemStats(&m2)

	// Memory should not have grown significantly
	var memoryGrowth uint64
	if m2.HeapInuse > m1.HeapInuse {
		memoryGrowth = m2.HeapInuse - m1.HeapInuse
	} else {
		memoryGrowth = 0 // Memory actually decreased
	}
	const maxAcceptableGrowth = 50 * 1024 * 1024 // 50MB

	if memoryGrowth > maxAcceptableGrowth {
		t.Errorf("Potential memory leak detected: memory grew by %d bytes (max acceptable: %d)",
			memoryGrowth, maxAcceptableGrowth)
	}

	t.Logf("Memory leak test: initial=%d bytes, final=%d bytes, growth=%d bytes",
		m1.HeapInuse, m2.HeapInuse, memoryGrowth)
}

// TestComprehensiveStressTest runs a comprehensive stress test covering all aspects
func TestComprehensiveStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping comprehensive stress test in short mode")
	}

	config := CacheConfig{
		EnableCaching:        true,
		CacheSize:            1000,
		ShardCount:           16,
		TTL:                  200 * time.Millisecond,
		CleanupInterval:      50 * time.Millisecond,
		EnableCompression:    true,
		EvictionPolicy:       "wtinylfu",
		AdmissionPolicy:      "probabilistic",
		AdmissionProbability: 0.8,
		MaxKeySize:           1024,
		MaxValueSize:         8192,
	}
	cache := NewStrategicCache(config)
	defer cache.Close()

	const duration = 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	var operations int64
	var setErrors int64
	var getErrors int64
	var wg sync.WaitGroup

	// Mixed workload workers
	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			localOps := 0
			localSetErrors := 0
			localGetErrors := 0

			ticker := time.NewTicker(time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					atomic.AddInt64(&operations, int64(localOps))
					atomic.AddInt64(&setErrors, int64(localSetErrors))
					atomic.AddInt64(&getErrors, int64(localGetErrors))
					return
				case <-ticker.C:
					// Mixed operations
					switch localOps % 10 {
					case 0, 1, 2, 3, 4: // 50% Set operations
						key := fmt.Sprintf("stress_%d_%d", workerID, localOps)
						value := make([]byte, 100+localOps%500)
						for j := range value {
							value[j] = byte(j % 256)
						}
						if !cache.Set(key, value) {
							localSetErrors++
						}
					case 5, 6, 7: // 30% Get operations
						key := fmt.Sprintf("stress_%d_%d", workerID, localOps-10)
						_, exists := cache.Get(key)
						if !exists && localOps > 10 {
							// This is expected for TTL expiration or eviction - not an error
						}
					case 8: // 10% Delete operations
						key := fmt.Sprintf("stress_%d_%d", workerID, localOps-5)
						cache.Delete(key)
					case 9: // 10% Stats operations
						stats := cache.GetStats()
						if stats.Size < 0 {
							localGetErrors++
						}
					}
					localOps++
				}
			}
		}(i)
	}

	wg.Wait()

	// Final verification
	finalStats := cache.GetStats()

	t.Logf("Comprehensive stress test completed:")
	t.Logf("  Duration: %v", duration)
	t.Logf("  Total operations: %d", operations)
	t.Logf("  Set errors: %d", setErrors)
	t.Logf("  Get errors: %d", getErrors)
	t.Logf("  Final cache size: %d", finalStats.Size)
	t.Logf("  Final hits: %d", finalStats.Hits)
	t.Logf("  Final misses: %d", finalStats.Misses)
	t.Logf("  Operations/second: %.2f", float64(operations)/duration.Seconds())

	// Set errors can be legitimate (admission policy, size limits, etc.)
	// With probabilistic admission at 0.8, we expect ~20% rejection rate
	// But should not be excessive beyond the expected rejection rate
	totalErrors := setErrors + getErrors
	expectedRejectionRate := 0.25 // Allow 25% to account for admission policy + other factors
	maxAllowedErrors := int64(float64(operations) * expectedRejectionRate)

	if totalErrors > maxAllowedErrors {
		t.Errorf("Too many errors: %d/%d (%.2f%%) - Expected around %.1f%% due to admission policy",
			totalErrors, operations, float64(totalErrors)*100/float64(operations), expectedRejectionRate*100)
	} else {
		t.Logf("Error rate within expected bounds: %d/%d (%.2f%%)",
			totalErrors, operations, float64(totalErrors)*100/float64(operations))
	}
}
