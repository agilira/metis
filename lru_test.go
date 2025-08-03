// lru_test.go: Unit tests for LRU implementation in Metis strategic caching library
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package metis

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNewLRU(t *testing.T) {
	// Test with positive capacity
	lru := NewLRU(10)
	if lru == nil {
		t.Fatal("NewLRU should not return nil")
	}
	if lru.capacity != 10 {
		t.Errorf("Expected capacity 10, got %d", lru.capacity)
	}
	if lru.list == nil {
		t.Error("List should be initialized")
	}
	if lru.cache == nil {
		t.Error("Cache map should be initialized")
	}
	if lru.hits != 0 {
		t.Errorf("Expected initial hits 0, got %d", lru.hits)
	}

	// Test with zero capacity
	lru = NewLRU(0)
	if lru.capacity != 0 {
		t.Errorf("Expected capacity 0, got %d", lru.capacity)
	}

	// Test with negative capacity
	lru = NewLRU(-5)
	if lru.capacity != -5 {
		t.Errorf("Expected capacity -5, got %d", lru.capacity)
	}
}

func TestLRU_Get(t *testing.T) {
	lru := NewLRU(3)

	// Test Get with non-existent key
	value, exists := lru.Get("nonexistent")
	if exists {
		t.Error("Get should return false for non-existent key")
	}
	if value != nil {
		t.Errorf("Expected nil value for non-existent key, got %v", value)
	}

	// Test Get with existing key
	lru.Set("key1", "value1")
	value, exists = lru.Get("key1")
	if !exists {
		t.Error("Get should return true for existing key")
	}
	if value != "value1" {
		t.Errorf("Expected value 'value1', got %v", value)
	}

	// Test that Get moves item to front
	lru.Set("key2", "value2")
	lru.Set("key3", "value3")

	// Access key1 to move it to front
	lru.Get("key1")

	// Add key4 to trigger eviction
	lru.Set("key4", "value4")

	// key2 should be evicted (least recently used)
	_, exists = lru.Get("key2")
	if exists {
		t.Error("key2 should have been evicted")
	}

	// key1 should still exist (was accessed)
	_, exists = lru.Get("key1")
	if !exists {
		t.Error("key1 should still exist after access")
	}
}

func TestLRU_Set(t *testing.T) {
	lru := NewLRU(2)

	// Test Set with new key
	success := lru.Set("key1", "value1")
	if !success {
		t.Error("Set should succeed for new key")
	}
	if lru.Size() != 1 {
		t.Errorf("Expected size 1, got %d", lru.Size())
	}

	// Test Set with existing key (update)
	success = lru.Set("key1", "updated_value")
	if !success {
		t.Error("Set should succeed for existing key")
	}
	value, exists := lru.Get("key1")
	if !exists {
		t.Error("Updated key should exist")
	}
	if value != "updated_value" {
		t.Errorf("Expected 'updated_value', got %v", value)
	}

	// Test Set with nil value
	success = lru.Set("nil_key", nil)
	if !success {
		t.Error("Set should succeed with nil value")
	}
	value, exists = lru.Get("nil_key")
	if !exists {
		t.Error("Nil value should be retrievable")
	}
	if value != nil {
		t.Errorf("Expected nil value, got %v", value)
	}

	// Test Set with empty string
	success = lru.Set("empty_key", "")
	if !success {
		t.Error("Set should succeed with empty string")
	}
	value, exists = lru.Get("empty_key")
	if !exists {
		t.Error("Empty string should be retrievable")
	}
	if value != "" {
		t.Errorf("Expected empty string, got %v", value)
	}
}

func TestLRU_Eviction(t *testing.T) {
	lru := NewLRU(2)

	// Fill cache
	lru.Set("key1", "value1")
	lru.Set("key2", "value2")

	// Add third item, should evict key1 (least recently used)
	lru.Set("key3", "value3")

	// key1 should be evicted
	_, exists := lru.Get("key1")
	if exists {
		t.Error("key1 should have been evicted")
	}

	// key2 and key3 should exist
	_, exists = lru.Get("key2")
	if !exists {
		t.Error("key2 should exist")
	}
	_, exists = lru.Get("key3")
	if !exists {
		t.Error("key3 should exist")
	}

	// Access key2 to make key3 least recently used
	lru.Get("key2")

	// Add key4, should evict key3
	lru.Set("key4", "value4")

	// key3 should be evicted
	_, exists = lru.Get("key3")
	if exists {
		t.Error("key3 should have been evicted")
	}

	// key2 and key4 should exist
	_, exists = lru.Get("key2")
	if !exists {
		t.Error("key2 should exist")
	}
	_, exists = lru.Get("key4")
	if !exists {
		t.Error("key4 should exist")
	}
}

