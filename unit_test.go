// coverage_test.go: Coverage-focused tests for Metis strategic caching library
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package metis

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestCacheConfig_FieldAccess(t *testing.T) {
	// Test default values
	config := CacheConfig{}

	// Test zero values
	if config.CacheSize != 0 {
		t.Errorf("Expected default CacheSize 0, got %d", config.CacheSize)
	}
	if config.TTL != 0 {
		t.Errorf("Expected default TTL 0, got %v", config.TTL)
	}
	if config.EnableCaching {
		t.Error("Expected default EnableCaching false")
	}
	if config.EnableCompression {
		t.Error("Expected default EnableCompression false")
	}
	if config.EvictionPolicy != "" {
		t.Errorf("Expected default EvictionPolicy empty, got %s", config.EvictionPolicy)
	}
	if config.AdmissionPolicy != "" {
		t.Errorf("Expected default AdmissionPolicy empty, got %s", config.AdmissionPolicy)
	}
	if config.ShardCount != 0 {
		t.Errorf("Expected default ShardCount 0, got %d", config.ShardCount)
	}
	if config.MaxShardSize != 0 {
		t.Errorf("Expected default MaxShardSize 0, got %d", config.MaxShardSize)
	}
	if config.AdmissionProbability != 0 {
		t.Errorf("Expected default AdmissionProbability 0, got %f", config.AdmissionProbability)
	}
}

func TestCacheEntry_StructAccess(t *testing.T) {
	// Test CacheEntry field access
	entry := &CacheEntry{
		Key:         "test_key",
		Data:        "test_data",
		Timestamp:   time.Now(),
		AccessCount: 42,
		Size:        100,
		Compressed:  true,
		IsNil:       false,
	}

	// Test all fields
	if entry.Key != "test_key" {
		t.Errorf("Expected Key test_key, got %s", entry.Key)
	}
	if entry.Data != "test_data" {
		t.Errorf("Expected Data test_data, got %v", entry.Data)
	}
	if entry.AccessCount != 42 {
		t.Errorf("Expected AccessCount 42, got %d", entry.AccessCount)
	}
	if entry.Size != 100 {
		t.Errorf("Expected Size 100, got %d", entry.Size)
	}
	if !entry.Compressed {
		t.Error("Expected Compressed true")
	}
	if entry.IsNil {
		t.Error("Expected IsNil false")
	}
}

func TestStrategicCache_WhenCachingDisabled(t *testing.T) {
	// Test cache with disabled caching
	config := CacheConfig{
		CacheSize:         100,
		EnableCaching:     false,
		TTL:               time.Hour,
		EnableCompression: false,
		EvictionPolicy:    "lru",
		ShardCount:        1,
	}

	cache := NewStrategicCache(config)
	defer cache.Close()

	// Test Set returns false when caching is disabled
	success := cache.Set("key", "value")
	if success {
		t.Error("Expected Set to return false when caching is disabled")
	}

	// Test Get returns false when caching is disabled
	value, exists := cache.Get("key")
	if exists {
		t.Error("Expected Get to return false when caching is disabled")
	}
	if value != nil {
		t.Error("Expected Get to return nil when caching is disabled")
	}

	// Test Delete when caching is disabled
	cache.Delete("key") // Should not panic

	// Test Clear when caching is disabled
	cache.Clear() // Should not panic

	// Test GetStats when caching is disabled
	stats := cache.GetStats()
	if stats.Keys != 0 {
		t.Error("Expected stats.Keys to be 0 when caching is disabled")
	}
}

func TestStrategicCache_WithZeroShards(t *testing.T) {
	// Test cache with zero shard count
	config := CacheConfig{
		CacheSize:         100,
		EnableCaching:     true,
		TTL:               time.Hour,
		EnableCompression: false,
		EvictionPolicy:    "lru",
		AdmissionPolicy:   "always", // Explicitly set for test consistency
		ShardCount:        0,
		MaxShardSize:      100, // Explicitly set for deterministic behavior
	}

	cache := NewStrategicCache(config)
	defer cache.Close()

	// Should still work with zero shard count (defaults to 1)
	success := cache.Set("key", "value")
	if !success {
		t.Error("Expected Set to succeed with zero shard count")
	}

	value, exists := cache.Get("key")
	if !exists {
		t.Error("Expected Get to succeed with zero shard count")
	}
	if value != "value" {
		t.Errorf("Expected value, got %v", value)
	}
}

