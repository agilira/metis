// wtinylfu_admission_test.go: Advanced tests for admission policy and edge cases
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package metis

import (
	"fmt"
	"testing"
)

// TestWTinyLFUShard_Set_AdvancedPaths tests the remaining uncovered paths in WTinyLFUShard.Set
// Targeting the 36.4% coverage to push it to 95%+
func TestWTinyLFUShard_Set_AdvancedPaths(t *testing.T) {
	t.Run("window_cache_fastset_false_simulation", func(t *testing.T) {
		// Create a shard with very specific settings
		wt := NewWTinyLFU(4, 1) // Very small cache to force edge cases
		shard := wt.shards[0]

		// Fill the window cache completely
		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("window_key_%d", i)
			value := fmt.Sprintf("window_value_%d", i)
			shard.Set(key, value)
		}

		// Try to add a new key that should trigger different paths
		result := shard.Set("edge_case_key", "edge_case_value")
		if !result {
			t.Error("Set should still return true even in edge cases")
		}
	})

	t.Run("main_cache_fastset_false_simulation", func(t *testing.T) {
		// Create a shard and force main cache to be full
		wt := NewWTinyLFU(6, 1) // Small cache
		shard := wt.shards[0]

		// Fill both window and main caches
		for i := 0; i < 20; i++ {
			key := fmt.Sprintf("fill_key_%d", i)
			value := fmt.Sprintf("fill_value_%d", i)
			shard.Set(key, value)

			// Access some keys to create frequency patterns
			if i%3 == 0 {
				shard.Get(key)
				shard.Get(key)
			}
		}

		// Now try to add a new key
		result := shard.Set("main_edge_key", "main_edge_value")
		if !result {
			t.Error("Set should return true even when main cache is full")
		}
	})

	t.Run("window_size_exactly_equal", func(t *testing.T) {
		wt := NewWTinyLFU(8, 1)
		shard := wt.shards[0]

		// Fill window cache to exactly its size
		expectedWindowSize := shard.windowSize

		// First fill with exactly windowSize items
		for i := 0; i < expectedWindowSize; i++ {
			key := fmt.Sprintf("exact_key_%d", i)
			value := fmt.Sprintf("exact_value_%d", i)
			shard.Set(key, value)
		}

		// Check that windowCache.size == windowSize
		t.Logf("Window cache size: %d, window size: %d", shard.windowCache.size, shard.windowSize)

		// Now add one more item - this should trigger the final path
		result := shard.Set("overflow_key", "overflow_value")
		if !result {
			t.Error("Set should return true for overflow case")
		}

		// Verify the item was added
		retrieved, exists := shard.Get("overflow_key")
		if !exists || retrieved != "overflow_value" {
			t.Errorf("Overflow item should be retrievable, got exists=%v, value=%v", exists, retrieved)
		}
	})

	t.Run("forced_window_cache_path", func(t *testing.T) {
		// Create a scenario where we force the final FastSet path
		wt := NewWTinyLFU(10, 1)
		shard := wt.shards[0]

		// Fill the cache in a specific pattern to force eviction scenarios
		successCount := 0
		for i := 0; i < 25; i++ {
			key := fmt.Sprintf("pattern_key_%d", i)
			value := fmt.Sprintf("pattern_value_%d", i)

			result := shard.Set(key, value)
			if result {
				successCount++
			}

			// Every 5th item, access it multiple times to create frequency
			if i%5 == 0 {
				for j := 0; j < 3; j++ {
					shard.Get(key)
				}
			}
		}

		// With admission filter, expect some rejections in full cache
		if successCount < 10 {
			t.Errorf("Expected at least 10 successful sets, got %d", successCount)
		}
		t.Logf("Admission filter allowed %d out of 25 sets (%.1f%%)", successCount, float64(successCount)/25*100)

		// Final test item - should have reasonable chance of success
		result := shard.Set("final_force_key", "final_force_value")
		t.Logf("Final set result: %v", result)
	})
}

