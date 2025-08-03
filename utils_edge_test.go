// utils_edge_test.go: Step 3.1 - Advanced tests for utils.go
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package metis

import (
	"math"
	"strconv"
	"testing"
	"unsafe"
)

// TestToBytes_UnsupportedTypes tests toBytes with types that cannot be serialized
func TestToBytes_UnsupportedTypes(t *testing.T) {
	testCases := []struct {
		name  string
		value interface{}
	}{
		{"channel", make(chan int)},
		{"function", func() int { return 42 }},
		{"unsafe_pointer", unsafe.Pointer(&[]int{1, 2, 3}[0])},
		{"complex_interface", interface{}(func() interface{} { return nil })},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := toBytes(tc.value)
			if err == nil {
				t.Errorf("Expected error for %s, but got nil", tc.name)
			}
			if result != nil {
				t.Errorf("Expected nil result for %s, but got %v", tc.name, result)
			}
		})
	}
}

// TestToBytes_AdditionalIntegerTypes tests toBytes with all integer types
func TestToBytes_AdditionalIntegerTypes(t *testing.T) {
	testCases := []struct {
		name  string
		value interface{}
	}{
		{"int8", int8(127)},
		{"int8_negative", int8(-128)},
		{"int16", int16(32767)},
		{"int16_negative", int16(-32768)},
		{"uint8", uint8(255)},
		{"uint16", uint16(65535)},
		{"float32", float32(3.14)},
		{"float32_negative", float32(-3.14)},
		{"float32_zero", float32(0.0)},
		{"float64_infinity", math.Inf(1)},
		{"float64_neg_infinity", math.Inf(-1)},
		{"float64_nan", math.NaN()},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := toBytes(tc.value)
			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tc.name, err)
			}
			if len(result) == 0 {
				t.Errorf("Expected non-empty result for %s", tc.name)
			}
		})
	}
}

// TestToBytes_PrimitiveBoxErrors tests PrimitiveBox with problematic values
func TestToBytes_PrimitiveBoxErrors(t *testing.T) {
	testCases := []struct {
		name string
		box  PrimitiveBox
	}{
		{"box_with_channel", PrimitiveBox{V: make(chan int)}},
		{"box_with_function", PrimitiveBox{V: func() {}}},
		{"box_with_unsafe_pointer", PrimitiveBox{V: unsafe.Pointer(&[]int{1}[0])}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := toBytes(tc.box)
			if err == nil {
				t.Errorf("Expected error for %s, but got nil", tc.name)
			}
			if result != nil {
				t.Errorf("Expected nil result for %s, but got %v", tc.name, result)
			}
		})
	}
}

// TestToBytes_PrimitiveBoxRecursion tests PrimitiveBox recursion
func TestToBytes_PrimitiveBoxRecursion(t *testing.T) {
	// Test nested PrimitiveBox
	innerBox := PrimitiveBox{V: "test"}
	outerBox := PrimitiveBox{V: innerBox}

	result, err := toBytes(outerBox)
	if err != nil {
		t.Errorf("Unexpected error for nested PrimitiveBox: %v", err)
	}
	if len(result) == 0 {
		t.Error("Expected non-empty result for nested PrimitiveBox")
	}

	// Test PrimitiveBox with nil value
	nilBox := PrimitiveBox{V: nil}
	result, err = toBytes(nilBox)
	if err != nil {
		t.Errorf("Unexpected error for PrimitiveBox with nil: %v", err)
	}
	if len(result) != 0 {
		t.Error("Expected empty result for PrimitiveBox with nil")
	}
}

// TestToBytes_GobEncodingFailures tests gob encoding failure scenarios
func TestToBytes_GobEncodingFailures(t *testing.T) {
	// Create a type that will cause gob encoding to fail
	type problemType struct {
		BadField chan int // channels can't be encoded by gob
	}

	badValue := problemType{BadField: make(chan int)}

	result, err := toBytes(badValue)
	if err == nil {
		t.Error("Expected error for gob encoding failure, but got nil")
	}
	if result != nil {
		t.Errorf("Expected nil result for gob encoding failure, but got %v", result)
	}
}

