// targeted_compression_test.go: Step 3.2 - Targeted tests for specific coverage gaps
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package metis

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestCacheClosedOperations tests operations on a closed cache
func TestCacheClosedOperations(t *testing.T) {
	config := CacheConfig{
		EnableCaching:     true,
		CacheSize:         100,
		ShardCount:        4,
		TTL:               time.Minute,
		CleanupInterval:   30 * time.Second,
		EnableCompression: false,
		EvictionPolicy:    "lru",
		AdmissionPolicy:   "always",
	}
	cache := NewStrategicCache(config)

	// Close the cache
	cache.Close()

	// Test operations on closed cache
	success := cache.Set("closed_test", "value")
	if success {
		t.Error("Set should fail on closed cache")
	}

	_, exists := cache.Get("closed_test")
	if exists {
		t.Error("Get should not find values in closed cache")
	}

	// Delete should not panic on closed cache
	cache.Delete("closed_test")

	// Stats should still work
	stats := cache.GetStats()
	if stats.Size != 0 {
		t.Errorf("Closed cache should have zero size, got %d", stats.Size)
	}

	// Multiple closes should not panic
	cache.Close()
	cache.Close()
}

// TestCleanupTimeout tests cleanup goroutine timeout scenario
func TestCleanupTimeout(t *testing.T) {
	config := CacheConfig{
		EnableCaching:     true,
		CacheSize:         10,
		ShardCount:        1,
		TTL:               10 * time.Millisecond,
		CleanupInterval:   5 * time.Millisecond,
		EnableCompression: false,
		EvictionPolicy:    "lru",
		AdmissionPolicy:   "always",
	}
	cache := NewStrategicCache(config)

	// Add items that will expire
	for i := 0; i < 5; i++ {
		cache.Set(fmt.Sprintf("expire_%d", i), i)
	}

	// Wait for expiration
	time.Sleep(50 * time.Millisecond)

	// Force close with timeout scenario
	done := make(chan struct{})
	go func() {
		cache.Close()
		close(done)
	}()

	select {
	case <-done:
		t.Log("Cache closed successfully")
	case <-time.After(10 * time.Second):
		t.Error("Cache close timed out")
	}
}

// TestShardBoundaryConditions tests shard boundary conditions
func TestShardBoundaryConditions(t *testing.T) {
	config := CacheConfig{
		EnableCaching:     true,
		CacheSize:         16,
		ShardCount:        4, // 4 items per shard
		EnableCompression: false,
		EvictionPolicy:    "lru",
		AdmissionPolicy:   "always",
		MaxShardSize:      4,
	}
	cache := NewStrategicCache(config)
	defer cache.Close()

	// Fill each shard to capacity
	for i := 0; i < 20; i++ { // More than total capacity to trigger evictions
		key := fmt.Sprintf("boundary_%d", i)
		cache.Set(key, i)
	}

	stats := cache.GetStats()
	if stats.Size > 16 {
		t.Errorf("Cache should not exceed max size, got %d", stats.Size)
	}

	// Test specific shard targeting
	shardKeys := make(map[int][]string)
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("shard_test_%d", i)
		cache.Set(key, i)

		// Try to determine which shard this key goes to by testing
		_, exists := cache.Get(key)
		if exists {
			shardIndex := i % 4 // Approximate shard assignment
			shardKeys[shardIndex] = append(shardKeys[shardIndex], key)
		}
	}

	t.Logf("Shard distribution test completed")
}