// TestWTinyLFU_Set_MaxSizeZero tests edge case with zero max size
func TestWTinyLFU_Set_MaxSizeZero(t *testing.T) {
	// Create cache with zero size to test edge cases
	wt := NewWTinyLFU(0, 1)

	// Try to set items in zero-size cache
	result := wt.Set("zero_key", "zero_value")
	if !result {
		t.Error("Set should handle zero-size cache gracefully")
	}

	// Verify behavior
	value, exists := wt.Get("zero_key")
	t.Logf("Zero-size cache get result: exists=%v, value=%v", exists, value)
}

// TestWTinyLFU_Set_LargeDataStress tests with large data to stress the system
func TestWTinyLFU_Set_LargeDataStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large data stress test in short mode")
	}

	wt := NewWTinyLFU(100, 2)

	// Test with progressively larger values
	for size := 1; size <= 10000; size *= 10 {
		key := fmt.Sprintf("large_key_%d", size)
		value := string(make([]byte, size))

		result := wt.Set(key, value)
		if !result {
			t.Errorf("Set should handle large data (size %d)", size)
		}

		// Verify we can retrieve it
		retrieved, exists := wt.Get(key)
		if !exists {
			t.Errorf("Should be able to retrieve large data (size %d)", size)
		} else if len(retrieved.(string)) != size {
			t.Errorf("Retrieved data size mismatch: expected %d, got %d", size, len(retrieved.(string)))
		}
	}
}

// TestWTinyLFU_DirectShardAccess tests direct shard access to cover internal paths
func TestWTinyLFU_DirectShardAccess(t *testing.T) {
	wt := NewWTinyLFU(20, 4) // Multiple shards

	// Test each shard individually
	for i, shard := range wt.shards {
		t.Run(fmt.Sprintf("shard_%d", i), func(t *testing.T) {
			// Fill this specific shard
			for j := 0; j < 10; j++ {
				key := fmt.Sprintf("shard_%d_key_%d", i, j)
				value := fmt.Sprintf("shard_%d_value_%d", i, j)

				result := shard.Set(key, value)
				if !result {
					t.Errorf("Direct shard set failed for shard %d", i)
				}

				// Verify retrieval
				retrieved, exists := shard.Get(key)
				if !exists || retrieved != value {
					t.Errorf("Shard %d: expected %v, got %v (exists: %v)", i, value, retrieved, exists)
				}
			}

			// Test window cache overflow on this shard
			for j := 10; j < 20; j++ {
				key := fmt.Sprintf("overflow_%d_key_%d", i, j)
				value := fmt.Sprintf("overflow_%d_value_%d", i, j)
				shard.Set(key, value)
			}
		})
	}
}

// TestWTinyLFU_Set_CacheEvictionScenarios tests specific eviction scenarios
func TestWTinyLFU_Set_CacheEvictionScenarios(t *testing.T) {
	t.Run("window_full_main_full", func(t *testing.T) {
		wt := NewWTinyLFU(12, 1) // Small cache to force evictions
		shard := wt.shards[0]

		// Phase 1: Fill window cache
		for i := 0; i < int(shard.windowSize); i++ {
			key := fmt.Sprintf("w_key_%d", i)
			value := fmt.Sprintf("w_value_%d", i)
			shard.Set(key, value)
		}

		// Phase 2: Force items into main cache by overfilling
		for i := 0; i < int(shard.mainSize); i++ {
			key := fmt.Sprintf("m_key_%d", i)
			value := fmt.Sprintf("m_value_%d", i)
			shard.Set(key, value)
		}

		// Phase 3: Now both caches should be full - test the edge case
		result := shard.Set("final_eviction_key", "final_eviction_value")
		if !result {
			t.Error("Set should succeed even when both caches are full")
		}

		// Verify cache still functions
		retrieved, exists := shard.Get("final_eviction_key")
		if !exists || retrieved != "final_eviction_value" {
			t.Errorf("Final eviction key should be retrievable, got exists=%v, value=%v", exists, retrieved)
		}
	})
}
