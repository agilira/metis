// main_test.go: Comprehensive test suite for Metis Debug CLI
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestMain runs setup and teardown for all tests
func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()
	os.Exit(code)
}

// captureOutput captures stdout during function execution
func captureOutput(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

// setArgs temporarily sets os.Args for testing
func setArgs(args []string, fn func()) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = args
	fn()
}

// TestShowHelp tests the help command output
func TestShowHelp(t *testing.T) {
	output := captureOutput(showHelp)

	expectedStrings := []string{
		"Metis Debug CLI",
		"USAGE:",
		"inspect",
		"version",
		"help",
		"-json",
		"-v",
		"-real",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Help output missing expected string: %s", expected)
		}
	}
}

// TestCmdVersion tests the version command
func TestCmdVersion(t *testing.T) {
	output := captureOutput(cmdVersion)

	expectedStrings := []string{
		"metis-debug version",
		VERSION,
		"Go version:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Version output missing expected string: %s", expected)
		}
	}
}

// TestMainCommandRouting tests command routing in main()
func TestMainCommandRouting(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
		exitCode int
	}{
		{
			name:     "no args shows help",
			args:     []string{"metis-debug"},
			expected: "USAGE:",
			exitCode: 0,
		},
		{
			name:     "help command",
			args:     []string{"metis-debug", "help"},
			expected: "USAGE:",
			exitCode: 0,
		},
		{
			name:     "version command",
			args:     []string{"metis-debug", "version"},
			expected: "metis-debug version",
			exitCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output string
			setArgs(tt.args, func() {
				output = captureOutput(main)
			})

			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected output to contain '%s', got: %s", tt.expected, output)
			}
		})
	}

	// Test unknown command separately since it calls os.Exit
	t.Run("unknown command", func(t *testing.T) {
		if os.Getenv("BE_CRASHER") == "1" {
			setArgs([]string{"metis-debug", "unknown"}, main)
			return
		}

		// Use subprocess to test os.Exit behavior
		cmd := exec.Command(os.Args[0], "-test.run=TestMainCommandRouting/unknown")
		cmd.Env = append(os.Environ(), "BE_CRASHER=1")

		output, err := cmd.CombinedOutput()
		if e, ok := err.(*exec.ExitError); ok && e.ExitCode() == 1 {
			// Expected exit code 1
			if !strings.Contains(string(output), "Unknown command") {
				t.Errorf("Expected output to contain 'Unknown command', got: %s", string(output))
			}
		} else {
			t.Errorf("Expected exit code 1, got: %v", err)
		}
	})
}

// TestPerformHealthCheck tests health check functionality
func TestPerformHealthCheck(t *testing.T) {
	tests := []struct {
		name       string
		jsonOutput bool
		expected   string
	}{
		{
			name:       "text output",
			jsonOutput: false,
			expected:   "Cache Performance Analysis",
		},
		{
			name:       "json output",
			jsonOutput: true,
			expected:   "", // Should return early for JSON
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				performHealthCheck(tt.jsonOutput)
			})

			if tt.expected == "" {
				if output != "" {
					t.Errorf("Expected no output for JSON mode, got: %s", output)
				}
			} else {
				if !strings.Contains(output, tt.expected) {
					t.Errorf("Expected output to contain '%s', got: %s", tt.expected, output)
				}
			}
		})
	}
}

// TestShowStats tests the estimated stats functionality
func TestShowStats(t *testing.T) {
	tests := []struct {
		name       string
		jsonOutput bool
		verbose    bool
	}{
		{
			name:       "text output",
			jsonOutput: false,
			verbose:    false,
		},
		{
			name:       "json output",
			jsonOutput: true,
			verbose:    false,
		},
		{
			name:       "verbose text output",
			jsonOutput: false,
			verbose:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				showStats(tt.jsonOutput, tt.verbose)
			})

			if tt.jsonOutput {
				// Test JSON output
				var jsonData map[string]interface{}
				err := json.Unmarshal([]byte(output), &jsonData)
				if err != nil {
					t.Errorf("Invalid JSON output: %v", err)
				}

				// Check required JSON fields
				if cache, ok := jsonData["cache"].(map[string]interface{}); ok {
					if _, ok := cache["estimated_ops_per_sec"]; !ok {
						t.Error("JSON missing estimated_ops_per_sec field")
					}
					if _, ok := cache["type"]; !ok {
						t.Error("JSON missing type field")
					}
				} else {
					t.Error("JSON missing cache section")
				}
			} else {
				// Test text output
				expectedStrings := []string{
					"Runtime Information",
					"Memory Statistics",
					"Cache Performance Metrics",
					"3,500,000",
					"W-TinyLFU",
				}

				for _, expected := range expectedStrings {
					if !strings.Contains(output, expected) {
						t.Errorf("Text output missing expected string: %s", expected)
					}
				}
			}
		})
	}
}

