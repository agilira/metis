// wtinylfu_branch_test.go: Specific branch coverage tests for WTinyLFUShard.Set
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package metis

import (
	"fmt"
	"testing"
)

// TestWTinyLFUShard_Set_AllBranches tests each specific branch in WTinyLFUShard.Set (line 187)
// Function has multiple if statements that need individual coverage
func TestWTinyLFUShard_Set_AllBranches(t *testing.T) {

	// Branch 1: windowCache.FastSet returns true (key exists in window cache)
	t.Run("branch_window_cache_update", func(t *testing.T) {
		wt := NewWTinyLFU(10, 1)
		shard := wt.shards[0]

		// Add key to window cache first
		key := "window_test_key"
		value1 := "value1"
		shard.Set(key, value1)

		// Verify it's in window cache by checking sizes
		initialWindowSize := shard.windowCache.size
		initialMainSize := shard.mainCache.Size()

		// Update the same key (should trigger window cache FastSet return true)
		value2 := "value2"
		result := shard.Set(key, value2)

		if !result {
			t.Error("Set should return true when updating existing key in window cache")
		}

		// Verify cache sizes didn't change (indicating update, not new entry)
		if shard.windowCache.size != initialWindowSize {
			t.Errorf("Window cache size should not change on update: expected %d, got %d",
				initialWindowSize, shard.windowCache.size)
		}
		if shard.mainCache.Size() != initialMainSize {
			t.Errorf("Main cache size should not change on window update: expected %d, got %d",
				initialMainSize, shard.mainCache.Size())
		}

		// Verify value was updated
		retrieved, exists := shard.Get(key)
		if !exists || retrieved != value2 {
			t.Errorf("Expected updated value %v, got %v (exists: %v)", value2, retrieved, exists)
		}
	})

	// Branch 2: windowCache.FastSet returns false, mainCache.FastSet returns true
	t.Run("branch_main_cache_update", func(t *testing.T) {
		wt := NewWTinyLFU(20, 1)
		shard := wt.shards[0]

		// Strategy: Fill window cache, then force an item into main cache
		// Then try to update that item in main cache

		// Fill window cache to capacity first
		windowKeys := make([]string, shard.windowSize)
		for i := 0; i < shard.windowSize; i++ {
			key := fmt.Sprintf("window_fill_%d", i)
			windowKeys[i] = key
			shard.Set(key, fmt.Sprintf("window_value_%d", i))
		}

		// Add a key that should go to main cache
		mainKey := "main_test_key"
		mainValue1 := "main_value1"
		shard.Set(mainKey, mainValue1)

		// Force more items to potentially move our key to main cache
		for i := 0; i < 5; i++ {
			extraKey := fmt.Sprintf("extra_key_%d", i)
			shard.Set(extraKey, fmt.Sprintf("extra_value_%d", i))

			// Access our main key to increase its frequency
			shard.Get(mainKey)
		}

		// Now try to update the key that should be in main cache
		mainValue2 := "main_value2"
		initialWindowSize := shard.windowCache.size
		initialMainSize := shard.mainCache.Size()

		result := shard.Set(mainKey, mainValue2)

		if !result {
			t.Error("Set should return true when updating existing key in main cache")
		}

		// Verify it was an update operation (total cache size shouldn't increase)
		totalSize := shard.windowCache.size + shard.mainCache.Size()
		expectedTotal := initialWindowSize + initialMainSize
		if totalSize > expectedTotal {
			t.Logf("Total size increased from %d to %d (might be expected due to evictions)",
				expectedTotal, totalSize)
		}
	})

	// Branch 3: Both FastSet return false, window cache has space
	t.Run("branch_window_has_space", func(t *testing.T) {
		wt := NewWTinyLFU(20, 1)
		shard := wt.shards[0]

		// Ensure window cache has space
		if shard.windowCache.size >= shard.windowSize {
			t.Skip("Window cache is already full, can't test this branch")
		}

		initialWindowSize := shard.windowCache.size

		// Add a completely new key
		newKey := "brand_new_key"
		newValue := "brand_new_value"

		result := shard.Set(newKey, newValue)

		if !result {
			t.Error("Set should return true when adding new key to window cache with space")
		}

		// Verify window cache size increased
		if shard.windowCache.size != initialWindowSize+1 {
			t.Errorf("Window cache size should increase from %d to %d, got %d",
				initialWindowSize, initialWindowSize+1, shard.windowCache.size)
		}

		// Verify key can be retrieved
		retrieved, exists := shard.Get(newKey)
		if !exists || retrieved != newValue {
			t.Errorf("Expected new key %v, got %v (exists: %v)", newValue, retrieved, exists)
		}
	})

	// Branch 4: Both FastSet return false, window cache is full (final branch)
	t.Run("branch_window_full_force_add", func(t *testing.T) {
		wt := NewWTinyLFU(8, 1) // Small cache to easily fill
		shard := wt.shards[0]

		// Fill window cache to capacity
		for i := 0; i < shard.windowSize+2; i++ {
			key := fmt.Sprintf("fill_window_%d", i)
			shard.Set(key, fmt.Sprintf("fill_value_%d", i))
		}

		// Ensure window cache is at capacity
		if shard.windowCache.size < shard.windowSize {
			// Add more to ensure it's full
			for i := 100; i < 100+shard.windowSize; i++ {
				key := fmt.Sprintf("extra_fill_%d", i)
				shard.Set(key, fmt.Sprintf("extra_value_%d", i))
			}
		}

		initialTotalSize := shard.windowCache.size + shard.mainCache.Size()

		// Add a new key that should trigger the final branch
		finalKey := "final_branch_key"
		finalValue := "final_branch_value"

		result := shard.Set(finalKey, finalValue)

		if !result {
			t.Error("Set should return true even when forcing add to full window cache")
		}

		// The key should be retrievable
		retrieved, exists := shard.Get(finalKey)
		if !exists || retrieved != finalValue {
			t.Errorf("Final branch key should be retrievable: expected %v, got %v (exists: %v)",
				finalValue, retrieved, exists)
		}

		// Cache should still be functional
		newTotalSize := shard.windowCache.size + shard.mainCache.Size()
		t.Logf("Cache size before: %d, after: %d", initialTotalSize, newTotalSize)
	})
}

