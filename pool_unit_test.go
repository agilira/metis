// pool_unit_test.go: Unit tests for pool functions
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package metis

import (
	"testing"
	"time"
)

// TestCacheEntry_Pool tests the cache entry pool functions
func TestCacheEntry_Pool(t *testing.T) {
	// Get entry from pool
	entry := GetCacheEntry()
	if entry == nil {
		t.Error("GetCacheEntry should not return nil")
		return
	}

	// Modify entry
	entry.Key = "test_key"
	entry.Data = []byte("test_data")
	entry.Timestamp = time.Now()
	entry.AccessCount = 5
	entry.Size = 100

	// Return entry to pool
	PutCacheEntry(entry)

	// Entry should be cleared
	if entry.Key != "" {
		t.Error("entry key should be cleared after putting back to pool")
	}
	if entry.Data != nil {
		t.Error("entry data should be cleared after putting back to pool")
	}
	if entry.AccessCount != 0 {
		t.Error("entry access count should be cleared after putting back to pool")
	}
	if entry.Size != 0 {
		t.Error("entry size should be cleared after putting back to pool")
	}
	// Ensure entry is valid after returning to pool
	// Test with nil entry
	PutCacheEntry(nil) // Should not panic
}

// TestStrategicCache_LRU_Set tests Set function with LRU to improve coverage
func TestStrategicCache_LRU_Set(t *testing.T) {
	config := CacheConfig{
		EnableCaching:  true,
		CacheSize:      10,
		EvictionPolicy: "lru",
		ShardCount:     1,
		MaxValueSize:   100,
		MaxKeySize:     50,
	}
	cache := NewStrategicCache(config)
	defer cache.Close()

	// Test normal set
	success := cache.Set("key1", "value1")
	if !success {
		t.Error("Set should succeed for normal key-value")
	}

	// Test with oversized key
	oversizedKey := string(make([]byte, 100)) // Bigger than MaxKeySize
	success = cache.Set(oversizedKey, "value")
	if success {
		t.Error("Set should fail for oversized key")
	}

	// Test with oversized value
	oversizedValue := string(make([]byte, 200)) // Bigger than MaxValueSize
	success = cache.Set("key2", oversizedValue)
	if success {
		t.Error("Set should fail for oversized value")
	}

	// Test with nil value
	success = cache.Set("key3", nil)
	if !success {
		t.Error("Set should succeed for nil value")
	}

	// Verify nil value can be retrieved
	value, exists := cache.Get("key3")
	if !exists {
		t.Error("nil value should exist in cache")
	}
	if value != nil {
		t.Errorf("retrieved value should be nil, got %v", value)
	}
}

// TestStrategicCache_Get_ErrorCases tests Get function error cases
func TestStrategicCache_Get_ErrorCases(t *testing.T) {
	config := CacheConfig{
		EnableCaching:     true,
		CacheSize:         10,
		EvictionPolicy:    "lru",
		ShardCount:        1,
		EnableCompression: true,
	}
	cache := NewStrategicCache(config)
	defer cache.Close()

	// Test getting non-existent key
	value, exists := cache.Get("non_existent")
	if exists {
		t.Error("non-existent key should not exist")
	}
	if value != nil {
		t.Errorf("non-existent key should return nil value, got %v", value)
	}

	// Test with compressed data
	largeValue := string(make([]byte, 1000))
	cache.Set("large_key", largeValue)

	retrievedValue, exists := cache.Get("large_key")
	if !exists {
		t.Error("compressed value should exist")
	}
	if retrievedValue != largeValue {
		t.Error("compressed value should be decompressed correctly")
	}
}

// TestWTinyLFU_WindowSize tests WindowSize function
func TestWTinyLFU_WindowSize(t *testing.T) {
	cache := NewWTinyLFU(100, 4)

	windowSize := cache.WindowSize()
	if windowSize < 0 {
		t.Errorf("window size should be non-negative, got %d", windowSize)
	}
}

// TestWTinyLFU_MainSize tests MainSize function
func TestWTinyLFU_MainSize(t *testing.T) {
	cache := NewWTinyLFU(100, 4)

	mainSize := cache.MainSize()
	if mainSize < 0 {
		t.Errorf("main size should be non-negative, got %d", mainSize)
	}
}

// TestWTinyLFU_AdmissionFilter tests AdmissionFilter function
func TestWTinyLFU_AdmissionFilter(t *testing.T) {
	cache := NewWTinyLFU(100, 4)

	// Add some items to establish frequency
	cache.Set("frequent", "value")
	cache.Get("frequent")
	cache.Get("frequent")

	// Test admission filter
	filter := cache.AdmissionFilter()
	// Result should not be nil
	if filter == nil {
		t.Error("admission filter should not be nil")
	}
	t.Logf("Admission filter: %v", filter)
}

// TestCompressionErrorPaths tests compression error handling
func TestCompressionErrorPaths(t *testing.T) {
	// Test compression with valid data
	testData := []byte("test data for compression")
	compressed, err := compressGzipWithHeader(testData, "test-header")

	if err != nil {
		t.Errorf("compression should succeed for valid data: %v", err)
	}

	if compressed == nil {
		t.Error("compressed data should not be nil")
	}

	// Test with empty data
	emptyData := []byte{}
	_, err = compressGzipWithHeader(emptyData, "empty-header")

	if err != nil {
		t.Errorf("compression should handle empty data: %v", err)
	}
}

// TestConfigValidation_EdgeCases tests edge cases in config validation
func TestConfigValidation_EdgeCases(t *testing.T) {
	// Test empty config
	emptyConfig := CacheConfig{}
	result := ValidateConfig(emptyConfig)

	// Should have warnings about zero cache size
	if result.IsValid {
		t.Error("empty config should be invalid")
	}
	if len(result.Warnings) == 0 {
		t.Error("empty config should have warnings")
	}

	// Test config with negative values
	negativeConfig := CacheConfig{
		CacheSize:  -100,
		ShardCount: -1,
		TTL:        -1 * time.Hour,
	}
	result = ValidateConfig(negativeConfig)

	if result.IsValid {
		t.Error("config with negative values should be invalid")
	}
}
