// metis_edge_cases_test.go: Final edge cases to maximize coverage
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package metis

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// TestStrategicCache_Set_CachingDisabled tests Set when caching is disabled
func TestStrategicCache_Set_CachingDisabled(t *testing.T) {
	config := CacheConfig{
		EnableCaching: false, // Caching disabled
		CacheSize:     100,
	}
	cache := NewStrategicCache(config)
	defer cache.Close()

	result := cache.Set("test", "value")
	if result {
		t.Error("Expected Set to return false when caching is disabled")
	}

	// Verify value was not stored
	_, found := cache.Get("test")
	if found {
		t.Error("Expected value not to be stored when caching is disabled")
	}
}

// TestStrategicCache_Set_KeyTooLarge tests behavior when key exceeds MaxKeySize
func TestStrategicCache_Set_KeyTooLarge(t *testing.T) {
	config := CacheConfig{
		EnableCaching: true,
		CacheSize:     100,
		MaxKeySize:    5, // Very small key size limit
	}
	cache := NewStrategicCache(config)
	defer cache.Close()

	// Key longer than MaxKeySize
	longKey := "this_key_is_too_long"

	result := cache.Set(longKey, "value")
	if result {
		t.Error("Expected Set to return false for oversized key")
	}

	// Verify value was not stored
	_, found := cache.Get(longKey)
	if found {
		t.Error("Expected oversized key not to be stored")
	}
}

// TestStrategicCache_Set_ValueTooLarge tests behavior when value exceeds MaxValueSize
func TestStrategicCache_Set_ValueTooLarge(t *testing.T) {
	config := CacheConfig{
		EnableCaching: true,
		CacheSize:     100,
		MaxValueSize:  10, // Very small value size limit
	}
	cache := NewStrategicCache(config)
	defer cache.Close()

	// Value larger than MaxValueSize
	largeValue := strings.Repeat("A", 50)

	result := cache.Set("test", largeValue)
	if result {
		t.Error("Expected Set to return false for oversized value")
	}

	// Verify value was not stored
	_, found := cache.Get("test")
	if found {
		t.Error("Expected oversized value not to be stored")
	}
}

// TestStrategicCache_Set_ExactSizeLimits tests values at exact size limits
func TestStrategicCache_Set_ExactSizeLimits(t *testing.T) {
	config := CacheConfig{
		EnableCaching: true,
		CacheSize:     100,
		MaxKeySize:    10,
		MaxValueSize:  20,
	}
	cache := NewStrategicCache(config)
	defer cache.Close()

	// Key exactly at limit (should work)
	exactKey := strings.Repeat("K", 10)
	result := cache.Set(exactKey, "value")
	if !result {
		t.Error("Expected Set to succeed for key at exact size limit")
	}

	// Value exactly at limit (should work)
	exactValue := strings.Repeat("V", 20)
	result = cache.Set("test2", exactValue)
	if !result {
		t.Error("Expected Set to succeed for value at exact size limit")
	}

	// Key one byte over limit (should fail)
	oversizedKey := strings.Repeat("K", 11)
	result = cache.Set(oversizedKey, "value")
	if result {
		t.Error("Expected Set to fail for key over size limit")
	}

	// Value one byte over limit (should fail)
	oversizedValue := strings.Repeat("V", 21)
	result = cache.Set("test3", oversizedValue)
	if result {
		t.Error("Expected Set to fail for value over size limit")
	}
}

// TestStrategicCache_Set_TTLExpiredEntry tests updating an already expired entry
func TestStrategicCache_Set_TTLExpiredEntry(t *testing.T) {
	config := CacheConfig{
		EnableCaching: true,
		CacheSize:     100,
		TTL:           time.Millisecond * 10, // Very short TTL
	}
	cache := NewStrategicCache(config)
	defer cache.Close()

	// Set initial value
	cache.Set("expiring", "initial")

	// Wait for expiration
	time.Sleep(time.Millisecond * 15)

	// Set new value for expired key
	result := cache.Set("expiring", "updated")
	if !result {
		t.Error("Expected Set to succeed for expired entry")
	}

	// Verify new value is stored
	value, found := cache.Get("expiring")
	if !found || value != "updated" {
		t.Error("Expected updated value to be stored and retrievable")
	}
}

// TestCompressGzipWithHeader_DataSizeBoundaries tests compression at various data sizes
func TestCompressGzipWithHeader_DataSizeBoundaries(t *testing.T) {
	testCases := []struct {
		name string
		size int
	}{
		{"Size_0", 0},
		{"Size_1", 1},
		{"Size_63", 63}, // Just under threshold
		{"Size_64", 64}, // At threshold
		{"Size_65", 65}, // Just over threshold
		{"Size_128", 128},
		{"Size_1024", 1024},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data := make([]byte, tc.size)
			for i := range data {
				data[i] = byte(i % 256)
			}

			compressed, err := compressGzipWithHeader(data, "TEST")
			if err != nil {
				t.Errorf("Expected no error for size %d, got: %v", tc.size, err)
			}

			if len(compressed) < 4 {
				t.Errorf("Expected compressed data to include 4-byte header, got %d bytes", len(compressed))
			}

			// Verify header
			header := string(compressed[:4])
			if header != "TEST" {
				t.Errorf("Expected header 'TEST', got '%s'", header)
			}

			// Verify decompression works
			decompressedHeader, payload, err := decompressGzipWithHeader(compressed)
			if err != nil {
				t.Errorf("Failed to decompress: %v", err)
			}
			if decompressedHeader != "TEST" {
				t.Errorf("Expected decompressed header 'TEST', got '%s'", decompressedHeader)
			}
			if !bytes.Equal(payload, data) {
				t.Error("Decompressed data doesn't match original")
			}
		})
	}
}