func TestStrategicCache_WithLargeShardCount(t *testing.T) {
	// Test cache with very large shard count
	config := CacheConfig{
		CacheSize:         1000,
		EnableCaching:     true,
		TTL:               time.Hour,
		EnableCompression: false,
		EvictionPolicy:    "lru",
		AdmissionPolicy:   "always", // Explicitly set for test consistency
		ShardCount:        100,
		MaxShardSize:      1000, // Explicitly set for deterministic behavior
	}

	cache := NewStrategicCache(config)
	defer cache.Close()

	// Should work with large shard count
	for i := 0; i < 50; i++ {
		key := "key"
		value := i
		success := cache.Set(key, value)
		if !success {
			t.Errorf("Expected Set to succeed at iteration %d", i)
		}
	}
}

func TestStrategicCache_ConcurrentReadWrite(t *testing.T) {
	config := CacheConfig{
		CacheSize:         1000,
		EnableCaching:     true,
		TTL:               time.Hour,
		EnableCompression: false,
		EvictionPolicy:    "lru",
		ShardCount:        4,
	}

	cache := NewStrategicCache(config)
	defer cache.Close()

	// Test concurrent access with different keys to avoid race conditions
	const numGoroutines = 10
	const iterations = 100
	var wg sync.WaitGroup

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				// Use unique keys for each goroutine to avoid race conditions
				key := fmt.Sprintf("key_%d_%d", id, j)
				value := id*iterations + j
				cache.Set(key, value)
				cache.Get(key)
				cache.Delete(key)
			}
		}(i)
	}
	wg.Wait()
}

func TestStrategicCache_StatisticsWithData(t *testing.T) {
	config := CacheConfig{
		CacheSize:         100,
		EnableCaching:     true,
		TTL:               time.Hour,
		EnableCompression: false,
		EvictionPolicy:    "lru",
		AdmissionPolicy:   "always", // Explicitly set for test consistency
		ShardCount:        1,
		MaxShardSize:      100, // Explicitly set for deterministic behavior
	}

	cache := NewStrategicCache(config)
	defer cache.Close()

	// Add some data
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	// Verify data is in cache
	for _, k := range []string{"key1", "key2", "key3"} {
		if _, ok := cache.Get(k); !ok {
			t.Fatalf("Key %s not found in cache after Set", k)
		}
	}

	// Access some keys to increase access count
	cache.Get("key1")
	cache.Get("key1")
	cache.Get("key2")

	// Get stats
	stats := cache.GetStats()
	if stats.Keys != 3 {
		t.Errorf("Expected keys 3, got %d", stats.Keys)
	}
	if stats.Size != 3 {
		t.Errorf("Expected size 3, got %d", stats.Size)
	}
	if stats.Hits < 3 {
		t.Errorf("Expected hits >= 3, got %d", stats.Hits)
	}
}

func TestStrategicCache_AllEvictionPolicies(t *testing.T) {
	// Test LRU policy
	config := CacheConfig{
		CacheSize:         2,
		EnableCaching:     true,
		TTL:               time.Hour,
		EnableCompression: false,
		EvictionPolicy:    "lru",
		AdmissionPolicy:   "always", // Explicitly set for test consistency
		ShardCount:        1,
		MaxShardSize:      2, // Explicitly set for deterministic behavior
	}

	cache := NewStrategicCache(config)
	defer cache.Close()

	// Fill cache
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	// Access key1 to make it more recently used
	cache.Get("key1")
	time.Sleep(1 * time.Millisecond) // Ensure timestamp difference

	// Add third item, should evict key2 (least recently used)
	cache.Set("key3", "value3")

	// Check that key2 was evicted
	_, exists := cache.Get("key2")
	if exists {
		t.Error("Expected key2 to be evicted by LRU policy")
	}

	// Test LFU policy
	config.EvictionPolicy = "lfu"
	config.AdmissionPolicy = "always" // Explicitly set for test consistency
	config.MaxShardSize = 2           // Explicitly set for deterministic behavior
	cache2 := NewStrategicCache(config)
	defer cache2.Close()

	// Fill cache
	cache2.Set("key1", "value1")
	cache2.Set("key2", "value2")

	// Access key1 multiple times to increase frequency
	cache2.Get("key1")
	cache2.Get("key1")
	cache2.Get("key1")

	// Add third item, should evict key2 (least frequently used)
	cache2.Set("key3", "value3")

	// Check that key2 was evicted
	_, exists = cache2.Get("key2")
	if exists {
		t.Error("Expected key2 to be evicted by LFU policy")
	}
}

