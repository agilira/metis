// simple_eviction_test.go: Simple tests for Metis strategic caching library
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package metis

import (
	"testing"
)

func TestSimpleEviction(t *testing.T) {
	cfg := CacheConfig{
		EnableCaching:   true,
		CacheSize:       2,
		EvictionPolicy:  "lru",
		AdmissionPolicy: "always",
		ShardCount:      1,
	}
	sc := NewStrategicCache(cfg)
	defer sc.Close()

	// Add first item
	sc.Set("a", "1")
	t.Logf("After adding 'a': cache size = %d", sc.GetStats().Size)

	// Add second item
	sc.Set("b", "2")
	t.Logf("After adding 'b': cache size = %d", sc.GetStats().Size)

	// Access 'a' to make it most recently used
	_, ok := sc.Get("a")
	t.Logf("After accessing 'a': found = %v", ok)

	// Add third item - should evict 'b' (least recently used)
	sc.Set("c", "3")
	t.Logf("After adding 'c': cache size = %d", sc.GetStats().Size)

	// Check what's in the cache
	a, ok := sc.Get("a")
	t.Logf("'a' in cache: %v, value: %v", ok, a)

	b, ok := sc.Get("b")
	t.Logf("'b' in cache: %v, value: %v", ok, b)

	c, ok := sc.Get("c")
	t.Logf("'c' in cache: %v, value: %v", ok, c)

	// Count how many items are in the cache
	count := 0
	if a != nil {
		count++
	}
	if b != nil {
		count++
	}
	if c != nil {
		count++
	}
	t.Logf("Total items in cache: %d", count)

	if count != 2 {
		t.Errorf("Expected 2 items in cache, got %d", count)
	}
}
