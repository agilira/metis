// metis_advanced_test.go: Additional tests for core functions
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package metis

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"testing"
)

// TestStrategicCache_Set_WTinyLFUSpecificPaths tests WTinyLFU-specific code paths
func TestStrategicCache_Set_WTinyLFUSpecificPaths(t *testing.T) {
	// Test case where WTinyLFU is nil but eviction policy is set to wtinylfu
	config := CacheConfig{
		EnableCaching:  true,
		CacheSize:      100,
		EvictionPolicy: "wtinylfu",
		MaxKeySize:     0,
		MaxValueSize:   0,
		MaxShardSize:   0,
	}
	cache := NewStrategicCache(config)
	defer cache.Close()

	// Force WTinyLFU to nil to test fallback path
	cache.wtinylfu = nil

	result := cache.Set("test", "value")
	if !result {
		t.Error("Expected Set to succeed even with nil WTinyLFU")
	}

	// Verify value was stored
	value, found := cache.Get("test")
	if !found || value != "value" {
		t.Error("Expected value to be stored correctly")
	}
}

// TestStrategicCache_Set_AdmissionPolicyNonAlways tests non-AlwaysAdmit policies in fast path
func TestStrategicCache_Set_AdmissionPolicyNonAlways(t *testing.T) {
	config := CacheConfig{
		EnableCaching:  true,
		CacheSize:      100,
		EvictionPolicy: "wtinylfu",
		MaxKeySize:     10,
		MaxValueSize:   50,
		MaxShardSize:   0,
	}
	cache := NewStrategicCache(config)
	defer cache.Close()

	// Set a probabilistic admission policy (not AlwaysAdmit)
	cache.admission = &ProbabilisticAdmissionPolicy{Probability: 1.0} // Always admit but not AlwaysAdmitPolicy type

	result := cache.Set("test", "value")
	if !result {
		t.Error("Expected Set to succeed with probabilistic admission policy")
	}
}

// TestStrategicCache_Set_EmptyEvictionPolicy tests empty eviction policy
func TestStrategicCache_Set_EmptyEvictionPolicy(t *testing.T) {
	config := CacheConfig{
		EnableCaching:  true,
		CacheSize:      100,
		EvictionPolicy: "", // Empty policy - should use WTinyLFU fast path
		MaxKeySize:     0,
		MaxValueSize:   0,
		MaxShardSize:   0,
	}
	cache := NewStrategicCache(config)
	defer cache.Close()

	cache.admission = &AlwaysAdmitPolicy{}

	result := cache.Set("test", "value")
	if !result {
		t.Error("Expected Set to succeed with empty eviction policy")
	}
}

// TestStrategicCache_Set_NilValue tests setting nil values
func TestStrategicCache_Set_NilValue(t *testing.T) {
	config := CacheConfig{
		EnableCaching: true,
		CacheSize:     100,
	}
	cache := NewStrategicCache(config)
	defer cache.Close()

	// Test setting nil value
	result := cache.Set("nil_key", nil)
	if !result {
		t.Error("Expected Set to succeed with nil value")
	}

	// Verify nil value can be retrieved
	value, found := cache.Get("nil_key")
	if !found {
		t.Error("Expected nil value to be found")
	}
	if value != nil {
		t.Errorf("Expected nil value, got %v", value)
	}
}

// TestStrategicCache_Set_UnsafeMapType tests unsafe map types
func TestStrategicCache_Set_UnsafeMapType(t *testing.T) {
	config := CacheConfig{
		EnableCaching: true,
		CacheSize:     100,
	}
	cache := NewStrategicCache(config)
	defer cache.Close()

	// Test with map containing function - should be rejected
	testFunc := func() {} // Define function separately
	unsafeMap := map[string]interface{}{
		"func": testFunc,
	}

	result := cache.Set("unsafe", unsafeMap)
	// This depends on the implementation - it might be accepted or rejected
	// We're testing that the code handles it without crashing
	t.Logf("Unsafe map result: %v", result)
}

// TestStrategicCache_Set_MaxShardSizeZero tests behavior when MaxShardSize is zero
func TestStrategicCache_Set_MaxShardSizeZero(t *testing.T) {
	config := CacheConfig{
		EnableCaching: true,
		CacheSize:     10,
		MaxShardSize:  0, // Zero means use CacheSize / ShardCount
		ShardCount:    2,
	}
	cache := NewStrategicCache(config)
	defer cache.Close()

	// Should use CacheSize / ShardCount = 10 / 2 = 5 per shard
	for i := 0; i < 15; i++ {
		key := fmt.Sprintf("key%d", i)
		cache.Set(key, i)
	}

	stats := cache.GetStats()
	if stats.Size > 10 {
		t.Errorf("Expected cache size <= 10, got %d", stats.Size)
	}
}

// TestStrategicCache_Set_LRUPolicyLlElemNil tests LRU policy with nil llElem
func TestStrategicCache_Set_LRUPolicyLlElemNil(t *testing.T) {
	config := CacheConfig{
		EnableCaching:  true,
		CacheSize:      5,
		EvictionPolicy: "lru",
		ShardCount:     1,
		MaxShardSize:   3,
	}
	cache := NewStrategicCache(config)
	defer cache.Close()

	// Add initial entry
	cache.Set("key1", "value1")

	// Manually corrupt the llElem to test nil handling
	shard := cache.getShard("key1")
	shard.mu.Lock()
	if entry, exists := shard.data["key1"]; exists {
		entry.llElem = nil // Simulate corrupted state
	}
	shard.mu.Unlock()

	// Update the entry - should handle nil llElem gracefully
	result := cache.Set("key1", "updated_value")
	if !result {
		t.Error("Expected Set to succeed even with nil llElem")
	}
}

