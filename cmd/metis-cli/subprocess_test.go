// /cmd/metis-cli/subprocess_test.go: Tests for Metis CLI subprocess execution
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0
package main

import (
	"os"
	"strings"
	"testing"
)

// TestCLISubprocess tests CLI execution using subprocess approach
func TestCLISubprocess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping subprocess tests in short mode")
	}

	helper := NewCLITestHelper(t)

	t.Run("development_preset_subprocess", func(t *testing.T) {
		// Create isolated test directory
		tempDir := t.TempDir()
		oldDir, _ := os.Getwd()
		defer os.Chdir(oldDir)
		os.Chdir(tempDir)

		stdout, stderr, exitCode := helper.RunCLIWithInputInDir("1\n", tempDir)

		// Check exit code
		if exitCode != 0 {
			t.Fatalf("CLI failed with exit code %d. Stderr: %s. Stdout: %s", exitCode, stderr, stdout)
		}

		// Check output contains expected strings
		helper.AssertContains(stdout, "🚀 Metis Configuration Generator")
		helper.AssertContains(stdout, "✅ Generated metis.json successfully!")

		// Check file was created
		if !helper.CheckFileExists("metis.json") {
			t.Error("metis.json was not created")
			return
		}

		// Verify JSON content
		jsonData := helper.ReadJSONFile("metis.json")
		if cacheSize, ok := jsonData["cache_size"]; !ok {
			t.Error("Missing cache_size field")
		} else if cacheSize != float64(1000) {
			t.Errorf("cache_size: got %v, want %v", cacheSize, float64(1000))
		}

		if ttl, ok := jsonData["ttl"]; !ok {
			t.Error("Missing ttl field")
		} else if ttl != "10m" {
			t.Errorf("ttl: got %v, want %v", ttl, "10m")
		}
	})

	t.Run("web_application_preset_subprocess", func(t *testing.T) {
		tempDir := t.TempDir()
		oldDir, _ := os.Getwd()
		defer os.Chdir(oldDir)
		os.Chdir(tempDir)

		stdout, _, exitCode := helper.RunCLIWithInputInDir("2\n", tempDir)

		if exitCode != 0 {
			t.Fatalf("CLI failed with exit code %d", exitCode)
		}

		helper.AssertContains(stdout, "Generated metis.json successfully!")

		if !helper.CheckFileExists("metis.json") {
			t.Error("metis.json was not created")
			return
		}

		jsonData := helper.ReadJSONFile("metis.json")
		if cacheSize, ok := jsonData["cache_size"]; !ok || cacheSize != float64(50000) {
			t.Errorf("cache_size: got %v, want %v", cacheSize, float64(50000))
		}
		if evictionPolicy, ok := jsonData["eviction_policy"]; !ok || evictionPolicy != "wtinylfu" {
			t.Errorf("eviction_policy: got %v, want %v", evictionPolicy, "wtinylfu")
		}
	})

	t.Run("custom_configuration_subprocess", func(t *testing.T) {
		tempDir := t.TempDir()
		oldDir, _ := os.Getwd()
		defer os.Chdir(oldDir)
		os.Chdir(tempDir)

		input := "5\n2500\n15m\ny\n"
		stdout, _, exitCode := helper.RunCLIWithInputInDir(input, tempDir)

		if exitCode != 0 {
			t.Fatalf("CLI failed with exit code %d", exitCode)
		}

		helper.AssertContains(stdout, "Generated metis.json successfully!")

		if !helper.CheckFileExists("metis.json") {
			t.Error("metis.json was not created")
			return
		}

		jsonData := helper.ReadJSONFile("metis.json")
		if cacheSize, ok := jsonData["cache_size"]; !ok || cacheSize != float64(2500) {
			t.Errorf("cache_size: got %v, want %v", cacheSize, float64(2500))
		}
		if ttl, ok := jsonData["ttl"]; !ok || ttl != "15m" {
			t.Errorf("ttl: got %v, want %v", ttl, "15m")
		}
		if compression, ok := jsonData["enable_compression"]; !ok || compression != true {
			t.Errorf("enable_compression: got %v, want %v", compression, true)
		}
	})

	t.Run("invalid_choice_subprocess", func(t *testing.T) {
		tempDir := t.TempDir()
		oldDir, _ := os.Getwd()
		defer os.Chdir(oldDir)
		os.Chdir(tempDir)

		stdout, _, exitCode := helper.RunCLIWithInputInDir("99\n", tempDir)

		if exitCode != 0 {
			t.Fatalf("CLI failed with exit code %d", exitCode)
		}

		helper.AssertContains(stdout, "Invalid choice, using development defaults")
		helper.AssertContains(stdout, "Generated metis.json successfully!")

		if !helper.CheckFileExists("metis.json") {
			t.Error("metis.json was not created")
			return
		}

		jsonData := helper.ReadJSONFile("metis.json")
		if cacheSize, ok := jsonData["cache_size"]; !ok || cacheSize != float64(1000) {
			t.Errorf("cache_size: got %v, want %v", cacheSize, float64(1000))
		}
	})
}

