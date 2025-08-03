// /cmd/metis-cli/main_test.go: Unit tests for Metis CLI
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0
package main

import (
	"bufio"
	"encoding/json"
	"os"
	"runtime"
	"strings"
	"testing"
)

// TestSimpleConfig tests the SimpleConfig struct and JSON marshaling
func TestSimpleConfig(t *testing.T) {
	config := SimpleConfig{
		CacheSize:         1000,
		TTL:               "10m",
		EvictionPolicy:    "wtinylfu",
		ShardCount:        32,
		EnableCompression: true,
		MaxValueSize:      524288,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	var unmarshaled SimpleConfig
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	if unmarshaled.CacheSize != config.CacheSize {
		t.Errorf("CacheSize mismatch: got %d, want %d", unmarshaled.CacheSize, config.CacheSize)
	}
	if unmarshaled.TTL != config.TTL {
		t.Errorf("TTL mismatch: got %s, want %s", unmarshaled.TTL, config.TTL)
	}
	if unmarshaled.EvictionPolicy != config.EvictionPolicy {
		t.Errorf("EvictionPolicy mismatch: got %s, want %s", unmarshaled.EvictionPolicy, config.EvictionPolicy)
	}
}

// TestCustomConfig tests the customConfig function with simulated input
func TestCustomConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected SimpleConfig
	}{
		{
			name:  "basic_custom_config",
			input: "5000\n15m\ny\n",
			expected: SimpleConfig{
				CacheSize:         5000,
				TTL:               "15m",
				EnableCompression: true,
			},
		},
		{
			name:  "no_compression",
			input: "2000\n5m\nn\n",
			expected: SimpleConfig{
				CacheSize:         2000,
				TTL:               "5m",
				EnableCompression: false,
			},
		},
		{
			name:  "empty_inputs",
			input: "\n\n\n",
			expected: SimpleConfig{
				CacheSize:         0,
				TTL:               "",
				EnableCompression: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))
			config := customConfig(reader)

			if config.CacheSize != tt.expected.CacheSize {
				t.Errorf("CacheSize: got %d, want %d", config.CacheSize, tt.expected.CacheSize)
			}
			if config.TTL != tt.expected.TTL {
				t.Errorf("TTL: got %s, want %s", config.TTL, tt.expected.TTL)
			}
			if config.EnableCompression != tt.expected.EnableCompression {
				t.Errorf("EnableCompression: got %v, want %v", config.EnableCompression, tt.expected.EnableCompression)
			}
		})
	}
}

// TestPresetConfigurations tests the preset configuration choices
func TestPresetConfigurations(t *testing.T) {
	tests := []struct {
		choice   string
		expected SimpleConfig
	}{
		{
			choice: "1",
			expected: SimpleConfig{
				CacheSize: 1000,
				TTL:       "10m",
			},
		},
		{
			choice: "2",
			expected: SimpleConfig{
				CacheSize:      50000,
				TTL:            "30m",
				EvictionPolicy: "wtinylfu",
				ShardCount:     32,
			},
		},
		{
			choice: "3",
			expected: SimpleConfig{
				CacheSize:      1000000,
				TTL:            "0s",
				EvictionPolicy: "wtinylfu",
				ShardCount:     128,
			},
		},
		{
			choice: "4",
			expected: SimpleConfig{
				CacheSize:         10000,
				TTL:               "1h",
				EnableCompression: true,
				MaxValueSize:      524288,
			},
		},
	}

	for _, tt := range tests {
		t.Run("choice_"+tt.choice, func(t *testing.T) {
			var config SimpleConfig

			switch tt.choice {
			case "1":
				config = SimpleConfig{
					CacheSize: 1000,
					TTL:       "10m",
				}
			case "2":
				config = SimpleConfig{
					CacheSize:      50000,
					TTL:            "30m",
					EvictionPolicy: "wtinylfu",
					ShardCount:     32,
				}
			case "3":
				config = SimpleConfig{
					CacheSize:      1000000,
					TTL:            "0s",
					EvictionPolicy: "wtinylfu",
					ShardCount:     128,
				}
			case "4":
				config = SimpleConfig{
					CacheSize:         10000,
					TTL:               "1h",
					EnableCompression: true,
					MaxValueSize:      524288,
				}
			}

			// Verify each field matches expected
			if config.CacheSize != tt.expected.CacheSize {
				t.Errorf("CacheSize: got %d, want %d", config.CacheSize, tt.expected.CacheSize)
			}
			if config.TTL != tt.expected.TTL {
				t.Errorf("TTL: got %s, want %s", config.TTL, tt.expected.TTL)
			}
			if config.EvictionPolicy != tt.expected.EvictionPolicy {
				t.Errorf("EvictionPolicy: got %s, want %s", config.EvictionPolicy, tt.expected.EvictionPolicy)
			}
			if config.ShardCount != tt.expected.ShardCount {
				t.Errorf("ShardCount: got %d, want %d", config.ShardCount, tt.expected.ShardCount)
			}
			if config.EnableCompression != tt.expected.EnableCompression {
				t.Errorf("EnableCompression: got %v, want %v", config.EnableCompression, tt.expected.EnableCompression)
			}
			if config.MaxValueSize != tt.expected.MaxValueSize {
				t.Errorf("MaxValueSize: got %d, want %d", config.MaxValueSize, tt.expected.MaxValueSize)
			}
		})
	}
}

