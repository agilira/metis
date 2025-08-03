// main.go: Profiler for Metis strategic caching library
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"time"

	"github.com/agilira/metis"
)

// Configuration constants for the profiler
const (
	duration        = 5 * time.Second // Reduced duration for testing
	workers         = 8               // Reduced workers for testing
	keySpaceSize    = 10_000          // Reduced key space for testing
	valueSize       = 64              // Reduced value size for testing
	workload        = "balanced"      // Type of workload: read-heavy, write-heavy, balanced
	enableCompress  = false           // Disabled compression for testing
	evictionPolicy  = "wtinylfu"
	admissionPolicy = "always" // Simplified admission policy
	shardCount      = 16       // Reduced shard count
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	cache := metis.NewStrategicCache(metis.CacheConfig{
		EnableCaching:        true,
		CacheSize:            keySpaceSize,
		TTL:                  5 * time.Minute,
		EnableCompression:    enableCompress,
		EvictionPolicy:       evictionPolicy,
		AdmissionPolicy:      admissionPolicy,
		ShardCount:           shardCount,
		CleanupInterval:      30 * time.Second,
		MaxKeySize:           256,
		MaxValueSize:         1024,
		AdmissionProbability: 1.0, // Always admit when policy is "always"
	})
	defer cache.Close()

	cpuFile, err := os.Create("cpu.prof")
	if err == nil {
		_ = pprof.StartCPUProfile(cpuFile)
		defer func() {
			pprof.StopCPUProfile()
			// Ignore close error for profiling tool
			_ = cpuFile.Close()
		}()
	}

	fmt.Println("[WARMUP] Populating cache...")
	for i := 0; i < keySpaceSize/10; i++ { // Reduced warmup for testing
		key := fmt.Sprintf("key_%d", i)
		val := make([]byte, valueSize)
		cache.Set(key, val)
		if i%1000 == 0 {
			fmt.Printf("[WARMUP] Populated %d keys...\n", i)
		}
	}
	fmt.Println("[WARMUP] Cache population completed")

	fmt.Println("[BENCHMARK] Starting benchmark workload")

	var setStat, getStat opStat
	var totalOps int64
	var wg sync.WaitGroup
	stop := make(chan struct{})

	// Pre-allocate values for reuse
	valuePool := make([][]byte, 100) // Reduced pool size
	for i := 0; i < 100; i++ {
		valuePool[i] = make([]byte, valueSize)
	}

	fmt.Printf("[BENCHMARK] Starting %d workers for %v\n", workers, duration)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fmt.Printf("[WORKER] Worker %d started\n", id)
			// Use math/rand for performance profiling - cryptographic security not needed
			// nosec G404 - This is a performance profiler, not a security-critical application
			localRand := rand.New(rand.NewSource(time.Now().UnixNano() + int64(id)))
			ops := 0
			for {
				select {
				case <-stop:
					fmt.Printf("[WORKER] Worker %d finished with %d operations\n", id, ops)
					return
				default:
					keyID := localRand.Intn(keySpaceSize)
					key := fmt.Sprintf("key_%d", keyID)
					val := valuePool[keyID%100] // Reuse pre-allocated values
					opType := localRand.Intn(100)
					if workload == "read-heavy" && opType < 90 || workload == "balanced" && opType < 50 {
						start := time.Now()
						cache.Get(key)
						getStat.Record(time.Since(start))
					} else {
						start := time.Now()
						cache.Set(key, val)
						setStat.Record(time.Since(start))
					}
					atomic.AddInt64(&totalOps, 1)
					ops++
				}
			}
		}(i)
	}

	fmt.Println("[BENCHMARK] Waiting for completion...")
	time.Sleep(duration)
	fmt.Println("[BENCHMARK] Stopping workers...")
	close(stop)
	wg.Wait()
	fmt.Println("[BENCHMARK] All workers stopped")

	runtime.ReadMemStats(&memStats)

	fmt.Println("--- Results ---")
	fmt.Printf("Total operations: %d\n", totalOps)
	fmt.Printf("Set:  avg=%v min=%v max=%v\n", setStat.Avg(), setStat.Min, setStat.Max)
	fmt.Printf("Get:  avg=%v min=%v max=%v\n", getStat.Avg(), getStat.Min, getStat.Max)
	fmt.Printf("Ops/sec: %.2f\n", float64(totalOps)/duration.Seconds())
	fmt.Printf("Heap alloc: %d MB, GCs: %d, GC fraction: %.2f%%\n",
		memStats.HeapAlloc/1024/1024, memStats.NumGC, memStats.GCCPUFraction*100)

	// Export CSV
	csvFile, err := os.Create("metis_results.csv")
	if err == nil {
		defer csvFile.Close()
		writer := csv.NewWriter(csvFile)
		defer writer.Flush()

		// Write CSV data - ignore write errors for profiling tool
		_ = writer.Write([]string{"metric", "value"})
		_ = writer.Write([]string{"total_ops", fmt.Sprintf("%d", totalOps)})
		_ = writer.Write([]string{"set_avg_ns", fmt.Sprintf("%d", setStat.Avg().Nanoseconds())})
		_ = writer.Write([]string{"get_avg_ns", fmt.Sprintf("%d", getStat.Avg().Nanoseconds())})
		_ = writer.Write([]string{"ops_per_sec", fmt.Sprintf("%.2f", float64(totalOps)/duration.Seconds())})
		_ = writer.Write([]string{"heap_alloc_mb", fmt.Sprintf("%d", memStats.HeapAlloc/1024/1024)})
		_ = writer.Write([]string{"gc_count", fmt.Sprintf("%d", memStats.NumGC)})
		_ = writer.Write([]string{"gc_fraction", fmt.Sprintf("%.2f", memStats.GCCPUFraction*100)})
	}

	// Export JSON
	jsonData := map[string]interface{}{
		"total_ops":     totalOps,
		"set_avg_ns":    setStat.Avg().Nanoseconds(),
		"set_min_ns":    setStat.Min.Nanoseconds(),
		"set_max_ns":    setStat.Max.Nanoseconds(),
		"get_avg_ns":    getStat.Avg().Nanoseconds(),
		"get_min_ns":    getStat.Min.Nanoseconds(),
		"get_max_ns":    getStat.Max.Nanoseconds(),
		"ops_per_sec":   float64(totalOps) / duration.Seconds(),
		"heap_alloc_mb": memStats.HeapAlloc / 1024 / 1024,
		"gc_count":      memStats.NumGC,
		"gc_fraction":   memStats.GCCPUFraction * 100,
	}
	jsonFile, err := os.Create("metis_results.json")
	if err == nil {
		defer jsonFile.Close()
		encoder := json.NewEncoder(jsonFile)
		encoder.SetIndent("", "  ")
		// Ignore encode error for profiling tool
		_ = encoder.Encode(jsonData)
	}
}

// Global memory statistics for reporting
var memStats runtime.MemStats

// opStat keeps track of latency metrics for an operation type
type opStat struct {
	Min   time.Duration
	Max   time.Duration
	Total time.Duration
	Count int64
}

// Record registers a single operation latency into the statistics
func (s *opStat) Record(d time.Duration) {
	if s.Count == 0 || d < s.Min {
		s.Min = d
	}
	if d > s.Max {
		s.Max = d
	}
	s.Total += d
	s.Count++
}

// Avg returns the average latency for the recorded operations
func (s *opStat) Avg() time.Duration {
	if s.Count == 0 {
		return 0
	}
	return time.Duration(int64(s.Total) / s.Count)
}
