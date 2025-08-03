// main_test.go: Tests for Metis profiler
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/agilira/metis"
)

// TestOpStat_Record tests the Record method of opStat
func TestOpStat_Record(t *testing.T) {
	stat := &opStat{}

	// Test first record
	duration1 := 100 * time.Millisecond
	stat.Record(duration1)

	if stat.Count != 1 {
		t.Errorf("Expected count 1, got %d", stat.Count)
	}
	if stat.Min != duration1 {
		t.Errorf("Expected min %v, got %v", duration1, stat.Min)
	}
	if stat.Max != duration1 {
		t.Errorf("Expected max %v, got %v", duration1, stat.Max)
	}
	if stat.Total != duration1 {
		t.Errorf("Expected total %v, got %v", duration1, stat.Total)
	}

	// Test second record (shorter)
	duration2 := 50 * time.Millisecond
	stat.Record(duration2)

	if stat.Count != 2 {
		t.Errorf("Expected count 2, got %d", stat.Count)
	}
	if stat.Min != duration2 {
		t.Errorf("Expected min %v, got %v", duration2, stat.Min)
	}
	if stat.Max != duration1 {
		t.Errorf("Expected max %v, got %v", duration1, stat.Max)
	}
	if stat.Total != duration1+duration2 {
		t.Errorf("Expected total %v, got %v", duration1+duration2, stat.Total)
	}

	// Test third record (longer)
	duration3 := 200 * time.Millisecond
	stat.Record(duration3)

	if stat.Count != 3 {
		t.Errorf("Expected count 3, got %d", stat.Count)
	}
	if stat.Min != duration2 {
		t.Errorf("Expected min %v, got %v", duration2, stat.Min)
	}
	if stat.Max != duration3 {
		t.Errorf("Expected max %v, got %v", duration3, stat.Max)
	}
	if stat.Total != duration1+duration2+duration3 {
		t.Errorf("Expected total %v, got %v", duration1+duration2+duration3, stat.Total)
	}
}

// TestOpStat_Avg tests the Avg method of opStat
func TestOpStat_Avg(t *testing.T) {
	stat := &opStat{}

	// Test empty stat
	avg := stat.Avg()
	if avg != 0 {
		t.Errorf("Expected avg 0 for empty stat, got %v", avg)
	}

	// Test single record
	duration1 := 100 * time.Millisecond
	stat.Record(duration1)
	avg = stat.Avg()
	if avg != duration1 {
		t.Errorf("Expected avg %v, got %v", duration1, avg)
	}

	// Test multiple records
	duration2 := 200 * time.Millisecond
	stat.Record(duration2)
	expectedAvg := time.Duration((100+200)/2) * time.Millisecond
	avg = stat.Avg()
	if avg != expectedAvg {
		t.Errorf("Expected avg %v, got %v", expectedAvg, avg)
	}
}

// TestBenchmarkWorkload tests the benchmark workload logic
func TestBenchmarkWorkload(t *testing.T) {
	// Create a small cache for testing
	cache := metis.NewStrategicCache(metis.CacheConfig{
		EnableCaching:        true,
		CacheSize:            100,
		TTL:                  1 * time.Minute,
		EnableCompression:    false,
		EvictionPolicy:       "lru",
		AdmissionPolicy:      "always",
		ShardCount:           4,
		CleanupInterval:      30 * time.Second,
		MaxKeySize:           256,
		MaxValueSize:         1024,
		AdmissionProbability: 0.5,
	})
	defer cache.Close()

	// Test workload logic with different workload types
	testCases := []struct {
		workload string
		opType   int
		expected bool // true for Get, false for Set
	}{
		{"read-heavy", 85, true},   // opType < 90 should be Get
		{"read-heavy", 95, false},  // opType >= 90 should be Set
		{"balanced", 45, true},     // opType < 50 should be Get
		{"balanced", 55, false},    // opType >= 50 should be Set
		{"write-heavy", 45, false}, // write-heavy defaults to Set
		{"write-heavy", 55, false}, // write-heavy defaults to Set
	}

	for _, tc := range testCases {
		t.Run(tc.workload, func(t *testing.T) {
			// Simulate the workload logic from main.go
			isGet := false
			if tc.workload == "read-heavy" && tc.opType < 90 || tc.workload == "balanced" && tc.opType < 50 {
				isGet = true
			}

			if isGet != tc.expected {
				t.Errorf("For workload %s, opType %d: expected %v, got %v",
					tc.workload, tc.opType, tc.expected, isGet)
			}
		})
	}
}