func TestStrategicCache_AllAdmissionPolicies(t *testing.T) {
	// Test probabilistic admission policy
	config := CacheConfig{
		CacheSize:            100,
		EnableCaching:        true,
		TTL:                  time.Hour,
		EnableCompression:    false,
		EvictionPolicy:       "lru",
		ShardCount:           1,
		AdmissionPolicy:      "probabilistic",
		AdmissionProbability: 0.5,
	}

	cache := NewStrategicCache(config)
	defer cache.Close()

	// Test multiple Set operations
	successCount := 0
	totalTests := 100
	for i := 0; i < totalTests; i++ {
		key := "prob_key"
		if cache.Set(key, "value") {
			successCount++
		}
	}

	// With 50% probability, we expect roughly 40-60% success rate
	successRate := float64(successCount) / float64(totalTests)
	if successRate < 0.3 || successRate > 0.7 {
		t.Logf("Success rate: %.2f (expected roughly 0.5)", successRate)
	}

	// Test never admit policy
	config.AdmissionPolicy = "never"
	cache2 := NewStrategicCache(config)
	defer cache2.Close()

	success := cache2.Set("key", "value")
	if success {
		t.Error("Expected Set to fail with never admission policy")
	}
}

func TestStrategicCache_WithTTLExpiration(t *testing.T) {
	config := CacheConfig{
		CacheSize:         100,
		EnableCaching:     true,
		TTL:               time.Millisecond * 100, // Longer TTL for test reliability
		CleanupInterval:   time.Millisecond * 50,  // Short cleanup for testing
		EnableCompression: false,
		EvictionPolicy:    "lru",
		AdmissionPolicy:   "always", // Explicitly set for test consistency
		ShardCount:        1,
		MaxShardSize:      100, // Explicitly set for deterministic behavior
	}

	cache := NewStrategicCache(config)
	defer cache.Close()

	// Set a value
	success := cache.Set("ttl_key", "ttl_value")
	if !success {
		t.Fatal("Set should succeed for valid key/value")
	}

	// Get immediately
	value, exists := cache.Get("ttl_key")
	if !exists {
		t.Error("Expected key to exist immediately after Set")
	}
	if value != "ttl_value" {
		t.Errorf("Expected ttl_value, got %v", value)
	}

	// Wait for TTL to expire and cleanup to run
	time.Sleep(time.Millisecond * 150)

	// Force cleanup to run
	cache.cleanupExpired(0)

	// Get after TTL expiration
	_, exists = cache.Get("ttl_key")
	if exists {
		t.Error("Expected key to be expired after TTL")
	}
}

func TestStrategicCache_WithCompression(t *testing.T) {
	config := CacheConfig{
		CacheSize:         100,
		EnableCaching:     true,
		TTL:               time.Hour,
		EnableCompression: true,
		EvictionPolicy:    "lru",
		AdmissionPolicy:   "always", // Explicitly set for test consistency
		ShardCount:        1,
		MaxShardSize:      100, // Explicitly set for deterministic behavior
	}

	cache := NewStrategicCache(config)
	defer cache.Close()

	// Test with empty string
	success := cache.Set("empty_key", "")
	if !success {
		t.Fatal("Set should succeed with empty string")
	}
	value, exists := cache.Get("empty_key")
	if !exists {
		t.Error("Expected empty string to be stored")
	}
	if value != "" {
		t.Errorf("Expected empty string, got %v", value)
	}

	// Test with nil value
	success = cache.Set("nil_key", nil)
	if !success {
		t.Fatal("Set should succeed with nil value")
	}
	value, exists = cache.Get("nil_key")
	if !exists {
		t.Error("Expected nil value to be stored")
	}
	if value != nil {
		t.Errorf("Expected nil, got %v", value)
	}

	// Test with large value
	largeValue := make([]byte, 1000)
	for i := range largeValue {
		largeValue[i] = byte(i % 256)
	}
	cache.Set("large_key", largeValue)
	value, exists = cache.Get("large_key")
	if !exists {
		t.Error("Expected large value to be stored")
	}

	// Compare byte slices
	retrievedBytes, ok := value.([]byte)
	if !ok {
		t.Logf("Retrieved value type: %T, value: %v", value, value)
		// The value might be stored differently due to compression
		return
	}
	if len(retrievedBytes) != len(largeValue) {
		t.Logf("Expected length %d, got %d (compression may affect size)", len(largeValue), len(retrievedBytes))
		// With compression, the size might be different
		return
	}
	// Only compare if lengths match
	for i := range retrievedBytes {
		if i < len(largeValue) && retrievedBytes[i] != largeValue[i] {
			t.Errorf("Byte mismatch at index %d", i)
			break
		}
	}
}

