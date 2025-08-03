// config_validator_test.go: Comprehensive tests for config validator functions
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package metis

import (
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestGetConfigRecommendation_AllUseCases tests all supported use cases
func TestGetConfigRecommendation_AllUseCases(t *testing.T) {
	testCases := []struct {
		name             string
		useCase          string
		expectedSize     int
		expectedPolicy   string
		expectedShards   int
		checkTTL         bool
		expectedTTL      time.Duration
		checkCompress    bool
		expectedCompress bool
	}{
		{
			name:           "Development case",
			useCase:        "development",
			expectedSize:   1000,
			expectedPolicy: "lru",
			expectedShards: 4,
			checkTTL:       true,
			expectedTTL:    10 * time.Minute,
			checkCompress:  false,
		},
		{
			name:           "Web server case",
			useCase:        "web-server",
			expectedSize:   50000,
			expectedPolicy: "wtinylfu",
			expectedShards: runtime.NumCPU(),
			checkTTL:       true,
			expectedTTL:    30 * time.Minute,
			checkCompress:  false,
		},
		{
			name:             "API gateway case",
			useCase:          "api-gateway",
			expectedSize:     1000000,
			expectedPolicy:   "wtinylfu",
			expectedShards:   runtime.NumCPU() * 2,
			checkTTL:         true,
			expectedTTL:      0, // No expiration
			checkCompress:    true,
			expectedCompress: false, // Speed over memory
		},
		{
			name:             "Memory efficient case",
			useCase:          "memory-efficient",
			expectedSize:     10000,
			expectedPolicy:   "wtinylfu",
			expectedShards:   runtime.NumCPU(),
			checkTTL:         true,
			expectedTTL:      1 * time.Hour,
			checkCompress:    true,
			expectedCompress: true,
		},
		{
			name:           "Unknown use case - defaults to default config",
			useCase:        "unknown-use-case",
			expectedSize:   1000000, // Default config has 1M entries
			expectedPolicy: "wtinylfu",
			expectedShards: 128,   // Default config has 128 shards
			checkTTL:       false, // Default config varies
			checkCompress:  false,
		},
		{
			name:           "Empty string use case - defaults to default config",
			useCase:        "",
			expectedSize:   1000000, // Default config has 1M entries
			expectedPolicy: "wtinylfu",
			expectedShards: 128, // Default config has 128 shards
			checkTTL:       false,
			checkCompress:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := GetConfigRecommendation(tc.useCase)

			// Test cache size
			if config.CacheSize != tc.expectedSize {
				t.Errorf("expected cache size %d, got %d", tc.expectedSize, config.CacheSize)
			}

			// Test eviction policy
			if config.EvictionPolicy != tc.expectedPolicy {
				t.Errorf("expected eviction policy %s, got %s", tc.expectedPolicy, config.EvictionPolicy)
			}

			// Test shard count
			if config.ShardCount != tc.expectedShards {
				t.Errorf("expected shard count %d, got %d", tc.expectedShards, config.ShardCount)
			}

			// Test TTL if required
			if tc.checkTTL && config.TTL != tc.expectedTTL {
				t.Errorf("expected TTL %v, got %v", tc.expectedTTL, config.TTL)
			}

			// Test compression if required
			if tc.checkCompress && config.EnableCompression != tc.expectedCompress {
				t.Errorf("expected compression %v, got %v", tc.expectedCompress, config.EnableCompression)
			}

			// Test that config is enabled
			if !config.EnableCaching {
				t.Error("expected caching to be enabled")
			}

			// Test that admission policy is set
			if config.AdmissionPolicy == "" {
				t.Error("admission policy should not be empty")
			}

			// Test the config works by creating a cache
			cache := NewWithConfig(config)
			defer cache.Close()

			// Basic functionality test
			cache.Set("test", "value")
			if value, exists := cache.Get("test"); !exists || value != "value" {
				t.Errorf("recommended config for %s should work correctly", tc.useCase)
			}
		})
	}
}

