// additional_unit_test.go: More tests for untested functions
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package metis

import (
	"testing"
)

// TestWTinyLFU_SetGet tests WTinyLFU SetGet operation
func TestWTinyLFU_SetGet(t *testing.T) {
	cache := NewWTinyLFU(100, 4)

	// Test setting value - should return the value just set
	value, _ := cache.SetGet("key1", "value1")
	if value != "value1" {
		t.Errorf("should return the value just set 'value1', got %v", value)
	}

	// Test updating existing value
	value, _ = cache.SetGet("key1", "value2")
	if value != "value2" {
		t.Errorf("should return the new value 'value2', got %v", value)
	}

	// Verify the final value is correct
	finalValue, exists := cache.Get("key1")
	if !exists {
		t.Error("key1 should exist")
	}
	if finalValue != "value2" {
		t.Errorf("final value should be 'value2', got %v", finalValue)
	}
}

// TestWTinyLFU_Exists tests WTinyLFU Exists operation
func TestWTinyLFU_Exists(t *testing.T) {
	cache := NewWTinyLFU(100, 4)

	// Key should not exist initially
	if cache.Exists("key1") {
		t.Error("key1 should not exist initially")
	}

	// Set key and check existence
	cache.Set("key1", "value1")
	if !cache.Exists("key1") {
		t.Error("key1 should exist after setting")
	}

	// Delete key and check existence
	cache.Delete("key1")
	if cache.Exists("key1") {
		t.Error("key1 should not exist after deletion")
	}
}

// TestWTinyLFU_HealthCheck tests WTinyLFU health check
func TestWTinyLFU_HealthCheck(t *testing.T) {
	cache := NewWTinyLFU(100, 4)

	health := cache.HealthCheck()
	if health == nil {
		t.Error("health check should not return nil")
	}

	// Should contain basic health information
	if len(health) == 0 {
		t.Error("health check should contain information")
	}
}

// TestConfig_GenerateOptimizedConfig tests generateOptimizedConfig function
func TestConfig_GenerateOptimizedConfig(t *testing.T) {
	testConfigs := []CacheConfig{
		{
			CacheSize:            1000,
			EvictionPolicy:       "lru",
			ShardCount:           1,
			EnableCompression:    false,
			AdmissionProbability: 0.5,
		},
		{
			CacheSize:            100000,
			EvictionPolicy:       "wtinylfu",
			ShardCount:           64,
			EnableCompression:    true,
			AdmissionProbability: 0.8,
		},
	}

	for i, config := range testConfigs {
		t.Run("", func(t *testing.T) {
			optimized := generateOptimizedConfig(config)

			// Verify optimized config has reasonable values
			if optimized.CacheSize <= 0 {
				t.Errorf("test %d: optimized cache size should be positive, got %d", i, optimized.CacheSize)
			}

			if optimized.ShardCount <= 0 {
				t.Errorf("test %d: optimized shard count should be positive, got %d", i, optimized.ShardCount)
			}

			if optimized.EvictionPolicy == "" {
				t.Errorf("test %d: optimized eviction policy should not be empty", i)
			}
		})
	}
}