// TestCompressionBoundaryEdgeCases tests compression with edge cases
func TestCompressionBoundaryEdgeCases(t *testing.T) {
	config := CacheConfig{
		EnableCaching:     true,
		CacheSize:         100,
		ShardCount:        4,
		EnableCompression: true,
		EvictionPolicy:    "lru",
		AdmissionPolicy:   "always",
	}
	cache := NewStrategicCache(config)
	defer cache.Close()

	// Test with data exactly at compression threshold (64 bytes)
	data64 := make([]byte, 64)
	for i := range data64 {
		data64[i] = byte(i % 256)
	}
	cache.Set("threshold_64", data64)

	retrieved, exists := cache.Get("threshold_64")
	if !exists {
		t.Error("64-byte data should be stored")
	}
	if retrievedBytes, ok := retrieved.([]byte); !ok || len(retrievedBytes) != 64 {
		t.Error("64-byte data should be retrieved correctly")
	}

	// Test with data just above threshold (65 bytes)
	data65 := make([]byte, 65)
	for i := range data65 {
		data65[i] = byte(i % 256)
	}
	cache.Set("threshold_65", data65)

	retrieved, exists = cache.Get("threshold_65")
	if !exists {
		t.Error("65-byte data should be stored")
	}
	if retrievedBytes, ok := retrieved.([]byte); !ok || len(retrievedBytes) != 65 {
		t.Error("65-byte data should be retrieved correctly")
	}

	// Test with highly compressible data
	compressibleData := make([]byte, 1000)
	for i := range compressibleData {
		compressibleData[i] = 'A' // Highly compressible
	}
	cache.Set("compressible", compressibleData)

	retrieved, exists = cache.Get("compressible")
	if !exists {
		t.Error("Compressible data should be stored")
	}
	if retrievedBytes, ok := retrieved.([]byte); !ok || len(retrievedBytes) != 1000 {
		t.Error("Compressible data should be retrieved correctly")
	}
}

// TestEvictionPolicyBoundaries tests eviction policy edge cases
func TestEvictionPolicyBoundaries(t *testing.T) {
	config := CacheConfig{
		EnableCaching:     true,
		CacheSize:         5,
		ShardCount:        1,
		EnableCompression: false,
		EvictionPolicy:    "lru",
		AdmissionPolicy:   "always",
		MaxShardSize:      5,
	}
	cache := NewStrategicCache(config)
	defer cache.Close()

	// Fill cache to capacity
	for i := 0; i < 5; i++ {
		cache.Set(fmt.Sprintf("item_%d", i), i)
	}

	// Add one more to trigger eviction
	cache.Set("eviction_trigger", "trigger")

	// Check that eviction occurred
	stats := cache.GetStats()
	if stats.Size > 5 {
		t.Errorf("Cache should not exceed capacity, got %d", stats.Size)
	}

	// Verify oldest item was evicted (item_0 should be gone)
	_, exists := cache.Get("item_0")
	if exists {
		t.Error("Oldest item should have been evicted")
	}

	// Verify newest item exists
	_, exists = cache.Get("eviction_trigger")
	if !exists {
		t.Error("Newest item should exist")
	}
}

// TestTTLBoundaryConditions tests TTL edge cases
func TestTTLBoundaryConditions(t *testing.T) {
	config := CacheConfig{
		EnableCaching:     true,
		CacheSize:         100,
		ShardCount:        4,
		TTL:               1 * time.Millisecond,   // Very short TTL
		CleanupInterval:   500 * time.Microsecond, // Very frequent cleanup
		EnableCompression: false,
		EvictionPolicy:    "lru",
		AdmissionPolicy:   "always",
	}
	cache := NewStrategicCache(config)
	defer cache.Close()

	// Set item with very short TTL
	cache.Set("short_ttl", "expires_quickly")

	// Item might already be expired by the time we check
	_, exists := cache.Get("short_ttl")
	t.Logf("Item with 1ms TTL exists: %v", exists)

	// Wait for certain expiration
	time.Sleep(10 * time.Millisecond)

	// Should definitely be expired now
	_, exists = cache.Get("short_ttl")
	if exists {
		t.Error("Item should be expired after 10ms with 1ms TTL")
	}

	// Test concurrent access during TTL expiration
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("concurrent_ttl_%d", id)
			cache.Set(key, id)
			time.Sleep(2 * time.Millisecond)
			cache.Get(key) // May or may not exist due to TTL
		}(i)
	}
	wg.Wait()
}

