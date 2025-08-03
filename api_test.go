// api_test.go: Tests for simplified API layer
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

func TestSimpleAPI_Basic(t *testing.T) {
	cache := New()
	defer cache.Close()

	// Test basic operations
	cache.Set("key1", "value1")
	cache.Set("key2", 42)
	cache.Set("key3", []string{"a", "b", "c"})

	// Test Get
	if value, exists := cache.Get("key1"); !exists || value != "value1" {
		t.Errorf("Expected 'value1', got %v", value)
	}

	if value, exists := cache.Get("key2"); !exists || value != 42 {
		t.Errorf("Expected 42, got %v", value)
	}

	if value, exists := cache.Get("key3"); !exists {
		t.Errorf("Expected slice, got %v", value)
	}

	// Test non-existent key
	if _, exists := cache.Get("nonexistent"); exists {
		t.Error("Expected false for non-existent key")
	}

	// Test Stats
	stats := cache.Stats()
	if stats.Size != 3 {
		t.Errorf("Expected size 3, got %d", stats.Size)
	}
}

func TestSimpleAPI_Delete(t *testing.T) {
	cache := New()
	defer cache.Close()

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	stats := cache.Stats()
	if stats.Size != 2 {
		t.Errorf("Expected size 2, got %d", stats.Size)
	}

	cache.Delete("key1")

	stats = cache.Stats()
	if stats.Size != 1 {
		t.Errorf("Expected size 1, got %d", stats.Size)
	}

	if _, exists := cache.Get("key1"); exists {
		t.Error("Expected key1 to be deleted")
	}

	if _, exists := cache.Get("key2"); !exists {
		t.Error("Expected key2 to still exist")
	}
}

func TestSimpleAPI_Clear(t *testing.T) {
	cache := New()
	defer cache.Close()

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	stats := cache.Stats()
	if stats.Size != 3 {
		t.Errorf("Expected size 3, got %d", stats.Size)
	}

	cache.Clear()

	stats = cache.Stats()
	if stats.Size != 0 {
		t.Errorf("Expected size 0, got %d", stats.Size)
	}

	if _, exists := cache.Get("key1"); exists {
		t.Error("Expected all keys to be cleared")
	}
}

func TestSimpleAPI_Stats(t *testing.T) {
	cache := New()
	defer cache.Close()

	// Add some data
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	// Access some keys
	cache.Get("key1") // Hit
	cache.Get("key1") // Hit
	cache.Get("key2") // Hit
	cache.Get("key3") // Miss

	stats := cache.Stats()

	if stats.Size != 2 {
		t.Errorf("Expected 2 keys, got %d", stats.Size)
	}

	if stats.Hits != 3 {
		t.Errorf("Expected 3 hits, got %d", stats.Hits)
	}

	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}

	if stats.HitRate < 75.0 || stats.HitRate > 76.0 {
		t.Errorf("Expected hit rate around 75.0%%, got %.1f%%", stats.HitRate)
	}

	// Test String method
	statsStr := stats.String()
	if statsStr == "" {
		t.Error("Expected non-empty stats string")
	}
}

func TestSimpleAPI_AdvancedUsage(t *testing.T) {
	// Test advanced usage with custom config
	customConfig := CacheConfig{
		EnableCaching:     true,
		CacheSize:         1000,
		TTL:               1 * time.Minute,
		EnableCompression: true,
		EvictionPolicy:    "lru",
		ShardCount:        16,
		AdmissionPolicy:   "always",
	}

	cache := NewWithConfig(customConfig)
	defer cache.Close()

	// Test that it works with custom config
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	if value, exists := cache.Get("key1"); !exists || value != "value1" {
		t.Errorf("Expected 'value1', got %v", value)
	}

	stats := cache.Stats()
	if stats.Size != 2 {
		t.Errorf("Expected size 2, got %d", stats.Size)
	}
}

func TestSimpleAPI_Concurrent(t *testing.T) {
	cache := New()
	defer cache.Close()

	// Test concurrent access
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				cache.Set(key, fmt.Sprintf("value_%d_%d", id, j))
				cache.Get(key)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have some items
	stats := cache.Stats()
	if stats.Size == 0 {
		t.Error("Expected some items after concurrent access")
	}
}

func TestConfiguration_Defaults(t *testing.T) {
	config := LoadConfig()

	// Test default configuration (performance)
	if config.CacheSize != 1000000 {
		t.Errorf("Expected default cache size 1000000, got %d", config.CacheSize)
	}

	if config.TTL != 0 {
		t.Errorf("Expected default TTL 0s, got %v", config.TTL)
	}

	if config.EvictionPolicy != "wtinylfu" {
		t.Errorf("Expected default eviction policy 'wtinylfu', got %s", config.EvictionPolicy)
	}

	if config.ShardCount != 128 {
		t.Errorf("Expected default shard count 128, got %d", config.ShardCount)
	}

	if config.EnableCompression != false {
		t.Errorf("Expected default compression false, got %v", config.EnableCompression)
	}

	if config.MaxKeySize != 0 {
		t.Errorf("Expected default max key size 0, got %d", config.MaxKeySize)
	}

	if config.MaxValueSize != 0 {
		t.Errorf("Expected default max value size 0, got %d", config.MaxValueSize)
	}
}

func TestConfiguration_Source(t *testing.T) {
	// Test configuration source detection
	source := GetConfigSource()

	// Should be default since no config files exist in test environment
	if source != "Default configuration" {
		t.Errorf("Expected 'Default configuration', got %s", source)
	}
}

func TestConfiguration_GlobalConfig(t *testing.T) {
	// Test global configuration override
	testConfig := CacheConfig{
		EnableCaching:     true,
		CacheSize:         5000,
		TTL:               5 * time.Minute,
		EnableCompression: true,
		EvictionPolicy:    "lru",
		ShardCount:        16,
		AdmissionPolicy:   "always",
	}

	SetGlobalConfig(testConfig)

	// Load config should return the global config
	loadedConfig := LoadConfig()

	if loadedConfig.CacheSize != 5000 {
		t.Errorf("Expected cache size 5000 from global config, got %d", loadedConfig.CacheSize)
	}

	if loadedConfig.EvictionPolicy != "lru" {
		t.Errorf("Expected eviction policy 'lru' from global config, got %s", loadedConfig.EvictionPolicy)
	}

	// Clear global config for other tests
	SetGlobalConfig(CacheConfig{})
}