// TestParsePrimitiveFromString_NumericOverflows tests numeric overflow scenarios
func TestParsePrimitiveFromString_NumericOverflows(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected interface{}
		success  bool
	}{
		{"max_int64", strconv.FormatInt(math.MaxInt64, 10), int64(math.MaxInt64), true},
		{"min_int64", strconv.FormatInt(math.MinInt64, 10), int64(math.MinInt64), true},
		{"max_uint64", strconv.FormatUint(math.MaxUint64, 10), uint64(math.MaxUint64), true},
		{"int64_overflow", "9223372036854775808", uint64(9223372036854775808), true},     // Will parse as uint64
		{"huge_number", "999999999999999999999999999999999999999", float64(1e+39), true}, // Will parse as float
		{"float_max", "1.7976931348623157e+308", 1.7976931348623157e+308, true},
		{"float_smallest", "2.2250738585072014e-308", 2.2250738585072014e-308, true},
		{"float_infinity", "inf", math.Inf(1), true}, // Will parse as float64 infinity
		{"float_negative_infinity", "-inf", math.Inf(-1), true},
		{"float_nan", "nan", true, true}, // Will parse as float64 NaN (special case - check for NaN)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, success := parsePrimitiveFromString(tc.input)
			if success != tc.success {
				t.Errorf("Expected success=%v, got %v for %s", tc.success, success, tc.name)
			}
			// For successful parses, check the type or value
			if success && tc.expected != nil {
				switch expected := tc.expected.(type) {
				case string:
					if result != expected {
						t.Errorf("Expected string %s, got %v for %s", expected, result, tc.name)
					}
				case int64:
					if resultInt, ok := result.(int64); !ok || resultInt != expected {
						t.Errorf("Expected int64 %d, got %v for %s", expected, result, tc.name)
					}
				case uint64:
					if resultUint, ok := result.(uint64); !ok || resultUint != expected {
						t.Errorf("Expected uint64 %d, got %v for %s", expected, result, tc.name)
					}
				case float64:
					if resultFloat, ok := result.(float64); !ok || resultFloat != expected {
						t.Errorf("Expected float64 %f, got %v for %s", expected, result, tc.name)
					}
				case bool:
					// Special case for NaN test
					if tc.name == "float_nan" {
						if resultFloat, ok := result.(float64); !ok || !math.IsNaN(resultFloat) {
							t.Errorf("Expected NaN float64, got %v for %s", result, tc.name)
						}
					}
				}
			}
		})
	}
}

// TestParsePrimitiveFromString_EmptyAndSpecial tests empty strings and special characters
func TestParsePrimitiveFromString_EmptyAndSpecial(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected interface{}
		success  bool
	}{
		{"empty_string", "", "", true},
		{"whitespace_only", "   ", "   ", true},
		{"tab_newline", "\t\n", "\t\n", true},
		{"unicode_string", "🚀", "🚀", true},
		{"zero_string", "0", int64(0), true},
		{"negative_zero", "-0", int64(0), true},
		{"positive_sign", "+42", int64(42), true},
		{"float_zero", "0.0", float64(0), true},
		{"scientific_notation", "1e5", float64(100000), true},
		{"scientific_negative", "1e-5", float64(0.00001), true},
		{"hex_like_string", "0x42", "0x42", true},      // Not parsed as hex, returned as string
		{"binary_like_string", "0b101", "0b101", true}, // Not parsed as binary, returned as string
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, success := parsePrimitiveFromString(tc.input)
			if success != tc.success {
				t.Errorf("Expected success=%v, got %v for %s", tc.success, success, tc.name)
			}
			if success {
				switch expected := tc.expected.(type) {
				case string:
					if result != expected {
						t.Errorf("Expected string %q, got %v for %s", expected, result, tc.name)
					}
				case int64:
					if resultInt, ok := result.(int64); !ok || resultInt != expected {
						t.Errorf("Expected int64 %d, got %v for %s", expected, result, tc.name)
					}
				case float64:
					if resultFloat, ok := result.(float64); !ok || resultFloat != expected {
						t.Errorf("Expected float64 %f, got %v for %s", expected, result, tc.name)
					}
				}
			}
		})
	}
}

