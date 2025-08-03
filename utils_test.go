// utils_test.go: Unit tests for utility functions in Metis strategic caching library
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package metis

import (
	"testing"
)

func TestCalculateSizeAdditional(t *testing.T) {
	// Test nil value
	if size := calculateSize(nil); size != 0 {
		t.Errorf("Expected size 0 for nil, got %d", size)
	}

	// Test string
	if size := calculateSize("hello"); size != 5 {
		t.Errorf("Expected size 5 for string, got %d", size)
	}

	// Test byte slice
	if size := calculateSize([]byte{1, 2, 3}); size != 3 {
		t.Errorf("Expected size 3 for byte slice, got %d", size)
	}

	// Test integers
	if size := calculateSize(42); size <= 0 {
		t.Errorf("Expected positive size for int, got %d", size)
	}

	// Test unsigned integers
	if size := calculateSize(uint(42)); size <= 0 {
		t.Errorf("Expected positive size for uint, got %d", size)
	}

	// Test floats
	if size := calculateSize(3.14); size <= 0 {
		t.Errorf("Expected positive size for float64, got %d", size)
	}

	// Test bool
	if size := calculateSize(true); size <= 0 {
		t.Errorf("Expected positive size for bool, got %d", size)
	}

	// Test PrimitiveBox
	box := PrimitiveBox{V: "test"}
	if size := calculateSize(box); size <= 0 {
		t.Errorf("Expected positive size for PrimitiveBox, got %d", size)
	}

	// Test nil pointer
	var nilPtr *int
	if size := calculateSize(nilPtr); size != 8 {
		t.Errorf("Expected size 8 for nil pointer, got %d", size)
	}

	// Test complex types that trigger reflection
	type TestStructAdditional struct {
		Field1 string
		Field2 int
	}
	testStruct := TestStructAdditional{Field1: "hello", Field2: 42}
	size := calculateSize(testStruct)
	if size <= 0 {
		t.Errorf("Expected positive size for struct, got %d", size)
	}

	// Test slice
	slice := []string{"a", "b", "c"}
	size = calculateSize(slice)
	if size <= 0 {
		t.Errorf("Expected positive size for slice, got %d", size)
	}

	// Test map
	testMap := map[string]int{"a": 1, "b": 2}
	size = calculateSize(testMap)
	if size <= 0 {
		t.Errorf("Expected positive size for map, got %d", size)
	}

	// Test pointer
	ptr := &testStruct
	size = calculateSize(ptr)
	if size <= 0 {
		t.Errorf("Expected positive size for pointer, got %d", size)
	}
}

func TestToBytesAdditional(t *testing.T) {
	// Test nil value
	bytes, err := toBytes(nil)
	if err != nil {
		t.Errorf("Expected no error for nil value, got %v", err)
	}
	if len(bytes) != 0 {
		t.Errorf("Expected empty slice for nil value, got %v", bytes)
	}

	// Test byte slice
	input := []byte{1, 2, 3}
	bytes, err = toBytes(input)
	if err != nil {
		t.Errorf("Expected no error for byte slice, got %v", err)
	}
	if len(bytes) != 3 {
		t.Errorf("Expected length 3, got %d", len(bytes))
	}

	// Test string
	inputStr := "hello"
	bytes, err = toBytes(inputStr)
	if err != nil {
		t.Errorf("Expected no error for string, got %v", err)
	}
	if string(bytes) != inputStr {
		t.Errorf("Expected %s, got %s", inputStr, string(bytes))
	}

	// Test PrimitiveBox
	box := PrimitiveBox{V: "test"}
	bytes, err = toBytes(box)
	if err != nil {
		t.Errorf("Expected no error for PrimitiveBox, got %v", err)
	}
	if len(bytes) == 0 {
		t.Error("Expected non-empty bytes for PrimitiveBox")
	}

	// Test complex type that triggers gob encoding
	testStruct := TestStructAdditional{Field: "test"}
	bytes, err = toBytes(testStruct)
	if err == nil {
		t.Error("Expected error for unregistered struct type")
	}
	if len(bytes) != 0 {
		t.Error("Expected empty bytes for unregistered struct type")
	}

	// Test types that can't be serialized with gob
	ch := make(chan int)
	bytes, err = toBytes(ch)
	if err == nil {
		t.Error("Expected error for channel (non-serializable type)")
	}
	if len(bytes) != 0 {
		t.Error("Expected empty bytes for non-serializable type")
	}

	// Test function (non-serializable)
	testFunc := func() int { return 42 }
	bytes, err = toBytes(testFunc)
	if err == nil {
		t.Error("Expected error for function (non-serializable type)")
	}
	if len(bytes) != 0 {
		t.Error("Expected empty bytes for non-serializable type")
	}
}

