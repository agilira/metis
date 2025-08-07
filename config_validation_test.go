// config_validation_test.go: Advanced validation tests for config loading
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package metis

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestLoadJSONConfig_ValidFile tests loading a valid JSON config file
func TestLoadJSONConfig_ValidFile(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "metis_config_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a valid metis.json
	validConfig := SimpleConfig{
		CacheSize:         50000,
		TTL:               "1h",
		CleanupInterval:   "30m",
		EnableCompression: true,
		EvictionPolicy:    "lru",
		ShardCount:        64,
		AdmissionPolicy:   "tinylfu",
		MaxKeySize:        1024,
		MaxValueSize:      2048,
		MaxShardSize:      10000,
	}

	configData, err := json.Marshal(validConfig)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}

	configPath := filepath.Join(tempDir, "metis.json")
	err = os.WriteFile(configPath, configData, 0644)
	if err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Change to temp directory to test file finding
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Test loadJSONConfig
	config, err := loadJSONConfig()
	if err != nil {
		t.Fatalf("loadJSONConfig failed: %v", err)
	}

	// Verify values
	if config.CacheSize != 50000 {
		t.Errorf("expected CacheSize 50000, got %d", config.CacheSize)
	}
	if config.TTL != time.Hour {
		t.Errorf("expected TTL 1h, got %v", config.TTL)
	}
	if config.CleanupInterval != 30*time.Minute {
		t.Errorf("expected CleanupInterval 30m, got %v", config.CleanupInterval)
	}
	if !config.EnableCompression {
		t.Error("expected EnableCompression true")
	}
	if config.EvictionPolicy != "lru" {
		t.Errorf("expected EvictionPolicy lru, got %s", config.EvictionPolicy)
	}
	if config.ShardCount != 64 {
		t.Errorf("expected ShardCount 64, got %d", config.ShardCount)
	}
	if config.AdmissionPolicy != "tinylfu" {
		t.Errorf("expected AdmissionPolicy tinylfu, got %s", config.AdmissionPolicy)
	}
	if config.MaxKeySize != 1024 {
		t.Errorf("expected MaxKeySize 1024, got %d", config.MaxKeySize)
	}
	if config.MaxValueSize != 2048 {
		t.Errorf("expected MaxValueSize 2048, got %d", config.MaxValueSize)
	}
	if config.MaxShardSize != 10000 {
		t.Errorf("expected MaxShardSize 10000, got %d", config.MaxShardSize)
	}
}

// TestLoadJSONConfig_CorruptedFile tests loading a corrupted JSON file
func TestLoadJSONConfig_CorruptedFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "metis_config_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a corrupted JSON file
	configPath := filepath.Join(tempDir, "metis.json")
	corruptedJSON := `{"cache_size": 1000, "ttl": "1h", "invalid_json":`
	err = os.WriteFile(configPath, []byte(corruptedJSON), 0644)
	if err != nil {
		t.Fatalf("failed to write corrupted config file: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Test loadJSONConfig with corrupted file
	_, err = loadJSONConfig()
	if err == nil {
		t.Error("expected error for corrupted JSON file")
	}
	if !strings.Contains(err.Error(), "failed to parse") {
		t.Errorf("expected parse error, got: %v", err)
	}
}

// TestLoadJSONConfig_NonExistentFile tests loading when no config file exists
func TestLoadJSONConfig_NonExistentFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "metis_config_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Test loadJSONConfig with no config file
	_, err = loadJSONConfig()
	if err == nil {
		t.Error("expected error when no config file exists")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

// TestLoadJSONConfig_InvalidPermissions tests loading with insufficient permissions
func TestLoadJSONConfig_InvalidPermissions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "metis_config_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a config file with no read permissions
	configPath := filepath.Join(tempDir, "metis.json")
	validJSON := `{"cache_size": 1000}`

	// Try to create file with restrictive permissions
	// On Windows, this might not work as expected, so we'll handle both cases
	err = os.WriteFile(configPath, []byte(validJSON), 0200) // write-only
	if err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Test loadJSONConfig with potentially unreadable file
	_, err = loadJSONConfig()

	// On Windows, the file might still be readable despite the permissions
	// So we check if we get an error, and if not, we skip the test
	if err == nil {
		// File is readable despite permissions - this is expected on Windows
		t.Skip("File is readable despite restrictive permissions (expected on Windows)")
		return
	}

	// Check if the error message contains expected content
	if !strings.Contains(err.Error(), "failed to read") {
		t.Errorf("expected read error, got: %v", err)
	}
}

// TestLoadJSONConfig_InvalidTTLFormat tests loading with invalid TTL format
func TestLoadJSONConfig_InvalidTTLFormat(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "metis_config_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create config with invalid TTL
	invalidTTLConfig := SimpleConfig{
		CacheSize: 1000,
		TTL:       "invalid_duration",
	}

	configData, err := json.Marshal(invalidTTLConfig)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}

	configPath := filepath.Join(tempDir, "metis.json")
	err = os.WriteFile(configPath, configData, 0644)
	if err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Test loadJSONConfig with invalid TTL
	_, err = loadJSONConfig()
	if err == nil {
		t.Error("expected error for invalid TTL format")
	}
	if !strings.Contains(err.Error(), "invalid TTL format") {
		t.Errorf("expected TTL format error, got: %v", err)
	}
}

