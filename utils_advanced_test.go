// utils_advanced_test.go: Additional tests to push toBytes
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package metis

import (
	"encoding/gob"
	"testing"
)

// TestToBytes_AllIntegerTypeCoverage tests all integer cases explicitly
func TestToBytes_AllIntegerTypeCoverage(t *testing.T) {
	testCases := []struct {
		name  string
		value interface{}
	}{
		// Ensure all integer types hit their specific switch cases
		{"int8_max", int8(127)},
		{"int8_min", int8(-128)},
		{"int8_zero", int8(0)},
		{"int16_max", int16(32767)},
		{"int16_min", int16(-32768)},
		{"int16_zero", int16(0)},
		{"uint8_max", uint8(255)},
		{"uint8_zero", uint8(0)},
		{"uint16_max", uint16(65535)},
		{"uint16_zero", uint16(0)},
		{"float32_positive", float32(123.456)},
		{"float32_negative", float32(-123.456)},
		{"float32_zero", float32(0.0)},
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

// TestToBytes_PrimitiveBoxWithComplexTypes tests PrimitiveBox recursion
func TestToBytes_PrimitiveBoxWithComplexTypes(t *testing.T) {
	// Register types for gob encoding to ensure successful serialization
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})

	testCases := []struct {
		name string
		box  PrimitiveBox
	}{
		{"box_with_map", PrimitiveBox{V: map[string]interface{}{"key": "value"}}},
		{"box_with_slice", PrimitiveBox{V: []interface{}{1, 2, 3}}},
		{"box_with_nested_box", PrimitiveBox{V: PrimitiveBox{V: "nested"}}},
		{"box_with_bool_true", PrimitiveBox{V: true}},
		{"box_with_bool_false", PrimitiveBox{V: false}},
		{"box_with_int", PrimitiveBox{V: 42}},
		{"box_with_float", PrimitiveBox{V: 3.14}},
		{"box_with_string", PrimitiveBox{V: "test"}},
		{"box_with_empty_string", PrimitiveBox{V: ""}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := toBytes(tc.box)
			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tc.name, err)
			}
			if len(result) == 0 && tc.box.V != "" {
				t.Errorf("Expected non-empty result for %s", tc.name)
			}
		})
	}
}

// TestToBytes_GobEncodingSuccessScenarios tests successful gob encoding
func TestToBytes_GobEncodingSuccessScenarios(t *testing.T) {
	// Register types for successful gob encoding
	gob.Register(map[string]string{})
	gob.Register([]string{})
	gob.Register(map[int]int{})

	testCases := []struct {
		name  string
		value interface{}
	}{
		{"map_string_string", map[string]string{"key1": "value1", "key2": "value2"}},
		{"slice_string", []string{"a", "b", "c"}},
		{"map_int_int", map[int]int{1: 10, 2: 20}},
		{"empty_map", map[string]string{}},
		{"empty_slice", []string{}},
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

// TestToBytes_BooleanEdgeCases tests both bool paths
func TestToBytes_BooleanEdgeCases(t *testing.T) {
	// Test true path
	resultTrue, err := toBytes(true)
	if err != nil {
		t.Errorf("Unexpected error for true: %v", err)
	}
	expectedTrue := []byte("true")
	if string(resultTrue) != string(expectedTrue) {
		t.Errorf("Expected %v for true, got %v", expectedTrue, resultTrue)
	}

	// Test false path
	resultFalse, err := toBytes(false)
	if err != nil {
		t.Errorf("Unexpected error for false: %v", err)
	}
	expectedFalse := []byte("false")
	if string(resultFalse) != string(expectedFalse) {
		t.Errorf("Expected %v for false, got %v", expectedFalse, resultFalse)
	}
}

// TestToBytes_BufferCopyMechanism tests the buffer copy logic in gob encoding
func TestToBytes_BufferCopyMechanism(t *testing.T) {
	// Register for successful encoding
	gob.Register(map[string]int{})

	// Create object that will use gob encoding path
	testObj := map[string]int{"test": 123}

	// Call toBytes multiple times to ensure buffer is properly copied each time
	for i := 0; i < 5; i++ {
		result, err := toBytes(testObj)
		if err != nil {
			t.Errorf("Unexpected error on iteration %d: %v", i, err)
		}
		if len(result) == 0 {
			t.Errorf("Expected non-empty result on iteration %d", i)
		}

		// Modify the result to ensure original buffer isn't affected
		if len(result) > 0 {
			originalFirst := result[0]
			result[0] = 255

			// Get another result and ensure it's not affected
			result2, err2 := toBytes(testObj)
			if err2 != nil {
				t.Errorf("Unexpected error on verification iteration %d: %v", i, err2)
			}
			if len(result2) > 0 && result2[0] == 255 {
				t.Errorf("Buffer reuse issue detected on iteration %d", i)
			}

			// Restore for next iteration
			result[0] = originalFirst
		}
	}
}

// TestToBytes_SpecificTypePaths ensures all switch case paths are hit
func TestToBytes_SpecificTypePaths(t *testing.T) {
	testCases := []struct {
		name     string
		value    interface{}
		typeName string
	}{
		{"[]byte_path", []byte{1, 2, 3}, "[]byte"},
		{"string_path", "test", "string"},
		{"int_path", int(42), "int"},
		{"int32_path", int32(42), "int32"},
		{"int64_path", int64(42), "int64"},
		{"uint_path", uint(42), "uint"},
		{"uint32_path", uint32(42), "uint32"},
		{"uint64_path", uint64(42), "uint64"},
		{"float32_path", float32(3.14), "float32"},
		{"float64_path", float64(3.14), "float64"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := toBytes(tc.value)
			if err != nil {
				t.Errorf("Unexpected error for %s (type %s): %v", tc.name, tc.typeName, err)
			}
			if len(result) == 0 {
				t.Errorf("Expected non-empty result for %s (type %s)", tc.name, tc.typeName)
			}
		})
	}
}

// TestToBytes_EdgeCaseMixedWithReal tests edge cases combined with real-world data
func TestToBytes_EdgeCaseMixedWithReal(t *testing.T) {
	// This ensures we hit different combinations of paths
	values := []interface{}{
		nil,                        // nil path
		[]byte{},                   // empty byte slice
		"",                         // empty string
		int(0),                     // zero int
		float64(0.0),               // zero float
		true,                       // bool true path
		false,                      // bool false path
		PrimitiveBox{V: nil},       // PrimitiveBox with nil (recursive call)
		PrimitiveBox{V: "content"}, // PrimitiveBox with content
	}

	for i, value := range values {
		t.Run(getTestName(i), func(t *testing.T) {
			result, err := toBytes(value)
			if err != nil {
				t.Errorf("Unexpected error for value %d: %v", i, err)
			}
			// Note: some values (like nil, empty string) should return empty results
			// This is expected behavior
			_ = result // Use the result variable
		})
	}
}

func getTestName(i int) string {
	names := []string{
		"nil_value",
		"empty_byte_slice",
		"empty_string",
		"zero_int",
		"zero_float",
		"true_bool",
		"false_bool",
		"primitive_box_nil",
		"primitive_box_content",
	}
	if i < len(names) {
		return names[i]
	}
	return "unknown"
}
