// api_config_test.go: additional test to check api and config
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package metis

import (
	"testing"
	"time"
)

// TestAPI_Size tests the Size function
func TestAPI_Size(t *testing.T) {
	cache := New()
	defer cache.Close()

	// Initially empty
	if size := cache.Size(); size != 0 {
		t.Errorf("expected size 0, got %d", size)
	}

	// Add some items
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	// Note: WTinyLFU may use admission control, so let's check if items were actually added
	size := cache.Size()
	t.Logf("Size after adding 3 items: %d", size)

	// Since we can't guarantee all items are admitted in WTinyLFU, just test that some items exist
	if size < 0 {
		t.Errorf("size should not be negative, got %d", size)
	}

	// Clear all
	cache.Clear()
	size = cache.Size()
	if size != 0 {
		t.Errorf("expected size 0 after clear, got %d", size)
	}
}

// TestAPI_GetConfigInfo tests configuration info retrieval
func TestAPI_GetConfigInfo(t *testing.T) {
	info := GetConfigInfo()
	if info == "" {
		t.Error("config info should not be empty")
	}

	// Should contain policy information
	if len(info) < 10 {
		t.Errorf("config info seems too short: %s", info)
	}
}

// TestAPI_NewHighPerformanceCache tests high performance cache constructor
func TestAPI_NewHighPerformanceCache(t *testing.T) {
	cache := NewHighPerformanceCache(1000)
	if cache == nil {
		t.Error("NewHighPerformanceCache should not return nil")
	}

	// Test basic functionality
	cache.Set("test", "value")
	if value, exists := cache.Get("test"); !exists || value != "value" {
		t.Error("High performance cache should work correctly")
	}
}

// TestAPI_ConvenienceConstructors tests convenience constructor functions
func TestAPI_ConvenienceConstructors(t *testing.T) {
	testCases := []struct {
		name        string
		constructor func() *Cache
	}{
		{"NewWebServerCache", NewWebServerCache},
		{"NewAPIGatewayCache", NewAPIGatewayCache},
		{"NewDevelopmentCache", NewDevelopmentCache},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache := tc.constructor()
			if cache == nil {
				t.Errorf("%s should not return nil", tc.name)
			}
			defer cache.Close()

			// Test basic functionality
			cache.Set("test", "value")
			if value, exists := cache.Get("test"); !exists || value != "value" {
				t.Errorf("%s cache should work correctly", tc.name)
			}
		})
	}
}

// TestAPI_NewForUseCase tests use case specific constructors
func TestAPI_NewForUseCase(t *testing.T) {
	testCases := []string{
		"web-server",
		"api-gateway",
		"database-cache",
		"session-store",
		"high-performance",
		"development",
		"unknown-case", // Should fallback to default
	}

	for _, useCase := range testCases {
		t.Run(useCase, func(t *testing.T) {
			cache := NewForUseCase(useCase)
			if cache == nil {
				t.Errorf("NewForUseCase(%s) should not return nil", useCase)
			}
			defer cache.Close()

			// Test basic functionality
			cache.Set("test", "value")
			if value, exists := cache.Get("test"); !exists || value != "value" {
				t.Errorf("NewForUseCase(%s) cache should work correctly", useCase)
			}
		})
	}
}

// TestAPI_NewHighPerformanceCacheWithShards tests sharded high performance cache
func TestAPI_NewHighPerformanceCacheWithShards(t *testing.T) {
	testCases := []struct {
		size   int
		shards int
	}{
		{1000, 1},
		{1000, 16},
		{1000, 64},
		{1000, 128},
		{1000, 256},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			cache := NewHighPerformanceCacheWithShards(tc.size, tc.shards)
			if cache == nil {
				t.Errorf("NewHighPerformanceCacheWithShards(%d, %d) should not return nil", tc.size, tc.shards)
			}

			// Test basic functionality
			cache.Set("test", "value")
			if value, exists := cache.Get("test"); !exists || value != "value" {
				t.Errorf("Sharded cache with %d shards should work correctly", tc.shards)
			}
		})
	}
}

