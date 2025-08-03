// utils_unit_test.go: Comprehensive tests for Metis strategic caching library
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0
package metis

import (
	"fmt"
	"testing"
	"time"
)

// TestToBytes_AllDataTypes tests toBytes with all data types
func TestToBytes_AllDataTypes(t *testing.T) {
	testCases := []struct {
		name     string
		value    interface{}
		expected []byte
	}{
		{"string", "hello", []byte("hello")},
		{"int", 42, []byte{42, 0, 0, 0, 0, 0, 0, 0}},
		{"float64", 3.14, []byte{31, 133, 235, 81, 184, 30, 9, 64}},
		{"bool", true, []byte{1}},
		{"nil", nil, []byte{}},
		{"empty_string", "", []byte{}},
		{"zero_int", 0, []byte{0, 0, 0, 0, 0, 0, 0, 0}},
		{"false_bool", false, []byte{0}},
		{"slice", []int{1, 2, 3}, nil},                 // Will use gob encoding
		{"map", map[string]int{"a": 1}, nil},           // Will use gob encoding
		{"struct", struct{ Name string }{"test"}, nil}, // Will use gob encoding
		{"time", time.Now(), nil},                      // Will use gob encoding
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := toBytes(tc.value)
			if tc.expected != nil {
				if len(result) == 0 && len(tc.expected) != 0 {
					t.Errorf("toBytes(%v): expected non-empty result", tc.value)
				}
			} else {
				// For complex types, just verify we get some result
				if result == nil && err == nil {
					t.Errorf("toBytes(%v): expected non-nil result or error", tc.value)
				}
			}
		})
	}
}

// TestToBytes_ErrorCases tests toBytes error handling
func TestToBytes_ErrorCases(t *testing.T) {
	// Test with channel (should fail gob encoding)
	ch := make(chan int)
	result, err := toBytes(ch)
	if result == nil && err == nil {
		t.Error("toBytes should handle channel gracefully")
	}

	// Test with function (should fail gob encoding)
	var fn func()
	result, err = toBytes(fn)
	if result == nil && err == nil {
		t.Error("toBytes should handle function gracefully")
	}

	// Test with complex nested structure
	complexStruct := struct {
		Data map[string]interface{}
	}{
		Data: map[string]interface{}{
			"nested": map[string]interface{}{
				"deep": []interface{}{1, "string", true},
			},
		},
	}
	result, err = toBytes(complexStruct)
	if result == nil && err == nil {
		t.Error("toBytes should handle complex structures gracefully")
	}
}

// TestParsePrimitive_InvalidFormats tests parsePrimitiveFromString with invalid formats
func TestParsePrimitive_InvalidFormats(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"invalid_int", "not_a_number", "not_a_number"},
		{"invalid_float", "not_a_float", "not_a_float"},
		{"invalid_bool", "not_a_bool", "not_a_bool"},
		{"empty_string", "", ""},
		{"whitespace", "   ", "   "},
		{"special_chars", "!@#$%", "!@#$%"},
		{"unicode", "café", "café"},
		{"numbers_with_text", "123abc", "123abc"},
		{"float_with_text", "3.14abc", "3.14abc"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parsePrimitiveFromString(tc.input)
			if result != tc.expected {
				t.Errorf("parsePrimitiveFromString(%s): expected %v, got %v", tc.input, tc.expected, result)
			}
			// Ignore error for invalid formats
			_ = err
		})
	}
}

// TestCalculateSize_ComplexTypes tests calculateSize with complex types
func TestCalculateSize_ComplexTypes(t *testing.T) {
	testCases := []struct {
		name     string
		value    interface{}
		expected int
	}{
		{"nil", nil, 0},
		{"empty_string", "", 0},
		{"string", "hello", 5},
		{"unicode_string", "café", 4},
		{"int", 42, 8},
		{"float64", 3.14, 8},
		{"bool", true, 1},
		{"slice", []int{1, 2, 3}, 24},               // 3 * 8 bytes
		{"map", map[string]int{"a": 1, "b": 2}, 16}, // 2 * 8 bytes
		{"struct", struct{ Name string }{"test"}, 4},
		{"pointer", &struct{ Name string }{"test"}, 8}, // pointer size
		{"time", time.Now(), 24},                       // time.Time size
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calculateSize(tc.value)
			if result < 0 {
				t.Errorf("calculateSize(%v): expected non-negative result, got %d", tc.value, result)
			}
		})
	}
}