func TestLRU_Exists(t *testing.T) {
	lru := NewLRU(3)

	// Test Exists with non-existent key
	if lru.Exists("nonexistent") {
		t.Error("Exists should return false for non-existent key")
	}

	// Test Exists with existing key
	lru.Set("key1", "value1")
	if !lru.Exists("key1") {
		t.Error("Exists should return true for existing key")
	}

	// Test Exists with empty key
	if lru.Exists("") {
		t.Error("Exists should return false for empty key")
	}

	// Test Exists after deletion
	lru.Delete("key1")
	if lru.Exists("key1") {
		t.Error("Exists should return false after deletion")
	}
}

func TestLRU_Delete(t *testing.T) {
	lru := NewLRU(3)

	// Test Delete with non-existent key
	lru.Delete("nonexistent") // Should not panic

	// Test Delete with existing key
	lru.Set("key1", "value1")
	lru.Set("key2", "value2")

	lru.Delete("key1")

	// key1 should not exist
	_, exists := lru.Get("key1")
	if exists {
		t.Error("key1 should not exist after deletion")
	}

	// key2 should still exist
	_, exists = lru.Get("key2")
	if !exists {
		t.Error("key2 should still exist after deleting key1")
	}

	// Test Delete with empty key
	lru.Set("", "empty_value")
	lru.Delete("")
	_, exists = lru.Get("")
	if exists {
		t.Error("Empty key should not exist after deletion")
	}

	// Test Delete reduces size
	initialSize := lru.Size()
	lru.Set("key3", "value3")
	if lru.Size() != initialSize+1 {
		t.Errorf("Expected size %d, got %d", initialSize+1, lru.Size())
	}

	lru.Delete("key3")
	if lru.Size() != initialSize {
		t.Errorf("Expected size %d after deletion, got %d", initialSize, lru.Size())
	}
}

func TestLRU_Clear(t *testing.T) {
	lru := NewLRU(5)

	// Add some items
	lru.Set("key1", "value1")
	lru.Set("key2", "value2")
	lru.Set("key3", "value3")

	if lru.Size() != 3 {
		t.Errorf("Expected size 3, got %d", lru.Size())
	}

	// Clear cache
	lru.Clear()

	// Size should be 0
	if lru.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", lru.Size())
	}

	// Items should not exist
	_, exists := lru.Get("key1")
	if exists {
		t.Error("key1 should not exist after clear")
	}
	_, exists = lru.Get("key2")
	if exists {
		t.Error("key2 should not exist after clear")
	}
	_, exists = lru.Get("key3")
	if exists {
		t.Error("key3 should not exist after clear")
	}

	// Test Clear on empty cache
	lru.Clear() // Should not panic
	if lru.Size() != 0 {
		t.Error("Size should remain 0 after clearing empty cache")
	}
}

func TestLRU_Size(t *testing.T) {
	lru := NewLRU(5)

	// Test initial size
	if lru.Size() != 0 {
		t.Errorf("Expected initial size 0, got %d", lru.Size())
	}

	// Test size after adding items
	lru.Set("key1", "value1")
	if lru.Size() != 1 {
		t.Errorf("Expected size 1, got %d", lru.Size())
	}

	lru.Set("key2", "value2")
	if lru.Size() != 2 {
		t.Errorf("Expected size 2, got %d", lru.Size())
	}

	// Test size after updating existing key
	lru.Set("key1", "updated_value")
	if lru.Size() != 2 {
		t.Errorf("Expected size 2 after update, got %d", lru.Size())
	}

	// Test size after deletion
	lru.Delete("key1")
	if lru.Size() != 1 {
		t.Errorf("Expected size 1 after deletion, got %d", lru.Size())
	}

	// Test size after clear
	lru.Clear()
	if lru.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", lru.Size())
	}
}