// TestCompressGzipWithHeader_HighlyCompressibleData tests compression efficiency
func TestCompressGzipWithHeader_HighlyCompressibleData(t *testing.T) {
	// Create highly compressible data (repeated pattern)
	data := bytes.Repeat([]byte("ABCD"), 1000) // 4000 bytes of "ABCDABCD..."

	compressed, err := compressGzipWithHeader(data, "COMP")
	if err != nil {
		t.Errorf("Expected no error for compressible data, got: %v", err)
	}

	// Should compress significantly
	if len(compressed) >= len(data) {
		t.Error("Expected data to compress to smaller size")
	}

	// Verify it can be decompressed correctly
	header, payload, err := decompressGzipWithHeader(compressed)
	if err != nil {
		t.Errorf("Failed to decompress highly compressible data: %v", err)
	}
	if header != "COMP" {
		t.Errorf("Expected header 'COMP', got '%s'", header)
	}
	if !bytes.Equal(payload, data) {
		t.Error("Decompressed data doesn't match original highly compressible data")
	}
}

// TestDecompressGzipWithHeader_CorruptedGzipStream tests handling of corrupted gzip data
func TestDecompressGzipWithHeader_CorruptedGzipStream(t *testing.T) {
	// Create valid compressed data first
	testData := []byte("Hello, compression!")
	compressed, err := compressGzipWithHeader(testData, "TEST")
	if err != nil {
		t.Fatalf("Failed to create test data: %v", err)
	}

	// Corrupt the gzip stream (keep header intact, corrupt payload)
	corrupted := make([]byte, len(compressed))
	copy(corrupted, compressed)

	// Corrupt multiple bytes in the gzip stream to ensure error
	if len(corrupted) > 15 {
		for i := 8; i < 15 && i < len(corrupted); i++ {
			corrupted[i] = ^corrupted[i] // Flip all bits
		}
	}

	// Should handle corruption gracefully
	header, payload, err := decompressGzipWithHeader(corrupted)

	// Even with error, header should still be extracted
	if header != "TEST" {
		t.Errorf("Expected header 'TEST' even with corrupted stream, got '%s'", header)
	}

	// Either it should error, or return valid uncompressed data
	if err != nil {
		// Error case - payload should be nil or empty
		if len(payload) > 0 {
			t.Error("Expected nil or empty payload when decompression fails")
		}
	} else {
		// Success case - it might have fallen back to treating as uncompressed
		t.Logf("Corruption handled as uncompressed data, payload length: %d", len(payload))
	}
}

// TestDecompressGzipWithHeader_TruncatedData tests various truncated data scenarios
func TestDecompressGzipWithHeader_TruncatedData(t *testing.T) {
	testCases := []struct {
		name      string
		dataLen   int
		expectErr bool
	}{
		{"Empty", 0, true},
		{"Size_1", 1, true},
		{"Size_2", 2, true},
		{"Size_3", 3, true},
		{"Size_4", 4, false}, // Just header, no payload
		{"Size_5", 5, false}, // Header + 1 byte
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create truncated data
			data := []byte("HEAD")
			if tc.dataLen > 4 {
				data = append(data, make([]byte, tc.dataLen-4)...)
			} else if tc.dataLen < 4 {
				data = data[:tc.dataLen]
			}

			header, payload, err := decompressGzipWithHeader(data)

			if tc.expectErr && err == nil {
				t.Errorf("Expected error for data length %d", tc.dataLen)
			}
			if !tc.expectErr && err != nil {
				t.Errorf("Expected no error for data length %d, got: %v", tc.dataLen, err)
			}

			// For data >= 4 bytes, header should be extractable
			if tc.dataLen >= 4 && header != "HEAD" {
				t.Errorf("Expected header 'HEAD', got '%s'", header)
			}

			// For valid cases with no gzip data, payload should be the remainder
			if !tc.expectErr && tc.dataLen > 4 {
				expectedPayload := data[4:]
				if !bytes.Equal(payload, expectedPayload) {
					t.Error("Payload doesn't match expected remainder")
				}
			}
		})
	}
}

// TestCompressGzipWithHeader_ZeroSizeThreshold tests the exact boundary at 64 bytes
func TestCompressGzipWithHeader_ZeroSizeThreshold(t *testing.T) {
	// Test data exactly under the 64-byte threshold
	data63 := make([]byte, 63)
	for i := range data63 {
		data63[i] = byte(i % 256)
	}

	compressed, err := compressGzipWithHeader(data63, "UND6")
	if err != nil {
		t.Errorf("Expected no error for 63-byte data, got: %v", err)
	}

	// Should not be compressed, just header + data
	expectedLen := 4 + 63 // header + data
	if len(compressed) != expectedLen {
		t.Errorf("Expected uncompressed data length %d, got %d", expectedLen, len(compressed))
	}

	// Verify it contains header + original data
	if string(compressed[:4]) != "UND6" {
		t.Error("Header not correct for uncompressed data")
	}
	if !bytes.Equal(compressed[4:], data63) {
		t.Error("Data portion not correct for uncompressed data")
	}
}

// TestDecompressGzipWithHeader_ExactGzipMagicNumber tests exact gzip magic number handling
func TestDecompressGzipWithHeader_ExactGzipMagicNumber(t *testing.T) {
	// Test data with exactly the gzip magic number but invalid
	testData := []byte("HEAD")
	testData = append(testData, 0x1f, 0x8b) // Exactly gzip magic but truncated

	header, payload, err := decompressGzipWithHeader(testData)
	if err == nil {
		t.Error("Expected error for truncated gzip magic number")
	}
	if header != "HEAD" {
		t.Errorf("Expected header 'HEAD', got '%s'", header)
	}
	if payload != nil {
		t.Error("Expected nil payload for invalid gzip data")
	}
}
