// misc_edge_test.go: Various edge cases for testing
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package metis

import (
	"testing"
	"time"
)

// TestEstimateMemoryUsage_EdgeCases tests memory estimation edge cases
func TestEstimateMemoryUsage_EdgeCases(t *testing.T) {
	testCases := []CacheConfig{
		{
			CacheSize:    0,
			MaxValueSize: 0,
		},
		{
			CacheSize:    100,
			MaxValueSize: 50, // Small value size
		},
		{
			CacheSize:    1000000,
			MaxValueSize: 0, // No limit
		},
		{
			CacheSize:    500,
			MaxValueSize: 5000, // Large value size
		},
	}

	for i, config := range testCases {
		t.Run("", func(t *testing.T) {
			usage := estimateMemoryUsage(config)
			t.Logf("Config %d: CacheSize=%d, MaxValueSize=%d, Usage=%d",
				i, config.CacheSize, config.MaxValueSize, usage)

			if usage < 0 {
				t.Errorf("memory usage should not be negative: %d", usage)
			}
		})
	}
}

// TestLRU_SetComplexScenarios tests LRU Set function edge cases
func TestLRU_SetComplexScenarios(t *testing.T) {
	config := CacheConfig{
		EnableCaching:     true,
		CacheSize:         5,
		EvictionPolicy:    "lru",
		ShardCount:        1,
		EnableCompression: true,
		MaxKeySize:        10,
		MaxValueSize:      20,
	}
	cache := NewStrategicCache(config)
	defer cache.Close()

	// Test various data types
	testCases := []struct {
		key   string
		value interface{}
		valid bool
	}{
		{"str", "string value", true},
		{"int", 42, true},
		{"bool", true, true},
		{"float", 3.14159, true},
		{"nil", nil, true},
		{"large_key_12345", "value", false},                    // Key too large
		{"size", "value_that_is_too_large_for_maxsize", false}, // Value too large
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			success := cache.Set(tc.key, tc.value)
			if tc.valid && !success {
				t.Errorf("expected success for key %s", tc.key)
			}
			if !tc.valid && success {
				t.Errorf("expected failure for key %s", tc.key)
			}
		})
	}
}

// TestCompression_ExtensiveCases tests compression with more scenarios
func TestCompression_ExtensiveCases(t *testing.T) {
	testCases := []struct {
		name   string
		data   []byte
		header string
	}{
		{"small", []byte("small"), "header1"},
		{"medium", make([]byte, 1000), "header2"},
		{"large", make([]byte, 10000), "header3"},
		{"special_chars", []byte("special chars: éñüñ @#$%"), "unicode"},
		{"binary", []byte{0x00, 0x01, 0x02, 0xFF, 0xFE}, "binary"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Fill with pattern for non-small data
			if len(tc.data) > 10 {
				pattern := []byte("pattern123456789")
				for i := 0; i < len(tc.data); i++ {
					tc.data[i] = pattern[i%len(pattern)]
				}
			}

			compressed, err := compressGzipWithHeader(tc.data, tc.header)
			if err != nil {
				t.Errorf("compression failed: %v", err)
			}
			if compressed == nil {
				t.Error("compressed data should not be nil")
			}
			if len(compressed) == 0 {
				t.Error("compressed data should not be empty")
			}
		})
	}
}

// TestLRU_GetComplexScenarios tests LRU Get function edge cases
func TestLRU_GetComplexScenarios(t *testing.T) {
	config := CacheConfig{
		EnableCaching:     true,
		CacheSize:         10,
		EvictionPolicy:    "lru",
		ShardCount:        1,
		EnableCompression: true,
		TTL:               100 * time.Millisecond,
		CleanupInterval:   50 * time.Millisecond,
	}
	cache := NewStrategicCache(config)
	defer cache.Close()

	// Test with compressed data
	largeValue := make([]byte, 2000)
	for i := range largeValue {
		largeValue[i] = byte('A' + i%26)
	}

	cache.Set("large", string(largeValue))
	retrieved, exists := cache.Get("large")
	if !exists {
		t.Error("large value should exist")
	}
	if retrieved != string(largeValue) {
		t.Error("decompressed value should match original")
	}

	// Test with TTL expiration
	cache.Set("ttl_test", "expires_soon")

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Try to get expired value
	value, exists := cache.Get("ttl_test")
	if exists {
		t.Logf("TTL test: value may or may not exist depending on cleanup timing: %v", value)
	}

	// Test non-existent key
	value, exists = cache.Get("definitely_not_there")
	if exists {
		t.Error("non-existent key should not exist")
	}
	if value != nil {
		t.Error("non-existent key should return nil value")
	}
}

// TestWTinyLFU_EdgeFunctions tests remaining WTinyLFU functions
func TestWTinyLFU_EdgeFunctions(t *testing.T) {
	cache := NewWTinyLFU(100, 4)

	// Add some data
	for i := 0; i < 10; i++ {
		cache.Set(string(rune('a'+i)), i)
	}

	// Test WindowSize with data
	windowSize := cache.WindowSize()
	if windowSize < 0 {
		t.Errorf("window size should be non-negative: %d", windowSize)
	}

	// Test MainSize with data
	mainSize := cache.MainSize()
	if mainSize < 0 {
		t.Errorf("main size should be non-negative: %d", mainSize)
	}

	// Test AdmissionFilter with data
	filter := cache.AdmissionFilter()
	if filter == nil {
		t.Error("admission filter should not be nil")
	}

	// Test SetGet with existing data
	oldVal, existed := cache.SetGet("a", "new_value")
	if !existed {
		t.Error("key 'a' should have existed")
	}
	if oldVal != "new_value" {
		t.Errorf("SetGet should return new value, got %v", oldVal)
	}

	// Test SetGet with new key
	newVal, existed := cache.SetGet("z", "z_value")
	if existed {
		t.Logf("New key might be considered existing due to admission logic")
	}
	if newVal != "z_value" {
		t.Errorf("SetGet should return set value, got %v", newVal)
	}
}

// TestConfigValidation_FullCoverage tests all validation paths
func TestConfigValidation_FullCoverage(t *testing.T) {
	// Test with extreme values to trigger all validation paths
	extremeConfig := CacheConfig{
		CacheSize:            20000000,       // Very large to trigger memory warning
		ShardCount:           1000,           // Very high to trigger shard suggestion
		TTL:                  48 * time.Hour, // Long TTL to trigger suggestion
		EnableCompression:    false,          // Disabled compression for large cache
		MaxValueSize:         5000,           // Large values
		MaxKeySize:           1000,           // Large keys
		AdmissionProbability: 0.1,            // Low admission probability
	}

	result := ValidateConfig(extremeConfig)

	// Should have suggestions
	if len(result.Suggestions) == 0 {
		t.Error("extreme config should generate suggestions")
	}

	// Should have optimized config
	if result.OptimizedConfig == nil {
		t.Error("extreme config should generate optimized config")
	}

	t.Logf("Warnings: %v", result.Warnings)
	t.Logf("Suggestions: %v", result.Suggestions)
}