func TestLRU_MaxSize(t *testing.T) {
	// Test with positive capacity
	lru := NewLRU(10)
	if lru.MaxSize() != 10 {
		t.Errorf("Expected MaxSize 10, got %d", lru.MaxSize())
	}

	// Test with zero capacity
	lru = NewLRU(0)
	if lru.MaxSize() != 0 {
		t.Errorf("Expected MaxSize 0, got %d", lru.MaxSize())
	}

	// Test with negative capacity
	lru = NewLRU(-5)
	if lru.MaxSize() != -5 {
		t.Errorf("Expected MaxSize -5, got %d", lru.MaxSize())
	}
}

func TestLRU_Hits(t *testing.T) {
	lru := NewLRU(3)

	// Test initial hits
	if lru.Hits() != 0 {
		t.Errorf("Expected initial hits 0, got %d", lru.Hits())
	}

	// Test hits after successful Get
	lru.Set("key1", "value1")
	lru.Get("key1")
	if lru.Hits() != 1 {
		t.Errorf("Expected hits 1, got %d", lru.Hits())
	}

	// Test hits after multiple successful Gets
	lru.Get("key1")
	lru.Get("key1")
	if lru.Hits() != 3 {
		t.Errorf("Expected hits 3, got %d", lru.Hits())
	}

	// Test hits after failed Get
	lru.Get("nonexistent")
	if lru.Hits() != 3 {
		t.Errorf("Expected hits 3 after failed Get, got %d", lru.Hits())
	}

	// Test hits after Set (should not increment)
	lru.Set("key2", "value2")
	if lru.Hits() != 3 {
		t.Errorf("Expected hits 3 after Set, got %d", lru.Hits())
	}
}

func TestLRU_Evict(t *testing.T) {
	lru := NewLRU(3)

	// Test Evict on empty cache
	key, value := lru.Evict()
	if key != "" {
		t.Errorf("Expected empty key for empty cache, got %s", key)
	}
	if value != nil {
		t.Errorf("Expected nil value for empty cache, got %v", value)
	}

	// Test Evict with items
	lru.Set("key1", "value1")
	lru.Set("key2", "value2")
	lru.Set("key3", "value3")

	// Evict should remove least recently used (key1)
	key, value = lru.Evict()
	if key != "key1" {
		t.Errorf("Expected evicted key 'key1', got %s", key)
	}
	if value != "value1" {
		t.Errorf("Expected evicted value 'value1', got %v", value)
	}

	// Size should be reduced
	if lru.Size() != 2 {
		t.Errorf("Expected size 2 after eviction, got %d", lru.Size())
	}

	// key1 should not exist
	_, exists := lru.Get("key1")
	if exists {
		t.Error("key1 should not exist after eviction")
	}

	// key2 and key3 should still exist
	_, exists = lru.Get("key2")
	if !exists {
		t.Error("key2 should exist after eviction")
	}
	_, exists = lru.Get("key3")
	if !exists {
		t.Error("key3 should exist after eviction")
	}
}

func TestLRU_evict(t *testing.T) {
	lru := NewLRU(3)

	// Test evict on empty cache
	lru.evict() // Should not panic
	if lru.Size() != 0 {
		t.Error("Size should remain 0 after evict on empty cache")
	}

	// Test evict with items
	lru.Set("key1", "value1")
	lru.Set("key2", "value2")
	lru.Set("key3", "value3")

	// Add fourth item to trigger eviction
	lru.Set("key4", "value4")

	// key1 should be evicted (least recently used)
	_, exists := lru.Get("key1")
	if exists {
		t.Error("key1 should have been evicted")
	}

	// Size should be 3
	if lru.Size() != 3 {
		t.Errorf("Expected size 3 after eviction, got %d", lru.Size())
	}
}

