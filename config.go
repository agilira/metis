// config.go: Configuration system for Metis strategic caching library
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package metis

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// SimpleConfig represents the complete configuration from metis.json
type SimpleConfig struct {
	CacheSize         int    `json:"cache_size"`
	TTL               string `json:"ttl"`
	CleanupInterval   string `json:"cleanup_interval"`
	EnableCompression bool   `json:"enable_compression"`
	EvictionPolicy    string `json:"eviction_policy"`
	ShardCount        int    `json:"shard_count"`
	AdmissionPolicy   string `json:"admission_policy"`
	MaxKeySize        int    `json:"max_key_size"`
	MaxValueSize      int    `json:"max_value_size"`
	MaxShardSize      int    `json:"max_shard_size"`
}

// Global configuration state
var (
	globalConfig *CacheConfig
	configMutex  sync.RWMutex
)

// SetGlobalConfig sets the global configuration for power users
// This should be called in init() function of a metis_config.go file
func SetGlobalConfig(config CacheConfig) {
	configMutex.Lock()
	defer configMutex.Unlock()
	globalConfig = &config
}

// GetGlobalConfig returns the current global configuration
func GetGlobalConfig() *CacheConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return globalConfig
}

// loadConfig loads configuration with priority: Go config > JSON config > defaults
func loadConfig() CacheConfig {
	// Check if power user has set global config via Go file
	if config := GetGlobalConfig(); config != nil {
		return *config
	}

	// Try to load from metis.json
	if config, err := loadJSONConfig(); err == nil {
		return config
	}

	// Return sensible defaults
	return getDefaultConfig()
}

// loadJSONConfig loads configuration from metis.json
func loadJSONConfig() (CacheConfig, error) {
	configPath := findConfigFile()
	if configPath == "" {
		return CacheConfig{}, fmt.Errorf("metis.json not found")
	}

	if filepath.Base(configPath) != "metis.json" || strings.Contains(configPath, "..") {
		return CacheConfig{}, fmt.Errorf("invalid config file path: %s", configPath)
	}
	// nosec G304 - configPath is validated above to prevent path traversal
	data, err := os.ReadFile(configPath)
	if err != nil {
		return CacheConfig{}, fmt.Errorf("failed to read %s: %v", configPath, err)
	}

	var simpleConfig SimpleConfig
	if err := json.Unmarshal(data, &simpleConfig); err != nil {
		return CacheConfig{}, fmt.Errorf("failed to parse %s: %v", configPath, err)
	}

	// Convert simple config to full config
	config := getDefaultConfig()

	// Apply simple config values
	if simpleConfig.CacheSize > 0 {
		config.CacheSize = simpleConfig.CacheSize
	}

	if simpleConfig.TTL != "" {
		if ttl, err := time.ParseDuration(simpleConfig.TTL); err == nil {
			config.TTL = ttl
		} else {
			return CacheConfig{}, fmt.Errorf("invalid TTL format in %s: %v", configPath, err)
		}
	}

	if simpleConfig.CleanupInterval != "" {
		if cleanupInterval, err := time.ParseDuration(simpleConfig.CleanupInterval); err == nil {
			config.CleanupInterval = cleanupInterval
		} else {
			return CacheConfig{}, fmt.Errorf("invalid cleanup_interval format in %s: %v", configPath, err)
		}
	}

	// Apply boolean and string configurations
	config.EnableCompression = simpleConfig.EnableCompression

	if simpleConfig.EvictionPolicy != "" {
		config.EvictionPolicy = simpleConfig.EvictionPolicy
	}

	if simpleConfig.ShardCount > 0 {
		config.ShardCount = simpleConfig.ShardCount
	}

	if simpleConfig.AdmissionPolicy != "" {
		config.AdmissionPolicy = simpleConfig.AdmissionPolicy
	}

	if simpleConfig.MaxKeySize > 0 {
		config.MaxKeySize = simpleConfig.MaxKeySize
	}

	if simpleConfig.MaxValueSize > 0 {
		config.MaxValueSize = simpleConfig.MaxValueSize
	}

	if simpleConfig.MaxShardSize > 0 {
		config.MaxShardSize = simpleConfig.MaxShardSize
	}

	return config, nil
}

// findConfigFile searches for metis.json in current and parent directories
func findConfigFile() string {
	// Start from current directory
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Search up to 5 parent directories
	for i := 0; i < 5; i++ {
		configPath := filepath.Join(dir, "metis.json")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			break // Reached root
		}
		dir = parent
	}

	return ""
}

// getDefaultConfig returns maximum performance configuration
func getDefaultConfig() CacheConfig {
	return CacheConfig{
		EnableCaching:     true,
		CacheSize:         1000000, // 1M entries for maximum performance
		TTL:               0,       // No TTL for maximum performance
		CleanupInterval:   0,       // No cleanup for maximum performance
		EnableCompression: false,   // No compression for maximum performance
		EvictionPolicy:    "wtinylfu",
		ShardCount:        128,      // Maximum shards for maximum concurrency
		AdmissionPolicy:   "always", // Always admit for maximum performance
		MaxKeySize:        0,        // No key size limit for maximum performance
		MaxValueSize:      0,        // No value size limit for maximum performance
		MaxShardSize:      0,        // No shard size limit for maximum performance
	}
}

// LoadConfig loads the current configuration (for debugging/inspection)
func LoadConfig() CacheConfig {
	return loadConfig()
}

// GetConfigSource returns information about the configuration source
func GetConfigSource() string {
	if GetGlobalConfig() != nil {
		return "Go configuration (metis_config.go)"
	}

	if findConfigFile() != "" {
		return "JSON configuration (metis.json)"
	}

	return "Default configuration"
}
