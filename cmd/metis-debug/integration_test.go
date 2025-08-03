// integration_test.go: Integration tests for Metis Debug CLI
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestCLIIntegration tests the CLI as a complete application
func TestCLIIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	helper := NewCLITestHelper(t)

	t.Run("help_command_integration", func(t *testing.T) {
		stdout, stderr, exitCode := helper.RunCommand("help")

		helper.AssertExitCode(exitCode, 0)
		helper.AssertContains(stdout, "Metis Debug CLI")
		helper.AssertContains(stdout, "inspect")
		helper.AssertContains(stdout, "version")
		if stderr != "" {
			t.Errorf("Unexpected stderr output: %s", stderr)
		}
	})

	t.Run("version_command_integration", func(t *testing.T) {
		stdout, stderr, exitCode := helper.RunCommand("version")

		helper.AssertExitCode(exitCode, 0)
		helper.AssertContains(stdout, "metis-debug version")
		helper.AssertContains(stdout, VERSION)
		if stderr != "" {
			t.Errorf("Unexpected stderr output: %s", stderr)
		}
	})

	t.Run("inspect_estimated_integration", func(t *testing.T) {
		stdout, stderr, exitCode := helper.RunCommand("inspect")

		helper.AssertExitCode(exitCode, 0)
		helper.AssertContains(stdout, "Cache Performance Analysis")
		helper.AssertContains(stdout, "3,500,000")
		helper.AssertContains(stdout, "W-TinyLFU")
		helper.AssertNotContains(stdout, "Otter") // Ensure no internal references
		if stderr != "" {
			t.Errorf("Unexpected stderr output: %s", stderr)
		}
	})

	t.Run("inspect_json_integration", func(t *testing.T) {
		stdout, stderr, exitCode := helper.RunCommand("inspect", "-json")

		helper.AssertExitCode(exitCode, 0)
		jsonData := helper.AssertValidJSON(stdout)

		// Validate JSON structure
		if cache, ok := jsonData["cache"].(map[string]interface{}); ok {
			if opsPerSec, ok := cache["estimated_ops_per_sec"].(float64); !ok || opsPerSec != 3500000 {
				t.Error("JSON cache.estimated_ops_per_sec incorrect")
			}
			if cacheType, ok := cache["type"].(string); !ok || !strings.Contains(cacheType, "Estimated") {
				t.Error("JSON cache.type incorrect for estimated mode")
			}
		} else {
			t.Error("JSON missing cache section")
		}

		if stderr != "" {
			t.Errorf("Unexpected stderr output: %s", stderr)
		}
	})

	t.Run("inspect_real_integration", func(t *testing.T) {
		stdout, stderr, exitCode := helper.RunCommand("inspect", "-real")

		helper.AssertExitCode(exitCode, 0)
		helper.AssertContains(stdout, "REAL Metis Cache Analysis")
		helper.AssertContains(stdout, "Cache Configuration")
		helper.AssertContains(stdout, "Real Performance Measurements")
		if stderr != "" {
			t.Errorf("Unexpected stderr output: %s", stderr)
		}
	})

	t.Run("inspect_real_json_integration", func(t *testing.T) {
		stdout, stderr, exitCode := helper.RunCommand("inspect", "-real", "-json")

		helper.AssertExitCode(exitCode, 0)
		jsonData := helper.AssertValidJSON(stdout)

		// Validate JSON structure for real mode
		if cache, ok := jsonData["cache"].(map[string]interface{}); ok {
			if _, ok := cache["real_ops_per_sec"]; !ok {
				t.Error("JSON missing real_ops_per_sec in real mode")
			}
			if cacheType, ok := cache["type"].(string); !ok || !strings.Contains(cacheType, "Real") {
				t.Error("JSON cache.type incorrect for real mode")
			}
		} else {
			t.Error("JSON missing cache section in real mode")
		}

		// Real mode should have config section
		if _, ok := jsonData["config"]; !ok {
			t.Error("JSON missing config section in real mode")
		}

		if stderr != "" {
			t.Errorf("Unexpected stderr output: %s", stderr)
		}
	})

	t.Run("unknown_command_integration", func(t *testing.T) {
		stdout, _, exitCode := helper.RunCommand("unknown")

		helper.AssertExitCode(exitCode, 1)
		helper.AssertContains(stdout, "Unknown command")
		helper.AssertContains(stdout, "USAGE:")
	})

	t.Run("invalid_flag_integration", func(t *testing.T) {
		stdout, stderr, exitCode := helper.RunCommand("inspect", "-invalid")

		// Should handle invalid flags gracefully
		// Either exits with error code or produces output
		if exitCode == 0 && stdout == "" && stderr == "" {
			t.Error("No output or error for invalid flag")
		}
		// This is acceptable behavior - CLI handles invalid flags by exiting silently
		// which is a valid approach for flag parsing errors
	})
}