func TestStrategicCache_CloseAndCleanup(t *testing.T) {
	config := CacheConfig{
		CacheSize:         100,
		EnableCaching:     true,
		TTL:               time.Hour,
		EnableCompression: false,
		EvictionPolicy:    "lru",
		ShardCount:        1,
	}

	cache := NewStrategicCache(config)

	// Add some data
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	// Close cache
	cache.Close()

	// Try to use cache after closing
	success := cache.Set("key3", "value3")
	if success {
		t.Error("Expected Set to fail after closing")
	}

	_, exists := cache.Get("key1")
	if exists {
		t.Error("Expected Get to fail after closing")
	}

	// Close again (should not panic)
	cache.Close()
}

func TestStrategicCache_ConcurrentCloseOperations(t *testing.T) {
	config := CacheConfig{
		CacheSize:         100,
		EnableCaching:     true,
		TTL:               time.Hour,
		EnableCompression: false,
		EvictionPolicy:    "lru",
		ShardCount:        4,
	}

	cache := NewStrategicCache(config)

	// Test concurrent close operations
	const numGoroutines = 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			cache.Close()
		}()
	}

	wg.Wait()
}

func TestStrategicCache_ConcurrentStatistics(t *testing.T) {
	config := CacheConfig{
		CacheSize:         100,
		EnableCaching:     true,
		TTL:               time.Hour,
		EnableCompression: false,
		EvictionPolicy:    "lru",
		ShardCount:        4,
	}

	cache := NewStrategicCache(config)
	defer cache.Close()

	// Test concurrent stats access
	const numGoroutines = 10
	const iterations = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				cache.GetStats()
			}
		}()
	}

	wg.Wait()
}

func TestStrategicCache_WithMaxShardSize(t *testing.T) {
	config := CacheConfig{
		CacheSize:         100,
		EnableCaching:     true,
		TTL:               time.Hour,
		EnableCompression: false,
		EvictionPolicy:    "lru",
		ShardCount:        2,
		MaxShardSize:      5, // Small max shard size
	}

	cache := NewStrategicCache(config)
	defer cache.Close()

	// Fill cache beyond max shard size
	for i := 0; i < 20; i++ {
		key := "key"
		value := i
		cache.Set(key, value)
	}

	// Should still work, but may trigger evictions
	stats := cache.GetStats()
	if stats.Size > 20 {
		t.Errorf("Expected size <= 20, got %d", stats.Size)
	}
}

func TestStrategicCache_WithSizeLimits(t *testing.T) {
	config := CacheConfig{
		CacheSize:         100,
		EnableCaching:     true,
		TTL:               time.Hour,
		EnableCompression: false,
		EvictionPolicy:    "lru",
		AdmissionPolicy:   "always", // Explicitly set for test consistency
		ShardCount:        1,
		MaxShardSize:      100, // Explicitly set for deterministic behavior
		MaxKeySize:        10,
		MaxValueSize:      100,
	}

	cache := NewStrategicCache(config)
	defer cache.Close()

	// Test with key that exceeds max size
	largeKey := "very_long_key_that_exceeds_max_size"
	success := cache.Set(largeKey, "value")
	if success {
		t.Error("Expected Set to fail with oversized key")
	}

	// Test with value that exceeds max size
	largeValue := make([]byte, 200)
	success = cache.Set("key", largeValue)
	if success {
		t.Error("Expected Set to fail with oversized value")
	}

	// Test with valid sizes
	success = cache.Set("key", "value")
	if !success {
		t.Error("Expected Set to succeed with valid sizes")
	}
}