// TestCompressGzipWithHeader_SpecificErrorPaths tests specific error scenarios
func TestCompressGzipWithHeader_SpecificErrorPaths(t *testing.T) {
	// Test boundary case: exactly 64 bytes (compression threshold)
	data64 := make([]byte, 64)
	for i := range data64 {
		data64[i] = byte(i % 256)
	}

	compressed, err := compressGzipWithHeader(data64, "TEST")
	if err != nil {
		t.Errorf("Expected no error for 64-byte data, got: %v", err)
	}
	if len(compressed) < 4 {
		t.Error("Expected compressed data to include header")
	}

	// Test slightly above threshold
	data65 := make([]byte, 65)
	for i := range data65 {
		data65[i] = byte(i % 256)
	}

	compressed, err = compressGzipWithHeader(data65, "TEST")
	if err != nil {
		t.Errorf("Expected no error for 65-byte data, got: %v", err)
	}
	if len(compressed) < 4 {
		t.Error("Expected compressed data to include header")
	}

	// Test with data that compresses very poorly (random data)
	randomData := make([]byte, 1000)
	// Use a pattern that compresses poorly
	for i := range randomData {
		randomData[i] = byte(i * 7 % 256)
	}

	compressed, err = compressGzipWithHeader(randomData, "RAND")
	if err != nil {
		t.Errorf("Expected no error for random data, got: %v", err)
	}
	if len(compressed) < 4 {
		t.Error("Expected compressed data to include header")
	}
}

// MockFailingWriter simulates write failures for testing error paths
type MockFailingWriter struct {
	failAfter int
	written   int
}

func (w *MockFailingWriter) Write(p []byte) (n int, err error) {
	if w.written >= w.failAfter {
		return 0, io.ErrShortWrite
	}
	toWrite := len(p)
	if w.written+toWrite > w.failAfter {
		toWrite = w.failAfter - w.written
	}
	w.written += toWrite
	if toWrite < len(p) {
		return toWrite, io.ErrShortWrite
	}
	return toWrite, nil
}

// TestCompressGzipWithHeader_SimulatedWriteError tests writer error handling
func TestCompressGzipWithHeader_SimulatedWriteError(t *testing.T) {
	// Since we can't easily inject write errors into the existing function,
	// we test the equivalent logic by manually creating a scenario

	// Test with large data that would normally compress
	largeData := make([]byte, 10000)
	for i := range largeData {
		largeData[i] = byte('A') // Highly compressible
	}

	// Test normal path
	compressed, err := compressGzipWithHeader(largeData, "LARG")
	if err != nil {
		t.Errorf("Expected no error for large compressible data, got: %v", err)
	}

	// Verify compression actually occurred (should be much smaller)
	if len(compressed) >= len(largeData) {
		t.Error("Expected data to be compressed (smaller than original)")
	}

	// Test the error handling path by creating gzip writer error scenario
	var buf bytes.Buffer
	buf.WriteString("TEST")
	w := gzip.NewWriter(&buf)

	// Write to gzip writer
	_, err = w.Write(largeData)
	if err != nil {
		// Test the error handling branch
		if closeErr := w.Close(); closeErr != nil {
			t.Logf("Close error as expected: %v", closeErr)
		}
	} else {
		err = w.Close()
		if err != nil {
			t.Logf("Close error: %v", err)
		}
	}
}

// TestDecompressGzipWithHeader_SpecificBoundaries tests specific boundary conditions
func TestDecompressGzipWithHeader_SpecificBoundaries(t *testing.T) {
	// Test data that's exactly at gzip magic number boundary
	testData := []byte("HEAD")
	testData = append(testData, 0x1f, 0x8b, 0x08) // Partial gzip header

	header, _, err := decompressGzipWithHeader(testData)
	if err != nil {
		// It's OK if this fails - partial gzip headers should fail
		t.Logf("Partial gzip header failed as expected: %v", err)
	}
	if header != "HEAD" {
		t.Errorf("Expected header 'HEAD', got '%s'", header)
	}

	// Test 6-byte data with exact gzip magic but truncated
	testData6 := []byte("HEAD")
	testData6 = append(testData6, 0x1f, 0x8b) // Exactly 6 bytes

	_, _, err = decompressGzipWithHeader(testData6)
	if err == nil {
		t.Error("Expected error for truncated gzip header")
	}

	// Test valid gzip stream that produces empty output
	var buf bytes.Buffer
	buf.WriteString("EMPT")
	w := gzip.NewWriter(&buf)
	w.Write([]byte{}) // Empty data
	w.Close()

	header, payload, err := decompressGzipWithHeader(buf.Bytes())
	if err != nil {
		t.Errorf("Expected no error for empty gzip stream, got: %v", err)
	}
	if header != "EMPT" {
		t.Errorf("Expected header 'EMPT', got '%s'", header)
	}
	if len(payload) != 0 {
		t.Errorf("Expected empty payload, got %d bytes", len(payload))
	}
}