// TestStructAdditional is used for testing serialization
type TestStructAdditional struct {
	Field string
}

func TestCalculateSizeEdgeCases(t *testing.T) {
	// Test with very large string
	largeString := string(make([]byte, 10000))
	size := calculateSize(largeString)
	if size != 10000 {
		t.Errorf("Expected size 10000 for large string, got %d", size)
	}

	// Test with empty string
	if size := calculateSize(""); size != 0 {
		t.Errorf("Expected size 0 for empty string, got %d", size)
	}

	// Test with empty slice
	if size := calculateSize([]byte{}); size != 0 {
		t.Errorf("Expected size 0 for empty slice, got %d", size)
	}

	// Test with zero values
	if size := calculateSize(0); size <= 0 {
		t.Errorf("Expected positive size for zero int, got %d", size)
	}

	if size := calculateSize(uint(0)); size <= 0 {
		t.Errorf("Expected positive size for zero uint, got %d", size)
	}

	if size := calculateSize(0.0); size <= 0 {
		t.Errorf("Expected positive size for zero float, got %d", size)
	}

	if size := calculateSize(false); size <= 0 {
		t.Errorf("Expected positive size for false bool, got %d", size)
	}
}

func TestToBytesEdgeCases(t *testing.T) {
	// Test with very large string
	largeString := string(make([]byte, 10000))
	bytes, err := toBytes(largeString)
	if err != nil {
		t.Errorf("Expected no error for large string, got %v", err)
	}
	if len(bytes) != 10000 {
		t.Errorf("Expected length 10000 for large string, got %d", len(bytes))
	}

	// Test with empty string
	bytes, err = toBytes("")
	if err != nil {
		t.Errorf("Expected no error for empty string, got %v", err)
	}
	if len(bytes) != 0 {
		t.Errorf("Expected empty bytes for empty string, got %d", len(bytes))
	}

	// Test with empty slice
	bytes, err = toBytes([]byte{})
	if err != nil {
		t.Errorf("Expected no error for empty slice, got %v", err)
	}
	if len(bytes) != 0 {
		t.Errorf("Expected empty bytes for empty slice, got %d", len(bytes))
	}

	// Test with zero values
	bytes, err = toBytes(0)
	if err != nil {
		t.Errorf("Expected no error for zero int, got %v", err)
	}
	if len(bytes) == 0 {
		t.Error("Expected non-empty bytes for zero int")
	}

	bytes, err = toBytes(uint(0))
	if err != nil {
		t.Errorf("Expected no error for zero uint, got %v", err)
	}
	if len(bytes) == 0 {
		t.Error("Expected non-empty bytes for zero uint")
	}

	bytes, err = toBytes(0.0)
	if err != nil {
		t.Errorf("Expected no error for zero float, got %v", err)
	}
	if len(bytes) == 0 {
		t.Error("Expected non-empty bytes for zero float")
	}

	bytes, err = toBytes(false)
	if err != nil {
		t.Errorf("Expected no error for false bool, got %v", err)
	}
	if len(bytes) == 0 {
		t.Error("Expected non-empty bytes for false bool")
	}
}