func TestLRU_ConcurrentAccess(t *testing.T) {
	lru := NewLRU(100)
	var wg sync.WaitGroup
	numGoroutines := 10
	operationsPerGoroutine := 100

	// Start multiple goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				value := fmt.Sprintf("value_%d_%d", id, j)

				// Set value
				lru.Set(key, value)

				// Get value
				if got, exists := lru.Get(key); exists {
					if got != value {
						t.Errorf("Expected %s, got %v", value, got)
					}
				}

				// Delete value
				lru.Delete(key)
			}
		}(i)
	}

	wg.Wait()

	// Test that cache is still functional after concurrent access
	lru.Set("final_key", "final_value")
	value, exists := lru.Get("final_key")
	if !exists {
		t.Error("Cache should still work after concurrent access")
	}
	if value != "final_value" {
		t.Errorf("Expected 'final_value', got %v", value)
	}
}

func TestLRU_EdgeCases(t *testing.T) {
	// Test with capacity 0
	lru := NewLRU(0)
	lru.Set("key1", "value1")
	if lru.Size() != 0 {
		t.Error("Cache with capacity 0 should not store items")
	}

	// Test with capacity 1
	lru = NewLRU(1)
	lru.Set("key1", "value1")
	lru.Set("key2", "value2")

	// key1 should be evicted
	_, exists := lru.Get("key1")
	if exists {
		t.Error("key1 should be evicted in capacity 1 cache")
	}

	// key2 should exist
	_, exists = lru.Get("key2")
	if !exists {
		t.Error("key2 should exist in capacity 1 cache")
	}

	// Test with very large capacity
	lru = NewLRU(10000)
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key_%d", i)
		value := fmt.Sprintf("value_%d", i)
		lru.Set(key, value)
	}

	if lru.Size() != 1000 {
		t.Errorf("Expected size 1000, got %d", lru.Size())
	}
}

func TestLRU_ComplexValues(t *testing.T) {
	lru := NewLRU(5)

	// Test with different value types
	testCases := []struct {
		key   string
		value interface{}
	}{
		{"string", "hello world"},
		{"int", 42},
		{"float", 3.14},
		{"bool", true},
		{"slice", []int{1, 2, 3}},
		{"map", map[string]int{"a": 1, "b": 2}},
		{"nil", nil},
		{"empty_string", ""},
	}

	for _, tc := range testCases {
		success := lru.Set(tc.key, tc.value)
		if !success {
			t.Errorf("Set should succeed for %s", tc.key)
		}

		value, exists := lru.Get(tc.key)
		if !exists {
			t.Errorf("Get should find %s", tc.key)
		}

		// For complex types, we can't easily compare, so just check it's not nil (except for nil case)
		if tc.value == nil {
			if value != nil {
				t.Errorf("Expected nil for %s, got %v", tc.key, value)
			}
		} else if value == nil {
			t.Errorf("Expected non-nil for %s, got nil", tc.key)
		}
	}
}

func TestLRU_Performance(t *testing.T) {
	lru := NewLRU(1000)

	// Benchmark Set operations
	start := time.Now()
	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("key_%d", i)
		value := fmt.Sprintf("value_%d", i)
		lru.Set(key, value)
	}
	setDuration := time.Since(start)

	// Benchmark Get operations
	start = time.Now()
	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("key_%d", i)
		lru.Get(key)
	}
	getDuration := time.Since(start)

	// Performance should be reasonable (less than 1 second for 10k operations)
	if setDuration > time.Second {
		t.Errorf("Set operations took too long: %v", setDuration)
	}
	if getDuration > time.Second {
		t.Errorf("Get operations took too long: %v", getDuration)
	}
}

func TestLRU_StressTest(t *testing.T) {
	lru := NewLRU(100)

	// Stress test with many operations
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key_%d", i)
		value := fmt.Sprintf("value_%d", i)

		// Set
		lru.Set(key, value)

		// Get
		if got, exists := lru.Get(key); exists {
			if got != value {
				t.Errorf("Expected %s, got %v", value, got)
			}
		}

		// Update
		newValue := fmt.Sprintf("updated_value_%d", i)
		lru.Set(key, newValue)

		// Get updated value
		if got, exists := lru.Get(key); exists {
			if got != newValue {
				t.Errorf("Expected %s, got %v", newValue, got)
			}
		}

		// Delete every 10th item
		if i%10 == 0 {
			lru.Delete(key)
			if lru.Exists(key) {
				t.Errorf("key %s should not exist after deletion", key)
			}
		}
	}

	// Final size should be reasonable
	if lru.Size() > 100 {
		t.Errorf("Final size %d should not exceed capacity 100", lru.Size())
	}
}
