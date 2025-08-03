// stress_test.go: Stress tests for Metis strategic caching library
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package metis

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"
)

// StressTestConfig defines the configuration for stress testing
type StressTestConfig struct {
	Duration           time.Duration
	ConcurrentWorkers  int
	OperationsPerSec   int
	KeySpaceSize       int
	ValueSize          int
	CompressionEnabled bool
	TTL                time.Duration
	EvictionPolicy     string
	AdmissionPolicy    string
	ShardCount         int
}

// StressTestResult contains the results of stress testing
type StressTestResult struct {
	TotalOperations  int64
	SuccessfulGets   int64
	SuccessfulSets   int64
	FailedOperations int64
	AverageLatency   time.Duration
	MaxLatency       time.Duration
	MinLatency       time.Duration
	MemoryUsage      runtime.MemStats
	CPUProfile       string
	HeapProfile      string
}

// RunStressTest executes a comprehensive stress test on the cache
func RunStressTest(config StressTestConfig) (*StressTestResult, error) {
	// Set defaults
	if config.Duration == 0 {
		config.Duration = 30 * time.Second
	}
	if config.ConcurrentWorkers == 0 {
		config.ConcurrentWorkers = 10
	}
	if config.OperationsPerSec == 0 {
		config.OperationsPerSec = 1000
	}
	if config.KeySpaceSize == 0 {
		config.KeySpaceSize = 10000
	}
	if config.ValueSize == 0 {
		config.ValueSize = 1024
	}
	if config.ShardCount == 0 {
		config.ShardCount = 16
	}

	// Create cache configuration
	cacheConfig := CacheConfig{
		EnableCaching:        true,
		CacheSize:            10000,
		TTL:                  config.TTL,
		EnableCompression:    config.CompressionEnabled,
		EvictionPolicy:       config.EvictionPolicy,
		AdmissionPolicy:      config.AdmissionPolicy,
		ShardCount:           config.ShardCount,
		MaxKeySize:           256,
		MaxValueSize:         1024 * 1024,
		CleanupInterval:      1 * time.Minute,
		AdmissionProbability: 0.5,
	}

	// Initialize cache
	cache := NewStrategicCache(cacheConfig)
	defer cache.Close()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), config.Duration)
	defer cancel()

	// Initialize metrics
	var (
		totalOps       int64
		successfulGets int64
		successfulSets int64
		failedOps      int64
		latencies      []time.Duration
		latencyMutex   sync.Mutex
		wg             sync.WaitGroup
	)

	// Start memory profiling
	startMemStats := runtime.MemStats{}
	runtime.ReadMemStats(&startMemStats)

	// Worker function
	worker := func(workerID int) {
		defer wg.Done()

		// Calculate operations per worker
		opsPerWorker := config.OperationsPerSec / config.ConcurrentWorkers
		if opsPerWorker == 0 {
			opsPerWorker = 1
		}

		interval := time.Second / time.Duration(opsPerWorker)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Generate random data using local random generator
		rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Random operation: 70% gets, 30% sets
				op := rng.Float64()

				start := time.Now()
				var success bool

				if op < 0.7 {
					// GET operation
					key := fmt.Sprintf("key_%d_%d", workerID, rng.Intn(config.KeySpaceSize))
					_, exists := cache.Get(key)
					success = true // GET always succeeds, even if key doesn't exist
					if exists {
						successfulGets++
					}
				} else {
					// SET operation
					key := fmt.Sprintf("key_%d_%d", workerID, rng.Intn(config.KeySpaceSize))
					value := generateRandomValue(config.ValueSize)
					success = cache.Set(key, value)
					if success {
						successfulSets++
					}
				}

				latency := time.Since(start)

				latencyMutex.Lock()
				latencies = append(latencies, latency)
				latencyMutex.Unlock()

				if success {
					totalOps++
				} else {
					failedOps++
				}
			}
		}
	}

	// Start workers
	for i := 0; i < config.ConcurrentWorkers; i++ {
		wg.Add(1)
		go worker(i)
	}

	// Wait for completion
	wg.Wait()

	// Calculate statistics
	var avgLatency, maxLatency, minLatency time.Duration
	if len(latencies) > 0 {
		totalLatency := time.Duration(0)
		minLatency = latencies[0]
		maxLatency = latencies[0]

		for _, latency := range latencies {
			totalLatency += latency
			if latency > maxLatency {
				maxLatency = latency
			}
			if latency < minLatency {
				minLatency = latency
			}
		}
		avgLatency = totalLatency / time.Duration(len(latencies))
	}

	// Get final memory stats
	endMemStats := runtime.MemStats{}
	runtime.ReadMemStats(&endMemStats)

	return &StressTestResult{
		TotalOperations:  totalOps,
		SuccessfulGets:   successfulGets,
		SuccessfulSets:   successfulSets,
		FailedOperations: failedOps,
		AverageLatency:   avgLatency,
		MaxLatency:       maxLatency,
		MinLatency:       minLatency,
		MemoryUsage:      endMemStats,
	}, nil
}

// generateRandomValue creates a random byte slice of specified size
func generateRandomValue(size int) []byte {
	value := make([]byte, size)
	for i := range value {
		value[i] = byte(rand.Intn(256))
	}
	return value
}

// RunChaosTest simulates adverse conditions and system failures
func RunChaosTest(cache *StrategicCache, duration time.Duration) error {
	log.Println("Starting chaos engineering test...")

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	var wg sync.WaitGroup

	// Simulate concurrent access during cache operations
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					key := fmt.Sprintf("chaos_key_%d_%d", workerID, time.Now().UnixNano())
					value := fmt.Sprintf("chaos_value_%d", workerID)

					// Random operations
					switch rand.Intn(4) {
					case 0:
						cache.Set(key, value)
					case 1:
						cache.Get(key)
					case 2:
						cache.Delete(key)
					case 3:
						cache.GetStats()
					}
				}
			}
		}(i)
	}

	// Simulate cache close and reopen
	go func() {
		time.Sleep(duration / 2)
		log.Println("Simulating cache close...")
		cache.Close()

		time.Sleep(1 * time.Second)
		log.Println("Simulating cache reopen...")
		// Note: In real scenario, you'd create a new cache
	}()

	wg.Wait()
	return nil
}

// BenchmarkCachePerformance runs comprehensive performance benchmarks
func BenchmarkCachePerformance(b *testing.B) {
	// Create optimized cache configuration for maximum performance
	config := CacheConfig{
		EnableCaching:        true,
		CacheSize:            10000,
		TTL:                  5 * time.Minute,
		EnableCompression:    false, // Disable compression for better performance in benchmarks
		EvictionPolicy:       "wtinylfu",
		AdmissionPolicy:      "always", // Use always for maximum throughput
		ShardCount:           32,
		MaxKeySize:           256,
		MaxValueSize:         1024 * 1024,
		CleanupInterval:      1 * time.Minute,
		AdmissionProbability: 1.0, // Always admit for maximum throughput
	}

	cache := NewStrategicCache(config)
	defer cache.Close()

	// Pre-populate cache
	for i := 0; i < 5000; i++ {
		key := fmt.Sprintf("key_%d", i)
		value := generateRandomValue(512)
		cache.Set(key, value)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key_%d", i%10000)
			value := generateRandomValue(256)

			// Mix of operations
			switch i % 3 {
			case 0:
				cache.Set(key, value)
			case 1:
				cache.Get(key)
			case 2:
				cache.Delete(key)
			}
			i++
		}
	})
}