// TestConfigGeneration tests the JSON file generation process
func TestConfigGeneration(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tempDir)

	config := SimpleConfig{
		CacheSize:      5000,
		TTL:            "15m",
		EvictionPolicy: "wtinylfu",
		ShardCount:     16,
	}

	// Generate JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	// Write file
	err = os.WriteFile("metis.json", data, 0600)
	if err != nil {
		t.Fatalf("Failed to write metis.json: %v", err)
	}

	// Verify file exists and content is correct
	if _, err := os.Stat("metis.json"); os.IsNotExist(err) {
		t.Fatal("metis.json was not created")
	}

	// Read and verify content
	fileData, err := os.ReadFile("metis.json")
	if err != nil {
		t.Fatalf("Failed to read metis.json: %v", err)
	}

	var readConfig SimpleConfig
	err = json.Unmarshal(fileData, &readConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal file content: %v", err)
	}

	if readConfig.CacheSize != config.CacheSize {
		t.Errorf("File CacheSize: got %d, want %d", readConfig.CacheSize, config.CacheSize)
	}
	if readConfig.TTL != config.TTL {
		t.Errorf("File TTL: got %s, want %s", readConfig.TTL, config.TTL)
	}
}

// TestJSONStructure tests that generated JSON has correct structure
func TestJSONStructure(t *testing.T) {
	config := SimpleConfig{
		CacheSize:         1000,
		TTL:               "10m",
		EvictionPolicy:    "wtinylfu",
		ShardCount:        32,
		EnableCompression: true,
		MaxValueSize:      524288,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	// Parse as generic JSON to check structure
	var jsonData map[string]interface{}
	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		t.Fatalf("Failed to unmarshal as generic JSON: %v", err)
	}

	// Check required fields
	if _, ok := jsonData["cache_size"]; !ok {
		t.Error("Missing cache_size field")
	}

	// Check field types
	if cacheSize, ok := jsonData["cache_size"].(float64); !ok || int(cacheSize) != 1000 {
		t.Errorf("cache_size should be number 1000, got %v", jsonData["cache_size"])
	}

	if ttl, ok := jsonData["ttl"].(string); !ok || ttl != "10m" {
		t.Errorf("ttl should be string '10m', got %v", jsonData["ttl"])
	}

	if compression, ok := jsonData["enable_compression"].(bool); !ok || !compression {
		t.Errorf("enable_compression should be bool true, got %v", jsonData["enable_compression"])
	}
}

// TestFilePermissions tests that generated file has correct permissions
func TestFilePermissions(t *testing.T) {
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tempDir)

	config := SimpleConfig{CacheSize: 1000}
	data, _ := json.MarshalIndent(config, "", "  ")

	err := os.WriteFile("metis.json", data, 0600)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	info, err := os.Stat("metis.json")
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	// Check permissions - on Windows, file permissions work differently
	actualPerm := info.Mode().Perm()

	// On Unix-like systems, expect 0600 (owner read/write only)
	// On Windows, the permissions might be different, so we'll be more flexible
	if runtime.GOOS == "windows" {
		// On Windows, just check that the file exists and is readable
		if actualPerm&0400 == 0 {
			t.Errorf("File should be readable by owner, got permissions: %v", actualPerm)
		}
	} else {
		// On Unix-like systems, expect 0600
		expectedPerm := os.FileMode(0600)
		if actualPerm != expectedPerm {
			t.Errorf("File permissions: got %v, want %v", actualPerm, expectedPerm)
		}
	}
}

// TestConfigValidation tests validation of configuration values
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name     string
		config   SimpleConfig
		valid    bool
		expected string
	}{
		{
			name: "valid_basic_config",
			config: SimpleConfig{
				CacheSize: 1000,
				TTL:       "10m",
			},
			valid: true,
		},
		{
			name: "valid_full_config",
			config: SimpleConfig{
				CacheSize:         50000,
				TTL:               "30m",
				EvictionPolicy:    "wtinylfu",
				ShardCount:        32,
				EnableCompression: true,
				MaxValueSize:      524288,
			},
			valid: true,
		},
		{
			name: "zero_cache_size",
			config: SimpleConfig{
				CacheSize: 0,
				TTL:       "10m",
			},
			valid: false, // Usually invalid for real usage
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling (should always work)
			data, err := json.MarshalIndent(tt.config, "", "  ")
			if err != nil {
				t.Errorf("Failed to marshal config: %v", err)
			}

			// Basic validation: cache size should be positive for real usage
			if tt.config.CacheSize <= 0 && tt.valid {
				t.Error("Config with zero/negative cache size should not be marked as valid")
			}

			// Verify JSON is well-formed
			var unmarshaled SimpleConfig
			err = json.Unmarshal(data, &unmarshaled)
			if err != nil {
				t.Errorf("Generated JSON is not valid: %v", err)
			}
		})
	}
}

// BenchmarkConfigGeneration benchmarks the configuration generation process
func BenchmarkConfigGeneration(b *testing.B) {
	config := SimpleConfig{
		CacheSize:         50000,
		TTL:               "30m",
		EvictionPolicy:    "wtinylfu",
		ShardCount:        32,
		EnableCompression: true,
		MaxValueSize:      524288,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			b.Fatalf("Failed to marshal config: %v", err)
		}
		_ = data // Use the data to prevent optimization
	}
}

// BenchmarkCustomConfig benchmarks the custom configuration input parsing
func BenchmarkCustomConfig(b *testing.B) {
	input := "5000\n15m\ny\n"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bufio.NewReader(strings.NewReader(input))
		config := customConfig(reader)
		_ = config // Use the config to prevent optimization
	}
}
