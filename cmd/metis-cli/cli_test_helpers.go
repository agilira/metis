// /cmd/metis-cli/cli_test_helpers.go: Helper functions for CLI testing
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// CLITestHelper provides utilities for testing the CLI application
type CLITestHelper struct {
	t       *testing.T
	mainDir string // Directory where main.go is located
}

// NewCLITestHelper creates a new CLI test helper
func NewCLITestHelper(t *testing.T) *CLITestHelper {
	mainDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	return &CLITestHelper{
		t:       t,
		mainDir: mainDir,
	}
}

// RunCLIWithInput runs the CLI with simulated user input using subprocess
func (h *CLITestHelper) RunCLIWithInput(input string) (string, string, int) {
	// Create temporary directory for test execution
	tempDir := h.t.TempDir()

	// Build command to run CLI from the main directory but execute in temp directory
	// This ensures metis.json is created in the temp directory
	mainPath := filepath.Join(h.mainDir, "main.go")
	// Validate path to prevent command injection (only check for actual shell metacharacters)
	// On Windows, \ and : are normal path characters, so we exclude them from the check
	if strings.ContainsAny(mainPath, ";&|`$(){}[]<>\"'") {
		h.t.Fatalf("Invalid path containing shell metacharacters: %s", mainPath)
	}
	//nolint:gosec -- This is a test helper, mainPath is validated above and contains only safe characters
	cmd := exec.Command("go", "run", mainPath)
	cmd.Dir = tempDir // Run from temp directory
	cmd.Stdin = strings.NewReader(input)

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	exitCode := 0

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			// If it's not an ExitError, it's likely a build/execution error
			exitCode = 1
		}
	}

	return string(output), "", exitCode
}

// RunCLIWithInputInDir runs CLI in a specific directory
func (h *CLITestHelper) RunCLIWithInputInDir(input string, workDir string) (string, string, int) {
	mainPath := filepath.Join(h.mainDir, "main.go")
	// Validate path to prevent command injection (only check for actual shell metacharacters)
	// On Windows, \ and : are normal path characters, so we exclude them from the check
	if strings.ContainsAny(mainPath, ";&|`$(){}[]<>\"'") {
		h.t.Fatalf("Invalid path containing shell metacharacters: %s", mainPath)
	}
	//nolint:gosec -- This is a test helper, mainPath is validated above and contains only safe characters
	cmd := exec.Command("go", "run", mainPath)
	cmd.Dir = workDir
	cmd.Stdin = strings.NewReader(input)

	output, err := cmd.CombinedOutput()
	exitCode := 0

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
	}

	return string(output), "", exitCode
}

// CheckFileExists verifies that a file exists in the current working directory
func (h *CLITestHelper) CheckFileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// CheckFileExistsInDir verifies that a file exists in a specific directory
func (h *CLITestHelper) CheckFileExistsInDir(filename string, dir string) bool {
	_, err := os.Stat(filepath.Join(dir, filename))
	return !os.IsNotExist(err)
}

// ReadJSONFile reads and validates a JSON file
func (h *CLITestHelper) ReadJSONFile(filename string) map[string]interface{} {
	data, err := os.ReadFile(filename)
	if err != nil {
		h.t.Fatalf("Failed to read %s: %v", filename, err)
	}

	var jsonData map[string]interface{}
	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		h.t.Fatalf("Failed to parse JSON from %s: %v", filename, err)
	}

	return jsonData
}

// ReadJSONFileFromDir reads and validates a JSON file from a specific directory
func (h *CLITestHelper) ReadJSONFileFromDir(filename string, dir string) map[string]interface{} {
	data, err := os.ReadFile(filepath.Join(dir, filename))
	if err != nil {
		h.t.Fatalf("Failed to read %s from %s: %v", filename, dir, err)
	}

	var jsonData map[string]interface{}
	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		h.t.Fatalf("Failed to parse JSON from %s: %v", filename, err)
	}

	return jsonData
}

// AssertValidJSON validates that a string contains valid JSON
func (h *CLITestHelper) AssertValidJSON(jsonStr string) map[string]interface{} {
	var jsonData map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &jsonData)
	if err != nil {
		h.t.Fatalf("Invalid JSON: %v", err)
	}
	return jsonData
}

// AssertContains checks if a string contains a substring
func (h *CLITestHelper) AssertContains(text, substring string) {
	if !strings.Contains(text, substring) {
		h.t.Errorf("Expected text to contain '%s', but it didn't.\nFull text: %s", substring, text)
	}
}

// AssertNotContains checks if a string does not contain a substring
func (h *CLITestHelper) AssertNotContains(text, substring string) {
	if strings.Contains(text, substring) {
		h.t.Errorf("Expected text NOT to contain '%s', but it did.\nFull text: %s", substring, text)
	}
}
