// types.go: Core types for Metis strategic caching library
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package metis

import (
	"container/list"
	"sync"
	"time"
)

// Object pools for memory optimization
var (
	entryPool = sync.Pool{
		New: func() interface{} {
			return &CacheEntry{}
		},
	}
)

// GetCacheEntry gets a cache entry from the pool
func GetCacheEntry() *CacheEntry {
	return entryPool.Get().(*CacheEntry)
}

// PutCacheEntry returns a cache entry to the pool after clearing it
func PutCacheEntry(entry *CacheEntry) {
	if entry != nil {
		// Clear the entry to prevent memory leaks
		entry.Key = ""
		entry.Data = nil
		entry.Timestamp = time.Time{}
		entry.LastAccess = time.Time{}
		entry.AccessCount = 0
		entry.Size = 0
		entry.Compressed = false
		entry.IsNil = false
		entry.llElem = nil
		entryPool.Put(entry)
	}
}

// Logger interface for optional debug and monitoring logging
type Logger interface {
	// Debug logs debug-level messages (cache hits, misses, etc.)
	Debug(msg string, fields ...interface{})
	// Info logs informational messages (cache operations, config changes)
	Info(msg string, fields ...interface{})
	// Warn logs warning messages (potential issues, degraded performance)
	Warn(msg string, fields ...interface{})
	// Error logs error messages (failed operations, critical issues)
	Error(msg string, fields ...interface{})
}

// CacheConfig defines the configuration for strategic caching
type CacheConfig struct {
	EnableCaching     bool          `json:"enable_caching"`
	CacheSize         int           `json:"cache_size"`
	TTL               time.Duration `json:"ttl"`
	CleanupInterval   time.Duration `json:"cleanup_interval"`
	MaxKeySize        int           `json:"max_key_size"`
	MaxValueSize      int           `json:"max_value_size"`
	EnableCompression bool          `json:"enable_compression"`
	EvictionPolicy    string        `json:"eviction_policy"` // "lru", "lfu", "tinylfu", "wtinylfu" (default: wtinylfu)
	// AdmissionProbability controls the probability (0.0-1.0) that a new item is admitted to the cache (for probabilistic admission policies). Default: -1 (unset, always admit).
	AdmissionProbability float64 `json:"admission_probability,omitempty"`
	// ShardCount controls the number of shards for the cache (striped locking). Default: 16.
	ShardCount int `json:"shard_count,omitempty"`
	// MaxShardSize controls the maximum number of entries per shard. Default: CacheSize / ShardCount.
	MaxShardSize int `json:"max_shard_size,omitempty"`
	// AdmissionPolicy controls the admission policy: "always", "never", "probabilistic". Default: "always".
	AdmissionPolicy string `json:"admission_policy,omitempty"`
	// Logger for debug and monitoring (optional, can be nil)
	Logger Logger `json:"-"`
}

// CacheEntry represents a single entry in the cache
type CacheEntry struct {
	Key         string        `json:"key"` // Key for efficient eviction (backward compatibility)
	Data        interface{}   `json:"data"`
	Timestamp   time.Time     `json:"timestamp"`   // Expiration timestamp
	LastAccess  time.Time     `json:"last_access"` // Last access timestamp for LRU
	AccessCount int64         `json:"access_count"`
	Size        int           `json:"size"`
	Compressed  bool          `json:"compressed"`
	IsNil       bool          `json:"is_nil"` // Flag to distinguish nil values from empty strings
	llElem      *list.Element // Pointer to node in the LRU/LFU list (internal use)
}