// TestCLIOutputFormat tests output formatting using subprocess
func TestCLIOutputFormat(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping output format tests in short mode")
	}

	helper := NewCLITestHelper(t)

	t.Run("output_contains_all_elements", func(t *testing.T) {
		tempDir := t.TempDir()

		stdout, _, exitCode := helper.RunCLIWithInputInDir("1\n", tempDir)

		if exitCode != 0 {
			t.Fatalf("CLI failed with exit code %d", exitCode)
		}

		expectedElements := []string{
			"🚀 Metis Configuration Generator",
			"===================================",
			"What's your primary use case?",
			"1. Development/Testing",
			"2. Web Application",
			"3. High-Performance API",
			"4. Memory-Constrained",
			"5. Custom configuration",
			"6. Exit",
			"Choose (1-6):",
			"✅ Generated metis.json successfully!",
			"📝 Content:",
			`"cache_size": 1000`,
			`"ttl": "10m"`,
			"🚀 You can now use metis.New() in your code!",
		}

		for _, element := range expectedElements {
			helper.AssertContains(stdout, element)
		}
	})

	t.Run("json_output_is_valid", func(t *testing.T) {
		tempDir := t.TempDir()

		stdout, _, exitCode := helper.RunCLIWithInputInDir("2\n", tempDir)

		if exitCode != 0 {
			t.Fatalf("CLI failed with exit code %d", exitCode)
		}

		// Extract JSON from output (it should be between "📝 Content:" and "🚀 You can")
		lines := strings.Split(stdout, "\n")
		var jsonStart, jsonEnd int
		for i, line := range lines {
			if strings.Contains(line, "📝 Content:") {
				jsonStart = i + 1
			}
			if strings.Contains(line, "🚀 You can now use") {
				jsonEnd = i
				break
			}
		}

		if jsonStart == 0 || jsonEnd == 0 {
			t.Fatal("Could not find JSON content in output")
		}

		jsonLines := lines[jsonStart:jsonEnd]
		jsonStr := strings.Join(jsonLines, "\n")
		jsonStr = strings.TrimSpace(jsonStr)

		// Validate JSON
		jsonData := helper.AssertValidJSON(jsonStr)

		// Check specific fields for web application preset
		if cacheSize, ok := jsonData["cache_size"]; !ok || cacheSize != float64(50000) {
			t.Errorf("cache_size in output: got %v, want %v", cacheSize, float64(50000))
		}
	})
}

// TestCLIErrorScenarios tests error handling using subprocess
func TestCLIErrorScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping error scenario tests in short mode")
	}

	helper := NewCLITestHelper(t)

	t.Run("invalid_custom_number", func(t *testing.T) {
		tempDir := t.TempDir()

		input := "5\nabc\n10m\nn\n"
		stdout, _, exitCode := helper.RunCLIWithInputInDir(input, tempDir)

		// Should still succeed but use default cache size (0)
		if exitCode != 0 {
			t.Fatalf("CLI should handle invalid input gracefully, got exit code %d", exitCode)
		}

		helper.AssertContains(stdout, "Generated metis.json successfully!")

		// Check that file was created in the temp directory
		if !helper.CheckFileExistsInDir("metis.json", tempDir) {
			t.Error("metis.json was not created")
			return
		}

		jsonData := helper.ReadJSONFileFromDir("metis.json", tempDir)
		if cacheSize, ok := jsonData["cache_size"]; !ok || cacheSize != float64(0) {
			t.Errorf("Invalid number should result in 0, got %v", cacheSize)
		}
	})
}

// BenchmarkCLISubprocess benchmarks subprocess CLI execution
func BenchmarkCLISubprocess(b *testing.B) {
	helper := &CLITestHelper{
		t:       &testing.T{},
		mainDir: func() string { dir, _ := os.Getwd(); return dir }(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tempDir := b.TempDir()

		_, _, exitCode := helper.RunCLIWithInputInDir("1\n", tempDir)
		if exitCode != 0 {
			b.Fatalf("CLI failed with exit code %d", exitCode)
		}
	}
}