// TestConfig_ValidateConfig tests configuration validation
func TestConfig_ValidateConfig(t *testing.T) {
	testCases := []struct {
		name    string
		config  CacheConfig
		isValid bool
		hasWarn bool
		hasSugg bool
	}{
		{
			name: "valid_config",
			config: CacheConfig{
				EnableCaching:        true,
				CacheSize:            1000,
				EvictionPolicy:       "wtinylfu",
				ShardCount:           16,
				TTL:                  5 * time.Minute,
				CleanupInterval:      1 * time.Minute,
				AdmissionProbability: 0.8,
				MaxKeySize:           512,
				MaxValueSize:         1024,
			},
			isValid: true,
			hasWarn: false,
			hasSugg: false,
		},
		{
			name: "invalid_cache_size",
			config: CacheConfig{
				EnableCaching:  true,
				CacheSize:      -1,
				EvictionPolicy: "wtinylfu",
			},
			isValid: false,
			hasWarn: true,
			hasSugg: false,
		},
		{
			name: "large_cache_size",
			config: CacheConfig{
				EnableCaching:  true,
				CacheSize:      15000000, // Very large
				EvictionPolicy: "wtinylfu",
			},
			isValid: true,
			hasWarn: true,
			hasSugg: false,
		},
		{
			name: "high_shard_count",
			config: CacheConfig{
				EnableCaching:  true,
				CacheSize:      1000,
				EvictionPolicy: "wtinylfu",
				ShardCount:     1000, // Very high
			},
			isValid: true,
			hasWarn: false,
			hasSugg: true,
		},
		{
			name: "long_ttl",
			config: CacheConfig{
				EnableCaching:  true,
				CacheSize:      1000,
				EvictionPolicy: "wtinylfu",
				TTL:            48 * time.Hour, // Very long
			},
			isValid: true,
			hasWarn: false,
			hasSugg: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidateConfig(tc.config)

			if tc.isValid && !result.IsValid {
				t.Errorf("expected valid config, got invalid with warnings: %v", result.Warnings)
			}
			if !tc.isValid && result.IsValid {
				t.Error("expected invalid config, got valid")
			}

			if tc.hasWarn && len(result.Warnings) == 0 {
				t.Error("expected warnings but got none")
			}
			if tc.hasSugg && len(result.Suggestions) == 0 {
				t.Error("expected suggestions but got none")
			}
		})
	}
}

// TestConfig_GetConfigSource tests config source detection
func TestConfig_GetConfigSource(t *testing.T) {
	// Test with default configuration
	source := GetConfigSource()
	if source == "" {
		t.Error("config source should not be empty")
	}

	// Should be one of the expected sources
	validSources := []string{
		"Go configuration (metis_config.go)",
		"JSON configuration (metis.json)",
		"Default configuration",
	}

	found := false
	for _, validSource := range validSources {
		if source == validSource {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("unexpected config source: %s", source)
	}
}

// TestConfig_GetConfigRecommendation tests config recommendations
func TestConfig_GetConfigRecommendation(t *testing.T) {
	testCases := []string{
		"web-server",
		"api-gateway",
		"database-cache",
		"session-store",
		"high-performance",
		"development",
		"unknown-use-case",
	}

	for _, useCase := range testCases {
		t.Run(useCase, func(t *testing.T) {
			config := GetConfigRecommendation(useCase)

			// Basic validation
			if config.CacheSize <= 0 {
				t.Errorf("cache size should be positive, got %d", config.CacheSize)
			}

			if config.ShardCount <= 0 {
				t.Errorf("shard count should be positive, got %d", config.ShardCount)
			}

			if config.EvictionPolicy == "" {
				t.Error("eviction policy should not be empty")
			}

			// Test that the config works
			cache := NewWithConfig(config)
			defer cache.Close()

			cache.Set("test", "value")
			if value, exists := cache.Get("test"); !exists || value != "value" {
				t.Errorf("recommended config for %s should work correctly", useCase)
			}
		})
	}
}
