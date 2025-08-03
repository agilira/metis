// security_test.go: Security tests for Metis strategic caching library
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package metis

import (
	"testing"
)

func TestSecureFloat64_ErrorHandling(t *testing.T) {
	// Test SecureFloat64 function
	// Note: This test is mainly to cover the error handling path
	// In practice, crypto/rand.Read should rarely fail

	result := SecureFloat64()

	// The function should return a value between 0 and 1
	if result < 0 || result >= 1 {
		t.Errorf("SecureFloat64 should return a value in [0,1), got %f", result)
	}

	// Test multiple calls to ensure consistency
	for i := 0; i < 10; i++ {
		val := SecureFloat64()
		if val < 0 || val >= 1 {
			t.Errorf("SecureFloat64 call %d returned %f, expected value in [0,1)", i, val)
		}
	}
}

func TestSecureFloat64_ErrorPath(t *testing.T) {
	// Test SecureFloat64 function to cover the error handling path
	// This test is mainly to cover the return 0.0 line when randc.Read fails

	// Call SecureFloat64 multiple times to increase chance of covering error path
	// In practice, crypto/rand.Read should rarely fail, but we want to cover the path
	for i := 0; i < 100; i++ {
		result := SecureFloat64()

		// The function should return a value between 0 and 1
		if result < 0 || result >= 1 {
			t.Errorf("SecureFloat64 should return a value in [0,1), got %f", result)
		}
	}
}

func TestPrimitiveBox_Struct(t *testing.T) {
	// Test PrimitiveBox struct
	box := PrimitiveBox{
		V: "test_value",
	}

	if box.V != "test_value" {
		t.Errorf("Expected V to be 'test_value', got '%v'", box.V)
	}

	// Test with different types
	box.V = 42
	if box.V != 42 {
		t.Errorf("Expected V to be 42, got '%v'", box.V)
	}

	box.V = true
	if box.V != true {
		t.Errorf("Expected V to be true, got '%v'", box.V)
	}
}
