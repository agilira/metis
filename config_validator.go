// config_validator.go: Smart configuration validation and optimization
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package metis

import (
	"fmt"
	"runtime"
	"time"
)

// ConfigValidationResult contains validation results and suggestions
type ConfigValidationResult struct {
	IsValid         bool         `json:"is_valid"`
	Warnings        []string     `json:"warnings"`
	Suggestions     []string     `json:"suggestions"`
	OptimizedConfig *CacheConfig `json:"optimized_config,omitempty"`
}

// ValidateConfig validates a configuration and provides optimization suggestions
func ValidateConfig(config CacheConfig) ConfigValidationResult {
	result := ConfigValidationResult{
		IsValid:     true,
		Warnings:    []string{},
		Suggestions: []string{},
	}

	// Memory usage estimation
	estimatedMemory := estimateMemoryUsage(config)

	// Validate cache size
	if config.CacheSize <= 0 {
		result.IsValid = false
		result.Warnings = append(result.Warnings, "Cache size must be greater than 0")
	} else if config.CacheSize > 10000000 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Large cache size (%d entries) may use ~%.1f GB memory",
			config.CacheSize, float64(estimatedMemory)/(1024*1024*1024)))
	}

	// Validate shard count
	numCPU := runtime.NumCPU()
	if config.ShardCount > numCPU*4 {
		result.Suggestions = append(result.Suggestions, fmt.Sprintf("Consider reducing shard count to %d (4x CPU cores) for optimal performance", numCPU*4))
	} else if config.ShardCount < numCPU && config.CacheSize > 10000 {
		result.Suggestions = append(result.Suggestions, fmt.Sprintf("Consider increasing shard count to %d for better concurrency", numCPU))
	} else if config.CacheSize >= 1000000 && config.ShardCount < numCPU*2 {
		// Special case for very large caches
		result.Suggestions = append(result.Suggestions, fmt.Sprintf("Consider increasing shard count to %d for large caches to improve concurrency and performance", numCPU*2))
	}

	// TTL validation
	if config.TTL > 24*time.Hour {
		result.Suggestions = append(result.Suggestions, "Very long TTL (>24h) may cause memory issues for large datasets")
	}

	// Compression suggestions
	if !config.EnableCompression && estimatedMemory > 100*1024*1024 { // >100MB
		result.Suggestions = append(result.Suggestions, "Consider enabling compression for large datasets to save memory")
	}

	// Generate optimized config
	if len(result.Suggestions) > 0 {
		result.OptimizedConfig = generateOptimizedConfig(config)
	}

	return result
}

// estimateMemoryUsage provides rough memory usage estimation
func estimateMemoryUsage(config CacheConfig) int64 {
	// Rough estimation: 200 bytes per entry (key + value + metadata)
	avgEntrySize := int64(200)
	if config.MaxValueSize > 0 && config.MaxValueSize < 1000 {
		avgEntrySize = int64(config.MaxValueSize) + 50 // value + overhead
	}

	return int64(config.CacheSize) * avgEntrySize
}

// generateOptimizedConfig creates an optimized version of the config
func generateOptimizedConfig(config CacheConfig) *CacheConfig {
	optimized := config
	numCPU := runtime.NumCPU()

	// Optimize shard count
	if optimized.ShardCount > numCPU*4 {
		optimized.ShardCount = numCPU * 2
	} else if optimized.ShardCount < numCPU && config.CacheSize > 10000 {
		optimized.ShardCount = numCPU
	}

	// Enable compression for large datasets
	estimatedMemory := estimateMemoryUsage(config)
	if !optimized.EnableCompression && estimatedMemory > 100*1024*1024 {
		optimized.EnableCompression = true
	}

	return &optimized
}

// GetConfigRecommendation provides configuration recommendations based on use case
func GetConfigRecommendation(useCase string) CacheConfig {
	switch useCase {
	case "development":
		return CacheConfig{
			EnableCaching:   true,
			CacheSize:       1000,
			TTL:             10 * time.Minute,
			ShardCount:      4,
			EvictionPolicy:  "lru", // Simpler for debugging
			AdmissionPolicy: "always",
		}
	case "web-server":
		return CacheConfig{
			EnableCaching:   true,
			CacheSize:       50000,
			TTL:             30 * time.Minute,
			ShardCount:      runtime.NumCPU(),
			EvictionPolicy:  "wtinylfu",
			AdmissionPolicy: "always",
		}
	case "api-gateway":
		return CacheConfig{
			EnableCaching:     true,
			CacheSize:         1000000,
			TTL:               0, // No expiration
			ShardCount:        runtime.NumCPU() * 2,
			EvictionPolicy:    "wtinylfu",
			AdmissionPolicy:   "always",
			EnableCompression: false, // Speed over memory
		}
	case "memory-efficient":
		return CacheConfig{
			EnableCaching:     true,
			CacheSize:         10000,
			TTL:               1 * time.Hour,
			ShardCount:        runtime.NumCPU(),
			EvictionPolicy:    "wtinylfu",
			AdmissionPolicy:   "always",
			EnableCompression: true,
			MaxValueSize:      524288, // 512KB limit
		}
	default:
		return getDefaultConfig()
	}
}