// TestShowRealStats tests the real stats functionality
func TestShowRealStats(t *testing.T) {
	tests := []struct {
		name       string
		jsonOutput bool
		verbose    bool
	}{
		{
			name:       "real text output",
			jsonOutput: false,
			verbose:    false,
		},
		{
			name:       "real json output",
			jsonOutput: true,
			verbose:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				showRealStats(tt.jsonOutput, tt.verbose)
			})

			if tt.jsonOutput {
				// Test JSON output
				var jsonData map[string]interface{}
				err := json.Unmarshal([]byte(output), &jsonData)
				if err != nil {
					t.Errorf("Invalid JSON output: %v", err)
				}

				// Check required JSON fields for real mode
				if cache, ok := jsonData["cache"].(map[string]interface{}); ok {
					if _, ok := cache["real_ops_per_sec"]; !ok {
						t.Error("JSON missing real_ops_per_sec field")
					}
					if cacheType, ok := cache["type"].(string); !ok || !strings.Contains(cacheType, "Real") {
						t.Error("JSON missing or incorrect type field for real mode")
					}
				} else {
					t.Error("JSON missing cache section")
				}

				// Check config section exists in real mode
				if _, ok := jsonData["config"]; !ok {
					t.Error("JSON missing config section in real mode")
				}
			} else {
				// Test text output
				expectedStrings := []string{
					"REAL Metis Cache Analysis",
					"Cache Configuration",
					"Real Performance Measurements",
					"wtinylfu",
					"Operations/sec:",
				}

				for _, expected := range expectedStrings {
					if !strings.Contains(output, expected) {
						t.Errorf("Real text output missing expected string: %s", expected)
					}
				}
			}
		})
	}
}

// TestFormatNumber tests the number formatting utility
func TestFormatNumber(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{123, "123"},
		{1234, "1,234"},
		{12345, "12,345"},
		{123456, "123,456"},
		{1234567, "1,234,567"},
		{1000000, "1,000,000"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("format_%d", tt.input), func(t *testing.T) {
			result := formatNumber(tt.input)
			if result != tt.expected {
				t.Errorf("formatNumber(%d) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

// TestMeasureRealPerformance tests the performance measurement function
func TestMeasureRealPerformance(t *testing.T) {
	// This test requires actual Metis integration
	// Skip if running in CI without proper setup
	if testing.Short() {
		t.Skip("Skipping real performance test in short mode")
	}

	// Note: This test actually creates a real cache and measures performance
	// It's more of an integration test
	t.Run("real_performance_measurement", func(t *testing.T) {
		// We can't easily test this without mocking the cache
		// For now, we'll test the structure
		// In a real scenario, you'd want to create a mock cache interface
		t.Skip("Real performance measurement requires integration test setup")
	})
}

// TestCmdInspectFlagParsing tests flag parsing in cmdInspect
func TestCmdInspectFlagParsing(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "no flags",
			args: []string{},
		},
		{
			name: "json flag",
			args: []string{"-json"},
		},
		{
			name: "verbose flag",
			args: []string{"-v"},
		},
		{
			name: "real flag",
			args: []string{"-real"},
		},
		{
			name: "multiple flags",
			args: []string{"-json", "-v", "-real"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that flag parsing doesn't panic
			output := captureOutput(func() {
				cmdInspect(tt.args)
			})

			// Basic validation - should produce some output
			if len(output) == 0 {
				t.Error("cmdInspect produced no output")
			}
		})
	}
}

// BenchmarkShowStats benchmarks the stats function
func BenchmarkShowStats(b *testing.B) {
	for i := 0; i < b.N; i++ {
		captureOutput(func() {
			showStats(false, false)
		})
	}
}

// BenchmarkShowStatsJSON benchmarks the JSON stats function
func BenchmarkShowStatsJSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		captureOutput(func() {
			showStats(true, false)
		})
	}
}

// TestConstants verifies important constants
func TestConstants(t *testing.T) {
	if VERSION == "" {
		t.Error("VERSION constant should not be empty")
	}

	if !strings.Contains(VERSION, ".") {
		t.Error("VERSION should be in semantic version format")
	}
}
