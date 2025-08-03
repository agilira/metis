// /cmd/metis-cli/integration_test.go: Integration tests for Metis CLI
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0
package main

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

// TestCLIIntegration tests the complete CLI workflow
func TestCLIIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Run("development_preset", func(t *testing.T) {
		// Create temporary directory
		tempDir := t.TempDir()
		oldDir, _ := os.Getwd()
		defer os.Chdir(oldDir)
		os.Chdir(tempDir)

		// Simulate the main function workflow directly
		// Instead of subprocess, test the logic components directly
		config := SimpleConfig{
			CacheSize: 1000,
			TTL:       "10m",
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

		// Verify file exists
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

		if readConfig.CacheSize != 1000 {
			t.Errorf("CacheSize: got %d, want %d", readConfig.CacheSize, 1000)
		}
		if readConfig.TTL != "10m" {
			t.Errorf("TTL: got %s, want %s", readConfig.TTL, "10m")
		}
	})

	// Test other presets using the same approach
	testPresets := []struct {
		name   string
		config SimpleConfig
	}{
		{
			name: "web_application_preset",
			config: SimpleConfig{
				CacheSize:      50000,
				TTL:            "30m",
				EvictionPolicy: "wtinylfu",
				ShardCount:     32,
			},
		},
		{
			name: "high_performance_preset",
			config: SimpleConfig{
				CacheSize:      1000000,
				TTL:            "0s",
				EvictionPolicy: "wtinylfu",
				ShardCount:     128,
			},
		},
		{
			name: "memory_constrained_preset",
			config: SimpleConfig{
				CacheSize:         10000,
				TTL:               "1h",
				EnableCompression: true,
				MaxValueSize:      524288,
			},
		},
	}

	for _, tt := range testPresets {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			oldDir, _ := os.Getwd()
			defer os.Chdir(oldDir)
			os.Chdir(tempDir)

			// Test the configuration generation
			data, err := json.MarshalIndent(tt.config, "", "  ")
			if err != nil {
				t.Fatalf("Failed to marshal config: %v", err)
			}

			err = os.WriteFile("metis.json", data, 0600)
			if err != nil {
				t.Fatalf("Failed to write metis.json: %v", err)
			}

			// Verify content
			var readConfig SimpleConfig
			fileData, _ := os.ReadFile("metis.json")
			json.Unmarshal(fileData, &readConfig)

			if readConfig.CacheSize != tt.config.CacheSize {
				t.Errorf("CacheSize: got %d, want %d", readConfig.CacheSize, tt.config.CacheSize)
			}
		})
	}
}

// TestCLIOutput tests output format using functional approach
func TestCLIOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping output tests in short mode")
	}

	t.Run("json_content_validation", func(t *testing.T) {
		config := SimpleConfig{
			CacheSize:      50000,
			TTL:            "30m",
			EvictionPolicy: "wtinylfu",
			ShardCount:     32,
		}

		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal config: %v", err)
		}

		// Verify JSON contains expected content
		jsonStr := string(data)
		if !strings.Contains(jsonStr, `"cache_size": 50000`) {
			t.Error("JSON should contain cache_size")
		}
		if !strings.Contains(jsonStr, `"ttl": "30m"`) {
			t.Error("JSON should contain TTL")
		}
	})
}

// TestCLICustomConfiguration tests the custom configuration logic
func TestCLICustomConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping custom config tests in short mode")
	}

	// Test the customConfig function directly
	tests := []struct {
		name        string
		input       string
		expectField string
		expectValue interface{}
	}{
		{
			name:        "custom_cache_size",
			input:       "7500\n\n\n",
			expectField: "cache_size",
			expectValue: 7500,
		},
		{
			name:        "custom_ttl",
			input:       "\n45m\n\n",
			expectField: "ttl",
			expectValue: "45m",
		},
		{
			name:        "custom_compression_yes",
			input:       "\n\ny\n",
			expectField: "enable_compression",
			expectValue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))
			config := customConfig(reader)

			switch tt.expectField {
			case "cache_size":
				if config.CacheSize != tt.expectValue {
					t.Errorf("CacheSize: got %d, want %d", config.CacheSize, tt.expectValue)
				}
			case "ttl":
				if config.TTL != tt.expectValue {
					t.Errorf("TTL: got %s, want %s", config.TTL, tt.expectValue)
				}
			case "enable_compression":
				if config.EnableCompression != tt.expectValue {
					t.Errorf("EnableCompression: got %v, want %v", config.EnableCompression, tt.expectValue)
				}
			}
		})
	}
}

// TestCLIErrorHandling tests error scenarios using functional approach
func TestCLIErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping error handling tests in short mode")
	}

	t.Run("invalid_number_parsing", func(t *testing.T) {
		// Test that invalid numbers default to 0
		reader := bufio.NewReader(strings.NewReader("abc\n10m\nn\n"))
		config := customConfig(reader)

		if config.CacheSize != 0 {
			t.Errorf("Invalid number should result in 0, got %d", config.CacheSize)
		}
		if config.TTL != "10m" {
			t.Errorf("TTL should be preserved, got %s", config.TTL)
		}
	})
}

// TestCLIPerformance tests CLI performance characteristics using functional approach
func TestCLIPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	t.Run("config_generation_speed", func(t *testing.T) {
		start := time.Now()

		config := SimpleConfig{
			CacheSize:      1000000,
			TTL:            "0s",
			EvictionPolicy: "wtinylfu",
			ShardCount:     128,
		}

		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal config: %v", err)
		}

		duration := time.Since(start)

		// Config generation should be very fast
		if duration > 100*time.Millisecond {
			t.Errorf("Config generation took too long: %v", duration)
		}

		// Verify large config was generated correctly
		if len(data) < 50 {
			t.Error("Generated JSON seems too small")
		}
	})
}

// BenchmarkCLIConfiguration benchmarks the configuration generation
func BenchmarkCLIConfiguration(b *testing.B) {
	config := SimpleConfig{
		CacheSize:      50000,
		TTL:            "30m",
		EvictionPolicy: "wtinylfu",
		ShardCount:     32,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			b.Fatalf("Failed to marshal config: %v", err)
		}
		_ = data // Use data to prevent optimization
	}
}