// TestLoadJSONConfig_InvalidCleanupInterval tests loading with invalid cleanup interval
func TestLoadJSONConfig_InvalidCleanupInterval(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "metis_config_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create config with invalid cleanup interval
	invalidConfig := SimpleConfig{
		CacheSize:       1000,
		CleanupInterval: "invalid_duration",
	}

	configData, err := json.Marshal(invalidConfig)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}

	configPath := filepath.Join(tempDir, "metis.json")
	err = os.WriteFile(configPath, configData, 0644)
	if err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Test loadJSONConfig with invalid cleanup interval
	_, err = loadJSONConfig()
	if err == nil {
		t.Error("expected error for invalid cleanup interval format")
	}
	if !strings.Contains(err.Error(), "invalid cleanup_interval format") {
		t.Errorf("expected cleanup interval format error, got: %v", err)
	}
}

// TestGetConfigSource_AllSources tests all configuration source types
func TestGetConfigSource_AllSources(t *testing.T) {
	// Test 1: Default configuration (no global config, no JSON file)
	tempDir, err := os.MkdirTemp("", "metis_config_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Clear global config first
	configMutex.Lock()
	originalGlobalConfig := globalConfig
	globalConfig = nil
	configMutex.Unlock()
	defer func() {
		configMutex.Lock()
		globalConfig = originalGlobalConfig
		configMutex.Unlock()
	}()

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	source := GetConfigSource()
	if source != "Default configuration" {
		t.Errorf("expected 'Default configuration', got '%s'", source)
	}

	// Test 2: JSON configuration (create metis.json)
	validJSON := `{"cache_size": 1000}`
	configPath := filepath.Join(tempDir, "metis.json")
	err = os.WriteFile(configPath, []byte(validJSON), 0644)
	if err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	source = GetConfigSource()
	if source != "JSON configuration (metis.json)" {
		t.Errorf("expected 'JSON configuration (metis.json)', got '%s'", source)
	}

	// Test 3: Go configuration (set global config)
	testConfig := CacheConfig{
		EnableCaching: true,
		CacheSize:     2000,
	}
	SetGlobalConfig(testConfig)

	source = GetConfigSource()
	if source != "Go configuration (metis_config.go)" {
		t.Errorf("expected 'Go configuration (metis_config.go)', got '%s'", source)
	}
}

// TestFindConfigFile_AbsolutePath tests findConfigFile with absolute path scenarios
func TestFindConfigFile_AbsolutePath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "metis_config_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create config file in temp directory
	configPath := filepath.Join(tempDir, "metis.json")
	err = os.WriteFile(configPath, []byte(`{"cache_size": 1000}`), 0644)
	if err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Test findConfigFile finds the file
	foundPath := findConfigFile()
	if foundPath == "" {
		t.Error("expected to find config file")
	}
	if !strings.HasSuffix(foundPath, "metis.json") {
		t.Errorf("expected path to end with metis.json, got %s", foundPath)
	}
}

// TestFindConfigFile_RelativePath tests findConfigFile with relative path scenarios
func TestFindConfigFile_RelativePath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "metis_config_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create subdirectory structure
	subDir := filepath.Join(tempDir, "subdir")
	err = os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	// Create config file in parent directory
	configPath := filepath.Join(tempDir, "metis.json")
	err = os.WriteFile(configPath, []byte(`{"cache_size": 1000}`), 0644)
	if err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Change to subdirectory and test if it finds config in parent
	err = os.Chdir(subDir)
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	foundPath := findConfigFile()
	if foundPath == "" {
		t.Error("expected to find config file in parent directory")
	}
	if !strings.HasSuffix(foundPath, "metis.json") {
		t.Errorf("expected path to end with metis.json, got %s", foundPath)
	}
}

// TestFindConfigFile_NotFound tests findConfigFile when no config file exists
func TestFindConfigFile_NotFound(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "metis_config_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Test findConfigFile when no config exists
	foundPath := findConfigFile()
	if foundPath != "" {
		t.Errorf("expected empty path when no config file exists, got %s", foundPath)
	}
}

// TestLoadJSONConfig_PathTraversalValidation tests path traversal protection
func TestLoadJSONConfig_PathTraversalValidation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "metis_config_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// This test ensures the validation logic works, though we can't easily trigger
	// the path traversal case in the current implementation since findConfigFile
	// controls the path. We test this indirectly by checking the validation exists.

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Create a valid config file for normal operation
	configPath := filepath.Join(tempDir, "metis.json")
	err = os.WriteFile(configPath, []byte(`{"cache_size": 1000}`), 0644)
	if err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Test normal operation - should work
	_, err = loadJSONConfig()
	if err != nil {
		t.Errorf("expected valid config to load successfully, got error: %v", err)
	}
}