func TestCalculateSizeComplexTypes(t *testing.T) {
	// Test with nested structs
	type NestedStruct struct {
		Field string
	}
	type ComplexStruct struct {
		Field1 string
		Field2 int
		Field3 NestedStruct
		Field4 []string
		Field5 map[string]int
	}

	complexStruct := ComplexStruct{
		Field1: "hello",
		Field2: 42,
		Field3: NestedStruct{Field: "nested"},
		Field4: []string{"a", "b", "c"},
		Field5: map[string]int{"x": 1, "y": 2},
	}

	size := calculateSize(complexStruct)
	if size <= 0 {
		t.Errorf("Expected positive size for complex struct, got %d", size)
	}

	// Test with slice of complex types
	slice := []ComplexStruct{complexStruct, complexStruct}
	size = calculateSize(slice)
	if size <= 0 {
		t.Errorf("Expected positive size for slice of complex structs, got %d", size)
	}

	// Test with map of complex types
	testMap := map[string]ComplexStruct{"key1": complexStruct, "key2": complexStruct}
	size = calculateSize(testMap)
	if size <= 0 {
		t.Errorf("Expected positive size for map of complex structs, got %d", size)
	}

	// Test with pointer to complex struct
	ptr := &complexStruct
	size = calculateSize(ptr)
	if size <= 0 {
		t.Errorf("Expected positive size for pointer to complex struct, got %d", size)
	}
}

func TestToBytesComplexTypes(t *testing.T) {
	// Test with nested structs
	type NestedStruct struct {
		Field string
	}
	type ComplexStruct struct {
		Field1 string
		Field2 int
		Field3 NestedStruct
		Field4 []string
		Field5 map[string]int
	}

	complexStruct := ComplexStruct{
		Field1: "hello",
		Field2: 42,
		Field3: NestedStruct{Field: "nested"},
		Field4: []string{"a", "b", "c"},
		Field5: map[string]int{"x": 1, "y": 2},
	}

	bytes, err := toBytes(complexStruct)
	if err == nil {
		t.Error("Expected error for unregistered complex struct type")
	}
	if len(bytes) != 0 {
		t.Error("Expected empty bytes for unregistered complex struct type")
	}

	// Test with slice of complex types
	slice := []ComplexStruct{complexStruct, complexStruct}
	bytes, err = toBytes(slice)
	if err == nil {
		t.Error("Expected error for slice of unregistered complex structs")
	}
	if len(bytes) != 0 {
		t.Error("Expected empty bytes for slice of unregistered complex structs")
	}

	// Test with map of complex types
	testMap := map[string]ComplexStruct{"key1": complexStruct, "key2": complexStruct}
	bytes, err = toBytes(testMap)
	if err == nil {
		t.Error("Expected error for map of unregistered complex structs")
	}
	if len(bytes) != 0 {
		t.Error("Expected empty bytes for map of unregistered complex structs")
	}

	// Test with pointer to complex struct
	ptr := &complexStruct
	bytes, err = toBytes(ptr)
	if err == nil {
		t.Error("Expected error for pointer to unregistered complex struct")
	}
	if len(bytes) != 0 {
		t.Error("Expected empty bytes for pointer to unregistered complex struct")
	}
}

func TestCalculateSizePerformance(t *testing.T) {
	// Test performance with many calculations
	for i := 0; i < 1000; i++ {
		size := calculateSize(i)
		if size <= 0 {
			t.Errorf("Expected positive size for int %d, got %d", i, size)
		}
	}

	// Test performance with strings
	for i := 0; i < 100; i++ {
		str := string(make([]byte, i))
		size := calculateSize(str)
		if size != i {
			t.Errorf("Expected size %d for string of length %d, got %d", i, i, size)
		}
	}
}

func TestToBytesPerformance(t *testing.T) {
	// Test performance with many conversions
	for i := 0; i < 1000; i++ {
		bytes, err := toBytes(i)
		if err != nil {
			t.Errorf("Expected no error for int %d, got %v", i, err)
		}
		if len(bytes) == 0 {
			t.Errorf("Expected non-empty bytes for int %d", i)
		}
	}

	// Test performance with strings
	for i := 0; i < 100; i++ {
		str := string(make([]byte, i))
		bytes, err := toBytes(str)
		if err != nil {
			t.Errorf("Expected no error for string of length %d, got %v", i, err)
		}
		if len(bytes) != i {
			t.Errorf("Expected length %d for string of length %d, got %d", i, i, len(bytes))
		}
	}
}
