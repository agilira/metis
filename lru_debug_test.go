//lru_debug_test.go: Unit tests for ExactLRU eviction policy in Metis strategic caching library
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package metis

import (
	"testing"
	"time"
)

func TestExactLRU(t *testing.T) {
	cfg := CacheConfig{
		EnableCaching:   true,
		CacheSize:       2,
		TTL:             24 * time.Hour, // Very long TTL to avoid expiration during test
		EvictionPolicy:  "lru",
		AdmissionPolicy: "always",
		ShardCount:      1, // For deterministic eviction in tests
	}
	sc := NewStrategicCache(cfg)
	sc.Set("a", "1")
	time.Sleep(10 * time.Millisecond) // ensure timestamp difference
	sc.Set("b", "2")
	time.Sleep(10 * time.Millisecond) // ensure timestamp difference
	// Access "a" to make "b" the least recently used
	_, _ = sc.Get("a")
	time.Sleep(10 * time.Millisecond) // ensure timestamp difference
	sc.Set("c", "3")                  // should evict "b"

	if _, ok := sc.Get("b"); ok {
		t.Error("expected 'b' to be evicted by LRU policy")
	}
	if _, ok := sc.Get("a"); !ok {
		t.Error("expected 'a' to remain in cache")
	}
	if _, ok := sc.Get("c"); !ok {
		t.Error("expected 'c' to be in cache")
	}
	sc.Close()
}