// TestValidateConfig_EdgeCases tests edge cases for config validation
func TestValidateConfig_EdgeCases(t *testing.T) {
	testCases := []struct {
		name              string
		config            CacheConfig
		expectValid       bool
		expectWarnings    int
		expectSuggestions int
		checkOptimized    bool
	}{
		{
			name: "Zero cache size - invalid",
			config: CacheConfig{
				CacheSize:  0,
				ShardCount: 4,
			},
			expectValid:       false,
			expectWarnings:    1,
			expectSuggestions: 0,
			checkOptimized:    false,
		},
		{
			name: "Negative cache size - invalid",
			config: CacheConfig{
				CacheSize:  -100,
				ShardCount: 4,
			},
			expectValid:       false,
			expectWarnings:    1,
			expectSuggestions: 0,
			checkOptimized:    false,
		},
		{
			name: "Very large cache size - warning",
			config: CacheConfig{
				CacheSize:  20000000, // 20M entries
				ShardCount: 4,
				TTL:        time.Hour,
			},
			expectValid:       true,
			expectWarnings:    1,  // Memory warning
			expectSuggestions: -1, // Variable based on runtime.NumCPU()
			checkOptimized:    true,
		},
		{
			name: "Too many shards - suggestion",
			config: CacheConfig{
				CacheSize:  10000,
				ShardCount: runtime.NumCPU() * 8, // 8x CPU cores
				TTL:        time.Hour,
			},
			expectValid:       true,
			expectWarnings:    0,
			expectSuggestions: 1, // Reduce shards
			checkOptimized:    true,
		},
		{
			name: "Too few shards for large cache - suggestion",
			config: CacheConfig{
				CacheSize:  50000, // Large cache
				ShardCount: 1,     // Too few shards
				TTL:        time.Hour,
			},
			expectValid:       true,
			expectWarnings:    0,
			expectSuggestions: 1, // Increase shards
			checkOptimized:    true,
		},
		{
			name: "Very long TTL - suggestion",
			config: CacheConfig{
				CacheSize:  10000,
				ShardCount: 4,
				TTL:        48 * time.Hour, // Very long TTL
			},
			expectValid:       true,
			expectWarnings:    0,
			expectSuggestions: 1, // TTL warning
			checkOptimized:    false,
		},
		{
			name: "Large dataset without compression - suggestion",
			config: CacheConfig{
				CacheSize:         1000000, // Large cache
				ShardCount:        4,
				TTL:               time.Hour,
				EnableCompression: false,
				MaxValueSize:      0, // No limit, will use default 200 bytes estimation
			},
			expectValid:       true,
			expectWarnings:    0,
			expectSuggestions: -1, // Variable based on runtime.NumCPU()
			checkOptimized:    true,
		},
		{
			name: "Perfect config - no warnings or suggestions",
			config: CacheConfig{
				CacheSize:         10000,
				ShardCount:        runtime.NumCPU(),
				TTL:               time.Hour,
				EnableCompression: false,
			},
			expectValid:       true,
			expectWarnings:    0,
			expectSuggestions: 0,
			checkOptimized:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidateConfig(tc.config)

			// Check validity
			if result.IsValid != tc.expectValid {
				t.Errorf("expected valid %v, got %v", tc.expectValid, result.IsValid)
			}

			// Check warnings count
			if len(result.Warnings) != tc.expectWarnings {
				t.Errorf("expected %d warnings, got %d: %v", tc.expectWarnings, len(result.Warnings), result.Warnings)
			}

			// Check suggestions count
			if tc.expectSuggestions == -1 {
				// Special case for tests with variable suggestion counts based on runtime.NumCPU()
				if tc.name == "Large dataset without compression - suggestion" || tc.name == "Very large cache size - warning" {
					t.Logf("runtime.NumCPU() = %d", runtime.NumCPU())
					for i, s := range result.Suggestions {
						t.Logf("Suggestion %d: %s", i, s)
					}

					// Must have at least compression suggestion for large datasets
					hasCompressionSuggestion := false
					for _, s := range result.Suggestions {
						if strings.Contains(s, "compression") {
							hasCompressionSuggestion = true
							break
						}
					}
					if !hasCompressionSuggestion {
						t.Errorf("expected compression suggestion, got: %v", result.Suggestions)
					}

					// Should have at least 1 suggestion (compression)
					if len(result.Suggestions) < 1 {
						t.Errorf("expected at least 1 suggestion, got %d: %v", len(result.Suggestions), result.Suggestions)
					}
				}
			} else if len(result.Suggestions) != tc.expectSuggestions {
				t.Errorf("expected %d suggestions, got %d: %v", tc.expectSuggestions, len(result.Suggestions), result.Suggestions)
			}

			// Check optimized config creation
			if tc.checkOptimized {
				if result.OptimizedConfig == nil {
					t.Error("expected optimized config to be generated")
				}
			}
		})
	}
}