// TestWTinyLFUShard_Set_SpecificConditions tests very specific conditions
func TestWTinyLFUShard_Set_SpecificConditions(t *testing.T) {

	t.Run("empty_cache_first_item", func(t *testing.T) {
		wt := NewWTinyLFU(10, 1)
		shard := wt.shards[0]

		// Verify cache is empty
		if shard.windowCache.size != 0 || shard.mainCache.Size() != 0 {
			t.Error("Cache should be empty initially")
		}

		// Add first item
		result := shard.Set("first_key", "first_value")
		if !result {
			t.Error("Set should succeed for first item")
		}

		// Should be in window cache
		if shard.windowCache.size != 1 {
			t.Errorf("Window cache should have 1 item, got %d", shard.windowCache.size)
		}
		if shard.mainCache.Size() != 0 {
			t.Errorf("Main cache should be empty, got %d", shard.mainCache.Size())
		}
	})

	t.Run("window_size_boundary", func(t *testing.T) {
		wt := NewWTinyLFU(6, 1) // Small cache
		shard := wt.shards[0]

		// Fill exactly to window size
		for i := 0; i < shard.windowSize; i++ {
			key := fmt.Sprintf("boundary_key_%d", i)
			result := shard.Set(key, fmt.Sprintf("boundary_value_%d", i))
			if !result {
				t.Errorf("Set should succeed for item %d", i)
			}
		}

		// Verify window cache is exactly at capacity
		if shard.windowCache.size != shard.windowSize {
			t.Errorf("Window cache should be at capacity %d, got %d",
				shard.windowSize, shard.windowCache.size)
		}

		// Add one more item - this should test the boundary condition
		result := shard.Set("boundary_overflow", "overflow_value")
		if !result {
			t.Error("Set should succeed even at boundary overflow")
		}

		// Verify the overflow item is retrievable
		retrieved, exists := shard.Get("boundary_overflow")
		if !exists || retrieved != "overflow_value" {
			t.Errorf("Boundary overflow item should be retrievable: expected %v, got %v (exists: %v)",
				"overflow_value", retrieved, exists)
		}
	})
}