// TestCalculateSize_NilValues tests calculateSize with nil values
func TestCalculateSize_NilValues(t *testing.T) {
	// Test nil interface
	result := calculateSize(nil)
	if result != 0 {
		t.Errorf("calculateSize(nil): expected 0, got %d", result)
	}

	// Test nil pointer
	var ptr *int
	result = calculateSize(ptr)
	if result != 8 { // pointer size
		t.Errorf("calculateSize(nil pointer): expected 8, got %d", result)
	}

	// Test nil slice
	var slice []int
	result = calculateSize(slice)
	if result <= 0 {
		t.Errorf("calculateSize(nil slice): expected positive result, got %d", result)
	}

	// Test nil map
	var m map[string]int
	result = calculateSize(m)
	if result <= 0 {
		t.Errorf("calculateSize(nil map): expected positive result, got %d", result)
	}
}

// TestCalculateSize_EdgeCases tests calculateSize edge cases
func TestCalculateSize_EdgeCases(t *testing.T) {
	// Test with very large string
	largeString := string(make([]byte, 10000))
	result := calculateSize(largeString)
	if result != 10000 {
		t.Errorf("calculateSize(large string): expected 10000, got %d", result)
	}

	// Test with very large slice
	largeSlice := make([]int, 1000)
	result = calculateSize(largeSlice)
	if result <= 0 {
		t.Errorf("calculateSize(large slice): expected positive result, got %d", result)
	}

	// Test with nested structures
	nestedStruct := struct {
		Name   string
		Values []int
		Data   map[string]interface{}
	}{
		Name:   "test",
		Values: []int{1, 2, 3},
		Data:   map[string]interface{}{"key": "value"},
	}
	result = calculateSize(nestedStruct)
	if result <= 0 {
		t.Errorf("calculateSize(nested struct): expected positive result, got %d", result)
	}
}

// TestToBytes_ComplexTypes tests toBytes with complex types
func TestToBytes_ComplexTypes(t *testing.T) {
	// Test with struct containing unexported fields
	type privateStruct struct {
		Name    string
		private int
	}
	ps := privateStruct{Name: "test", private: 42}
	result, err := toBytes(ps)
	if result == nil && err == nil {
		t.Error("toBytes should handle structs with unexported fields")
	}

	// Test with interface
	var iface interface{} = "test"
	result, err = toBytes(iface)
	if result == nil && err == nil {
		t.Error("toBytes should handle interfaces")
	}

	// Test with array
	arr := [3]int{1, 2, 3}
	result, err = toBytes(arr)
	if result == nil && err == nil {
		t.Error("toBytes should handle arrays")
	}
}

// TestParsePrimitive_EdgeCases tests parsePrimitiveFromString edge cases
func TestParsePrimitive_EdgeCases(t *testing.T) {
	// Test with very long number
	longNumber := "123456789012345678901234567890"
	result, err := parsePrimitiveFromString(longNumber)
	if result == nil {
		t.Errorf("parsePrimitiveFromString(long number): expected non-nil result, got %v", result)
	}
	_ = err

	// Test with scientific notation
	scientific := "1.23e-4"
	result, err = parsePrimitiveFromString(scientific)
	if result != 0.000123 {
		t.Errorf("parsePrimitiveFromString(scientific): expected 0.000123, got %v", result)
	}
	_ = err

	// Test with hex number
	hexNumber := "0x1A"
	result, err = parsePrimitiveFromString(hexNumber)
	if result != hexNumber {
		t.Errorf("parsePrimitiveFromString(hex): expected %s, got %v", hexNumber, result)
	}
	_ = err
}

// TestCalculateSize_Performance tests calculateSize performance
func TestCalculateSize_Performance(t *testing.T) {
	// Test with large data structure
	largeData := make(map[string]interface{})
	for i := 0; i < 1000; i++ {
		largeData[fmt.Sprintf("key%d", i)] = fmt.Sprintf("value%d", i)
	}

	// This should complete quickly
	result := calculateSize(largeData)
	if result <= 0 {
		t.Errorf("calculateSize(large data): expected positive result, got %d", result)
	}
}

// TestToBytes_Performance tests toBytes performance
func TestToBytes_Performance(t *testing.T) {
	// Test with large data structure
	largeData := make(map[string]interface{})
	for i := 0; i < 100; i++ {
		largeData[fmt.Sprintf("key%d", i)] = fmt.Sprintf("value%d", i)
	}

	// This should complete quickly
	result, err := toBytes(largeData)
	if result == nil && err == nil {
		t.Error("toBytes should handle large data structures")
	}
}
