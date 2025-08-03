// entrypool_test.go: Unit tests for EntryPool in Metis strategic caching library
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package metis

import (
	"testing"
	"time"
)

func TestEntryPool_PutWithNilEntry(t *testing.T) {
	// Test that Put handles nil entries gracefully
	pool := NewEntryPool()

	// This should not panic
	pool.Put(nil)
}

func TestEntryPool_CreateEntryWithNegativeTTL(t *testing.T) {
	// Test CreateEntry with negative TTL
	pool := NewEntryPool()

	// Test with negative TTL (should set timestamp to past)
	entry := pool.CreateEntry("test_key", "test_value", -1*time.Hour, nil)

	if entry.Key != "test_key" {
		t.Errorf("Expected key 'test_key', got '%s'", entry.Key)
	}

	if entry.Data != "test_value" {
		t.Errorf("Expected data 'test_value', got '%v'", entry.Data)
	}

	// Timestamp should be in the past
	if !entry.Timestamp.Before(time.Now()) {
		t.Error("Expected timestamp to be in the past for negative TTL")
	}
}

func TestEntryPool_UpdateEntryWithNegativeTTL(t *testing.T) {
	// Test UpdateEntry with negative TTL
	pool := NewEntryPool()

	entry := pool.CreateEntry("test_key", "initial_value", 1*time.Hour, nil)

	// Update with negative TTL
	pool.UpdateEntry(entry, "updated_value", -1*time.Hour)

	if entry.Data != "updated_value" {
		t.Errorf("Expected data 'updated_value', got '%v'", entry.Data)
	}

	if entry.AccessCount != 0 {
		t.Errorf("Expected access count 0, got %d", entry.AccessCount)
	}

	// Timestamp should be in the past
	if !entry.Timestamp.Before(time.Now()) {
		t.Error("Expected timestamp to be in the past for negative TTL")
	}
}

func TestEntryPool_UpdateEntryWithZeroTTL(t *testing.T) {
	// Test UpdateEntry with zero TTL
	pool := NewEntryPool()

	entry := pool.CreateEntry("test_key", "initial_value", 1*time.Hour, nil)

	// Update with zero TTL
	pool.UpdateEntry(entry, "updated_value", 0)

	if entry.Data != "updated_value" {
		t.Errorf("Expected data 'updated_value', got '%v'", entry.Data)
	}

	// Timestamp should be zero
	if !entry.Timestamp.IsZero() {
		t.Error("Expected zero timestamp for zero TTL")
	}
}

func TestEntryPool_CreateEntryWithNilData(t *testing.T) {
	// Test CreateEntry with nil data
	pool := NewEntryPool()

	entry := pool.CreateEntry("test_key", nil, 1*time.Hour, nil)

	if entry.Key != "test_key" {
		t.Errorf("Expected key 'test_key', got '%s'", entry.Key)
	}

	if entry.Data != nil {
		t.Errorf("Expected nil data, got '%v'", entry.Data)
	}

	if !entry.IsNil {
		t.Error("Expected IsNil to be true for nil data")
	}
}

func TestEntryPool_UpdateEntryWithNilData(t *testing.T) {
	// Test UpdateEntry with nil data
	pool := NewEntryPool()

	entry := pool.CreateEntry("test_key", "initial_value", 1*time.Hour, nil)

	// Update with nil data
	pool.UpdateEntry(entry, nil, 1*time.Hour)

	if entry.Data != nil {
		t.Errorf("Expected nil data, got '%v'", entry.Data)
	}

	if !entry.IsNil {
		t.Error("Expected IsNil to be true for nil data")
	}
}

func TestEntryPool_IncrementAccess(t *testing.T) {
	// Test IncrementAccess
	pool := NewEntryPool()

	entry := pool.CreateEntry("test_key", "test_value", 1*time.Hour, nil)

	// Initial access count should be 0
	if entry.AccessCount != 0 {
		t.Errorf("Expected initial access count 0, got %d", entry.AccessCount)
	}

	// Increment access count
	pool.IncrementAccess(entry)
	if entry.AccessCount != 1 {
		t.Errorf("Expected access count 1, got %d", entry.AccessCount)
	}

	// Increment again
	pool.IncrementAccess(entry)
	if entry.AccessCount != 2 {
		t.Errorf("Expected access count 2, got %d", entry.AccessCount)
	}
}

func TestEntryPool_IsExpired(t *testing.T) {
	// Test IsExpired
	pool := NewEntryPool()

	// Test with zero timestamp (not expired)
	entry := pool.CreateEntry("test_key", "test_value", 0, nil)
	if pool.IsExpired(entry) {
		t.Error("Expected not expired for zero timestamp")
	}

	// Test with future timestamp (not expired)
	entry = pool.CreateEntry("test_key", "test_value", 1*time.Hour, nil)
	if pool.IsExpired(entry) {
		t.Error("Expected not expired for future timestamp")
	}

	// Test with past timestamp (expired)
	entry = pool.CreateEntry("test_key", "test_value", -1*time.Hour, nil)
	if !pool.IsExpired(entry) {
		t.Error("Expected expired for past timestamp")
	}
}

func TestEntryPool_ResetEntry(t *testing.T) {
	// Test ResetEntry
	pool := NewEntryPool()

	// Create an entry with some data
	entry := pool.CreateEntry("test_key", "test_value", 1*time.Hour, nil)
	entry.AccessCount = 5
	entry.IsNil = true

	// Reset the entry
	pool.ResetEntry(entry)

	// Check that all fields are reset
	if entry.Key != "" {
		t.Errorf("Expected empty key, got '%s'", entry.Key)
	}

	if entry.Data != nil {
		t.Errorf("Expected nil data, got '%v'", entry.Data)
	}

	if !entry.Timestamp.IsZero() {
		t.Error("Expected zero timestamp")
	}

	if entry.AccessCount != 0 {
		t.Errorf("Expected access count 0, got %d", entry.AccessCount)
	}

	if entry.llElem != nil {
		t.Error("Expected nil llElem")
	}

	if entry.IsNil {
		t.Error("Expected IsNil to be false")
	}
}

func TestEntryPool_Reuse(t *testing.T) {
	// Test that entries are properly reused from the pool
	pool := NewEntryPool()

	// Create and return an entry
	entry1 := pool.CreateEntry("key1", "value1", 1*time.Hour, nil)
	pool.Put(entry1)

	// Get a new entry (should be the same one, reset)
	entry2 := pool.Get()

	// Check that the entry is reset
	if entry2.Key != "" {
		t.Errorf("Expected empty key, got '%s'", entry2.Key)
	}

	if entry2.Data != nil {
		t.Errorf("Expected nil data, got '%v'", entry2.Data)
	}

	if !entry2.Timestamp.IsZero() {
		t.Error("Expected zero timestamp")
	}

	if entry2.AccessCount != 0 {
		t.Errorf("Expected access count 0, got %d", entry2.AccessCount)
	}

	if entry2.llElem != nil {
		t.Error("Expected nil llElem")
	}

	if entry2.IsNil {
		t.Error("Expected IsNil to be false")
	}
}