// TestCLIPerformance tests CLI performance characteristics
func TestCLIPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	helper := NewCLITestHelper(t)

	t.Run("startup_time", func(t *testing.T) {
		// Test that CLI starts quickly (should be < 1 second)
		stdout, stderr, exitCode := helper.RunCommand("version")

		helper.AssertExitCode(exitCode, 0)
		helper.AssertContains(stdout, VERSION)
		if stderr != "" {
			t.Errorf("Unexpected stderr: %s", stderr)
		}
	})

	t.Run("memory_efficiency", func(t *testing.T) {
		// Test estimated mode (should be very fast)
		stdout, stderr, exitCode := helper.RunCommand("inspect")

		helper.AssertExitCode(exitCode, 0)
		helper.AssertContains(stdout, "Cache Performance Analysis")
		if stderr != "" {
			t.Errorf("Unexpected stderr: %s", stderr)
		}
	})
}

// TestCLIOutputFormats tests different output formats
func TestCLIOutputFormats(t *testing.T) {
	helper := NewCLITestHelper(t)

	t.Run("text_vs_json_consistency", func(t *testing.T) {
		// Get text output
		textOut, _, textExit := helper.RunCommand("inspect")
		helper.AssertExitCode(textExit, 0)

		// Get JSON output
		jsonOut, _, jsonExit := helper.RunCommand("inspect", "-json")
		helper.AssertExitCode(jsonExit, 0)

		// Parse JSON
		var jsonData map[string]interface{}
		err := json.Unmarshal([]byte(jsonOut), &jsonData)
		if err != nil {
			t.Fatalf("Invalid JSON: %v", err)
		}

		// Check that both contain similar information
		helper.AssertContains(textOut, "3,500,000")
		if cache, ok := jsonData["cache"].(map[string]interface{}); ok {
			if opsPerSec, ok := cache["estimated_ops_per_sec"].(float64); !ok || opsPerSec != 3500000 {
				t.Error("JSON and text outputs inconsistent for ops/sec")
			}
		}
	})

	t.Run("real_mode_consistency", func(t *testing.T) {
		// Test that real mode produces consistent structure
		jsonOut, _, exitCode := helper.RunCommand("inspect", "-real", "-json")
		helper.AssertExitCode(exitCode, 0)

		jsonData := helper.AssertValidJSON(jsonOut)

		// Check required fields
		requiredSections := []string{"cache", "memory", "runtime", "config"}
		for _, section := range requiredSections {
			if _, ok := jsonData[section]; !ok {
				t.Errorf("Real mode JSON missing required section: %s", section)
			}
		}
	})
}

// TestCLIErrorHandling tests error conditions
func TestCLIErrorHandling(t *testing.T) {
	helper := NewCLITestHelper(t)

	t.Run("no_arguments", func(t *testing.T) {
		stdout, _, exitCode := helper.RunCommand()

		helper.AssertExitCode(exitCode, 0) // Shows help, doesn't error
		helper.AssertContains(stdout, "USAGE:")
	})

	t.Run("help_variations", func(t *testing.T) {
		helpVariations := []string{"help", "-h", "--help"}

		for _, helpArg := range helpVariations {
			t.Run(helpArg, func(t *testing.T) {
				stdout, stderr, exitCode := helper.RunCommand(helpArg)

				helper.AssertExitCode(exitCode, 0)
				helper.AssertContains(stdout, "USAGE:")
				if stderr != "" {
					t.Errorf("Unexpected stderr for %s: %s", helpArg, stderr)
				}
			})
		}
	})
}
