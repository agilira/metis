// cli_test_helpers.go: Helper functions for CLI testing
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// CLITestHelper provides utilities for testing CLI applications
type CLITestHelper struct {
	t *testing.T
}

// NewCLITestHelper creates a new CLI test helper
func NewCLITestHelper(t *testing.T) *CLITestHelper {
	return &CLITestHelper{t: t}
}

// RunCommand runs the CLI as a subprocess and captures output
func (h *CLITestHelper) RunCommand(args ...string) (stdout, stderr string, exitCode int) {
	// Validate args to prevent command injection (only check for actual shell metacharacters)
	// On Windows, \ and : are normal path characters, so we exclude them from the check
	for _, arg := range args {
		if strings.ContainsAny(arg, ";&|`$(){}[]<>\"'") {
			h.t.Fatalf("Invalid argument containing shell metacharacters: %s", arg)
		}
	}
	//nolint:gosec -- This is a test helper, args are controlled test inputs and main.go is a fixed filename
	cmd := exec.Command("go", append([]string{"run", "main.go"}, args...)...)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
	}

	return outBuf.String(), errBuf.String(), exitCode
}

// AssertContains checks if output contains expected string
func (h *CLITestHelper) AssertContains(output, expected string) {
	h.t.Helper()
	if !strings.Contains(output, expected) {
		h.t.Errorf("Expected output to contain '%s', got: %s", expected, output)
	}
}

// AssertNotContains checks if output does not contain a string
func (h *CLITestHelper) AssertNotContains(output, notExpected string) {
	h.t.Helper()
	if strings.Contains(output, notExpected) {
		h.t.Errorf("Expected output to NOT contain '%s', got: %s", notExpected, output)
	}
}

// AssertValidJSON checks if output is valid JSON
func (h *CLITestHelper) AssertValidJSON(output string) map[string]interface{} {
	h.t.Helper()
	var jsonData map[string]interface{}
	err := json.Unmarshal([]byte(output), &jsonData)
	if err != nil {
		h.t.Fatalf("Invalid JSON output: %v\nOutput: %s", err, output)
	}
	return jsonData
}

// AssertExitCode checks if command exited with expected code
func (h *CLITestHelper) AssertExitCode(actual, expected int) {
	h.t.Helper()
	if actual != expected {
		h.t.Errorf("Expected exit code %d, got %d", expected, actual)
	}
}

// TempFile creates a temporary file for testing
func (h *CLITestHelper) TempFile(content string) string {
	h.t.Helper()
	file, err := os.CreateTemp("", "metis-test-*")
	if err != nil {
		h.t.Fatalf("Failed to create temp file: %v", err)
	}

	if content != "" {
		_, err = file.WriteString(content)
		if err != nil {
			h.t.Fatalf("Failed to write to temp file: %v", err)
		}
	}

	if err := file.Close(); err != nil {
		h.t.Fatalf("Failed to close temp file: %v", err)
	}
	h.t.Cleanup(func() {
		if err := os.Remove(file.Name()); err != nil {
			h.t.Logf("Failed to remove temp file: %v", err)
		}
	})

	return file.Name()
}

// SetEnv temporarily sets an environment variable
func (h *CLITestHelper) SetEnv(key, value string) {
	h.t.Helper()
	oldValue := os.Getenv(key)
	if err := os.Setenv(key, value); err != nil {
		h.t.Fatalf("Failed to set environment variable: %v", err)
	}
	h.t.Cleanup(func() {
		if oldValue == "" {
			if err := os.Unsetenv(key); err != nil {
				h.t.Logf("Failed to unset environment variable: %v", err)
			}
		} else {
			if err := os.Setenv(key, oldValue); err != nil {
				h.t.Logf("Failed to restore environment variable: %v", err)
			}
		}
	})
}