// TestAdmissionPolicyEdgeCases tests admission policy edge cases
func TestAdmissionPolicyEdgeCases(t *testing.T) {
	// Test with never admit policy
	config := CacheConfig{
		EnableCaching:     true,
		CacheSize:         100,
		ShardCount:        4,
		EnableCompression: false,
		EvictionPolicy:    "lru",
		AdmissionPolicy:   "never",
	}
	cache := NewStrategicCache(config)
	defer cache.Close()

	// Set should fail with never admit policy
	success := cache.Set("never_admit", "value")
	if success {
		t.Error("Set should fail with never admit policy")
	}

	// Test with probabilistic policy at boundaries
	config.AdmissionPolicy = "probabilistic"
	config.AdmissionProbability = 0.0 // Should never admit
	cache2 := NewStrategicCache(config)
	defer cache2.Close()

	success = cache2.Set("prob_0", "value")
	if success {
		t.Error("Set should fail with 0.0 admission probability")
	}

	// Test with 1.0 probability
	config.AdmissionProbability = 1.0 // Should always admit
	cache3 := NewStrategicCache(config)
	defer cache3.Close()

	success = cache3.Set("prob_1", "value")
	if !success {
		t.Error("Set should succeed with 1.0 admission probability")
	}
}

// TestConcurrentCleanup tests concurrent cleanup operations
func TestConcurrentCleanup(t *testing.T) {
	config := CacheConfig{
		EnableCaching:     true,
		CacheSize:         50,
		ShardCount:        4,
		TTL:               20 * time.Millisecond,
		CleanupInterval:   5 * time.Millisecond,
		EnableCompression: false,
		EvictionPolicy:    "lru",
		AdmissionPolicy:   "always",
	}
	cache := NewStrategicCache(config)
	defer cache.Close()

	// Fill cache with items that will expire
	for i := 0; i < 40; i++ {
		cache.Set(fmt.Sprintf("cleanup_%d", i), i)
	}

	// Run concurrent operations while cleanup is happening
	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			count := 0
			for {
				select {
				case <-ctx.Done():
					return
				default:
					key := fmt.Sprintf("concurrent_%d_%d", workerID, count)
					cache.Set(key, count)
					cache.Get(key)
					count++
					time.Sleep(time.Millisecond)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify cache is still functional
	cache.Set("cleanup_final", "success")
	value, exists := cache.Get("cleanup_final")
	if !exists || value != "success" {
		t.Error("Cache should still be functional after concurrent cleanup")
	}
}

// TestMaxSizeValidation tests max size validation edge cases
func TestMaxSizeValidation(t *testing.T) {
	config := CacheConfig{
		EnableCaching:     true,
		CacheSize:         100,
		ShardCount:        4,
		EnableCompression: false,
		EvictionPolicy:    "lru",
		AdmissionPolicy:   "always",
		MaxKeySize:        10,
		MaxValueSize:      50,
	}
	cache := NewStrategicCache(config)
	defer cache.Close()

	// Test key exactly at limit
	key10 := "1234567890" // Exactly 10 characters
	success := cache.Set(key10, "value")
	if !success {
		t.Error("Set should succeed with key exactly at max size")
	}

	// Test key over limit
	key11 := "12345678901" // 11 characters
	success = cache.Set(key11, "value")
	if success {
		t.Error("Set should fail with key over max size")
	}

	// Test value exactly at limit
	value50 := make([]byte, 50)
	for i := range value50 {
		value50[i] = byte('A')
	}
	success = cache.Set("value_test", value50)
	if !success {
		t.Error("Set should succeed with value exactly at max size")
	}

	// Test value over limit
	value51 := make([]byte, 51)
	for i := range value51 {
		value51[i] = byte('A')
	}
	success = cache.Set("value_test2", value51)
	if success {
		t.Error("Set should fail with value over max size")
	}
}