// TestParsePrimitiveFromString_NumericEdgeCases tests edge cases in numeric parsing
func TestParsePrimitiveFromString_NumericEdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected interface{}
		success  bool
	}{
		{"leading_zeros", "000042", int64(42), true},
		{"leading_zeros_float", "000042.5", float64(42.5), true},
		{"just_minus", "-", "-", true},                                                                // Will return as string
		{"just_plus", "+", "+", true},                                                                 // Will return as string
		{"just_dot", ".", ".", true},                                                                  // Will return as string
		{"multiple_dots", "1.2.3", "1.2.3", true},                                                     // Will return as string
		{"number_with_spaces", "42 ", "42 ", true},                                                    // Will return as string (trailing space)
		{"spaces_in_number", "4 2", "4 2", true},                                                      // Will return as string
		{"very_long_number", "123456789012345678901234567890", float64(1.2345678901234568e+29), true}, // Too big, will parse as float64
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, success := parsePrimitiveFromString(tc.input)
			if success != tc.success {
				t.Errorf("Expected success=%v, got %v for %s", tc.success, success, tc.name)
			}
			if success {
				switch expected := tc.expected.(type) {
				case string:
					if result != expected {
						t.Errorf("Expected string %q, got %v for %s", expected, result, tc.name)
					}
				case int64:
					if resultInt, ok := result.(int64); !ok || resultInt != expected {
						t.Errorf("Expected int64 %d, got %v for %s", expected, result, tc.name)
					}
				case float64:
					if resultFloat, ok := result.(float64); !ok || resultFloat != expected {
						t.Errorf("Expected float64 %f, got %v for %s", expected, result, tc.name)
					}
				}
			}
		})
	}
}

// TestToBytes_BufferReuseIssues tests buffer reuse scenarios
func TestToBytes_BufferReuseIssues(t *testing.T) {
	// Create multiple simple objects that can be serialized successfully
	objects := []interface{}{
		map[string]string{"key1": "value1"},
		map[string]string{"key2": "value2"},
		map[string]string{"key3": "value3"},
	}

	results := make([][]byte, len(objects))

	// Convert all objects to bytes (this should test buffer reuse)
	for i, obj := range objects {
		result, err := toBytes(obj)
		if err != nil {
			t.Errorf("Unexpected error for object %d: %v", i, err)
		}
		results[i] = result
	}

	// Verify that results are independent (no buffer corruption)
	for i, result := range results {
		if len(result) == 0 {
			t.Errorf("Expected non-empty result for object %d", i)
		}

		// Results should be different for different objects
		for j, otherResult := range results {
			if i != j && len(result) == len(otherResult) {
				equal := true
				for k := range result {
					if result[k] != otherResult[k] {
						equal = false
						break
					}
				}
				if equal {
					t.Errorf("Results for objects %d and %d should be different", i, j)
				}
			}
		}
	}
}

// TestToBytes_ConcurrencyBufferSafety tests buffer safety under concurrent access
func TestToBytes_ConcurrencyBufferSafety(t *testing.T) {
	// This test ensures that concurrent calls to toBytes don't corrupt each other's buffers
	const numGoroutines = 10
	const numIterations = 100

	done := make(chan bool, numGoroutines)

	for g := 0; g < numGoroutines; g++ {
		go func(goroutineID int) {
			defer func() { done <- true }()

			for i := 0; i < numIterations; i++ {
				// Use simple map that can be serialized
				testData := map[string]int{
					"id":   goroutineID*numIterations + i,
					"data": goroutineID + i,
				}

				result, err := toBytes(testData)
				if err != nil {
					t.Errorf("Unexpected error in goroutine %d, iteration %d: %v", goroutineID, i, err)
					return
				}

				if len(result) == 0 {
					t.Errorf("Expected non-empty result in goroutine %d, iteration %d", goroutineID, i)
					return
				}
			}
		}(g)
	}

	// Wait for all goroutines to complete
	for g := 0; g < numGoroutines; g++ {
		<-done
	}
}