// TestEstimateMemoryUsage_ValidatorCases tests memory estimation with various value sizes for validator
func TestEstimateMemoryUsage_ValidatorCases(t *testing.T) {
	testCases := []struct {
		name        string
		config      CacheConfig
		expectedMin int64
		expectedMax int64
	}{
		{
			name: "Default estimation - no max value size",
			config: CacheConfig{
				CacheSize:    1000,
				MaxValueSize: 0, // Will use default 200 bytes
			},
			expectedMin: 190000, // ~190KB (1000 * 190)
			expectedMax: 210000, // ~210KB (1000 * 210)
		},
		{
			name: "Small max value size",
			config: CacheConfig{
				CacheSize:    1000,
				MaxValueSize: 100, // Small values
			},
			expectedMin: 140000, // 1000 * (100 + 40)
			expectedMax: 160000, // 1000 * (100 + 60)
		},
		{
			name: "Large max value size - should use default",
			config: CacheConfig{
				CacheSize:    1000,
				MaxValueSize: 2000, // Large values, will use default 200
			},
			expectedMin: 190000, // Back to default estimation
			expectedMax: 210000,
		},
		{
			name: "Zero cache size",
			config: CacheConfig{
				CacheSize:    0,
				MaxValueSize: 100,
			},
			expectedMin: 0,
			expectedMax: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			estimated := estimateMemoryUsage(tc.config)

			if estimated < tc.expectedMin || estimated > tc.expectedMax {
				t.Errorf("expected memory usage between %d and %d, got %d", tc.expectedMin, tc.expectedMax, estimated)
			}
		})
	}
}

// TestGenerateOptimizedConfig_EdgeCases tests config optimization scenarios
func TestGenerateOptimizedConfig_EdgeCases(t *testing.T) {
	testCases := []struct {
		name            string
		input           CacheConfig
		expectOptShards bool
		expectCompress  bool
	}{
		{
			name: "Too many shards - should optimize",
			input: CacheConfig{
				CacheSize:         10000,
				ShardCount:        runtime.NumCPU() * 8,
				EnableCompression: false,
			},
			expectOptShards: true,
			expectCompress:  false,
		},
		{
			name: "Too few shards for large cache - should optimize",
			input: CacheConfig{
				CacheSize:         50000, // Large cache
				ShardCount:        1,     // Too few
				EnableCompression: false,
			},
			expectOptShards: true,
			expectCompress:  false,
		},
		{
			name: "Large dataset without compression - should enable compression",
			input: CacheConfig{
				CacheSize:         1000000, // Very large cache
				ShardCount:        runtime.NumCPU(),
				EnableCompression: false,
			},
			expectOptShards: false,
			expectCompress:  true,
		},
		{
			name: "Small cache - no optimization needed",
			input: CacheConfig{
				CacheSize:         1000,
				ShardCount:        runtime.NumCPU(),
				EnableCompression: false,
			},
			expectOptShards: false,
			expectCompress:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			optimized := generateOptimizedConfig(tc.input)

			// Check shard optimization
			if tc.expectOptShards {
				if optimized.ShardCount == tc.input.ShardCount {
					t.Error("expected shard count to be optimized")
				}
				// Should be reasonable
				if optimized.ShardCount <= 0 || optimized.ShardCount > runtime.NumCPU()*4 {
					t.Errorf("optimized shard count %d is not reasonable", optimized.ShardCount)
				}
			}

			// Check compression optimization
			if tc.expectCompress {
				if !optimized.EnableCompression {
					t.Error("expected compression to be enabled in optimized config")
				}
			}

			// Optimized config should always be valid
			result := ValidateConfig(*optimized)
			if !result.IsValid {
				t.Error("optimized config should be valid")
			}
		})
	}
}