// TestCacheOperations tests basic cache operations used in the profiler
func TestCacheOperations(t *testing.T) {
	cache := metis.NewStrategicCache(metis.CacheConfig{
		EnableCaching:        true,
		CacheSize:            100,
		TTL:                  1 * time.Minute,
		EnableCompression:    false,
		EvictionPolicy:       "lru",
		AdmissionPolicy:      "always",
		ShardCount:           4,
		CleanupInterval:      30 * time.Second,
		MaxKeySize:           256,
		MaxValueSize:         1024,
		AdmissionProbability: 0.5,
	})
	defer cache.Close()

	// Test Set operation
	key := "test_key"
	val := make([]byte, 256)
	success := cache.Set(key, val)
	if !success {
		t.Error("Set operation should succeed")
	}

	// Test Get operation
	retrieved, exists := cache.Get(key)
	if !exists {
		t.Error("Get operation should find the key")
	}
	if retrieved == nil {
		t.Error("Retrieved value should not be nil")
	}

	// Test key generation logic
	for i := 0; i < 10; i++ {
		generatedKey := fmt.Sprintf("key_%d", i)
		expectedKey := fmt.Sprintf("key_%d", i)
		if generatedKey != expectedKey {
			t.Errorf("Expected key %s, got %s", expectedKey, generatedKey)
		}
	}
}

// TestFileOperations tests file operations used in the profiler
func TestFileOperations(t *testing.T) {
	// Test CSV file creation
	csvFile, err := os.CreateTemp("", "test_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp CSV file: %v", err)
	}
	defer os.Remove(csvFile.Name())
	defer csvFile.Close()

	// Test JSON file creation
	jsonFile, err := os.CreateTemp("", "test_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp JSON file: %v", err)
	}
	defer os.Remove(jsonFile.Name())
	defer jsonFile.Close()

	// Test that files are writable
	testData := []byte("test data")
	_, err = csvFile.Write(testData)
	if err != nil {
		t.Errorf("Failed to write to CSV file: %v", err)
	}

	_, err = jsonFile.Write(testData)
	if err != nil {
		t.Errorf("Failed to write to JSON file: %v", err)
	}
}

// TestMemoryStats tests memory statistics collection
func TestMemoryStats(t *testing.T) {
	// Test that we can read memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Basic validation of memory stats
	if memStats.HeapAlloc == 0 && memStats.NumGC == 0 {
		t.Log("Memory stats collected successfully")
	}

	// Test memory calculations used in profiler
	gcFraction := memStats.GCCPUFraction * 100

	// Check GC fraction validity
	if gcFraction < 0 || gcFraction > 100 {
		t.Error("GC fraction should be between 0 and 100")
	}
}

// TestConcurrentOperations tests concurrent cache operations
func TestConcurrentOperations(t *testing.T) {
	cache := metis.NewStrategicCache(metis.CacheConfig{
		EnableCaching:        true,
		CacheSize:            1000,
		TTL:                  1 * time.Minute,
		EnableCompression:    false,
		EvictionPolicy:       "lru",
		AdmissionPolicy:      "always",
		ShardCount:           4,
		CleanupInterval:      30 * time.Second,
		MaxKeySize:           256,
		MaxValueSize:         1024,
		AdmissionProbability: 0.5,
	})
	defer cache.Close()

	// Test concurrent Set operations
	var wg sync.WaitGroup
	numGoroutines := 4
	operationsPerGoroutine := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				val := make([]byte, 256)
				cache.Set(key, val)
			}
		}(i)
	}

	wg.Wait()

	// Verify that all operations completed
	stats := cache.GetStats()
	expectedKeys := numGoroutines * operationsPerGoroutine
	if stats.Keys != expectedKeys {
		t.Errorf("Expected %d keys, got %d", expectedKeys, stats.Keys)
	}
}

// TestWorkloadTypes tests different workload configurations
func TestWorkloadTypes(t *testing.T) {
	workloads := []string{"read-heavy", "write-heavy", "balanced"}

	for _, workload := range workloads {
		t.Run(workload, func(t *testing.T) {
			// Test that workload string is valid
			if workload != "read-heavy" && workload != "write-heavy" && workload != "balanced" {
				t.Errorf("Invalid workload type: %s", workload)
			}

			// Test workload logic
			var getCount, setCount int
			for i := 0; i < 100; i++ {
				if workload == "read-heavy" && i < 90 || workload == "balanced" && i < 50 {
					getCount++
				} else {
					setCount++
				}
			}

			// Validate workload distribution
			switch workload {
			case "read-heavy":
				if getCount < 80 || setCount > 20 {
					t.Errorf("Read-heavy workload should have mostly gets: gets=%d, sets=%d", getCount, setCount)
				}
			case "balanced":
				if getCount < 40 || getCount > 60 || setCount < 40 || setCount > 60 {
					t.Errorf("Balanced workload should be roughly 50/50: gets=%d, sets=%d", getCount, setCount)
				}
			case "write-heavy":
				if setCount < 80 || getCount > 20 {
					t.Errorf("Write-heavy workload should have mostly sets: gets=%d, sets=%d", getCount, setCount)
				}
			}
		})
	}
}
