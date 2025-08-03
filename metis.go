// metis.go: Strategic caching library for high-performance applications
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package metis

import (
	"bytes"
	"compress/gzip"
	"container/list"
	"context"
	randc "crypto/rand"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"hash/crc32"
	"io"
	"reflect"
	"sync"
	"time"
)

func init() {
	// Register PrimitiveBox type for robust gob encoding/decoding
	gob.Register(PrimitiveBox{})
	// Register common primitive types that will be contained in PrimitiveBox.V
	gob.Register(int(0))
	gob.Register(int32(0))
	gob.Register(int64(0))
	gob.Register(uint(0))
	gob.Register(uint32(0))
	gob.Register(uint64(0))
	gob.Register(float32(0))
	gob.Register(float64(0))
	gob.Register(bool(false))
	gob.Register(string(""))
	gob.Register([]byte{})

	// Register common interface types for robust serialization
	gob.Register(map[string]interface{}{})
	gob.Register(map[string]string{})
	gob.Register(map[string]int{})
	gob.Register(map[string]float64{})
	gob.Register(map[string]bool{})
	gob.Register([]interface{}{})
	gob.Register([]string{})
	gob.Register([]int{})
	gob.Register([]float64{})
	gob.Register([]bool{})
}

// cacheShard represents a single shard of the cache, with its own map, mutex, and LRU/LFU list
type cacheShard struct {
	data   map[string]*CacheEntry
	mu     sync.RWMutex
	ll     *list.List // Doubly-linked list for LRU/LFU optimization
	hits   int64
	misses int64
}

// EvictionPolicy defines the interface for cache eviction strategies
// The policy decides which key to evict when the cache is full
type EvictionPolicy interface {
	EvictKey(cache map[string]*CacheEntry, ll *list.List) string
}

// LRUPolicy evicts the least recently used entry
// (uses Timestamp as last access time)
type LRUPolicy struct{}

// EvictKey selects the key to evict based on LRU policy.
func (p *LRUPolicy) EvictKey(cache map[string]*CacheEntry, ll *list.List) string {
	// The key to evict is always the last element in the list
	if ll == nil || ll.Back() == nil {
		return ""
	}
	entry, ok := ll.Back().Value.(*CacheEntry)
	if !ok {
		return ""
	}
	return entry.Key
}

// AdmissionPolicy defines the interface for cache admission strategies
// The policy decides whether to admit a new key-value pair into the cache
type AdmissionPolicy interface {
	Allow(key string, value interface{}) bool
}

// AlwaysAdmitPolicy always admits new entries
type AlwaysAdmitPolicy struct{}

// Allow always returns true
func (p *AlwaysAdmitPolicy) Allow(key string, value interface{}) bool { return true }

// ProbabilisticAdmissionPolicy admits entries with a given probability
type ProbabilisticAdmissionPolicy struct {
	Probability float64
}

// Allow returns true with the configured probability
func (p *ProbabilisticAdmissionPolicy) Allow(key string, value interface{}) bool {
	// Handle edge cases
	if p.Probability <= 0.0 {
		return false
	}
	if p.Probability >= 1.0 {
		return true
	}
	return SecureFloat64() < p.Probability
}

// NeverAdmitPolicy never admits new entries
type NeverAdmitPolicy struct{}

// Allow always returns false
func (p *NeverAdmitPolicy) Allow(key string, value interface{}) bool { return false }

// SecureFloat64 returns a cryptographically secure random float64 in [0,1)
func SecureFloat64() float64 {
	var b [8]byte
	_, err := randc.Read(b[:])
	if err != nil {
		return 0.0
	}
	return float64(binary.LittleEndian.Uint64(b[:])) / (1 << 64)
}

// PrimitiveBox wraps primitive types for robust gob encoding/decoding
type PrimitiveBox struct {
	V interface{}
}

// StrategicCache provides high-performance, thread-safe caching with multiple eviction policies,
// sharding, TTL support, compression, and comprehensive statistics
type StrategicCache struct {
	config     CacheConfig
	shards     []cacheShard
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	closed     bool
	closedMu   sync.RWMutex // Protect the closed field
	policy     EvictionPolicy
	admission  AdmissionPolicy
	shardCount uint32
	entryPool  *EntryPool // Object pool for CacheEntry reuse
	wtinylfu   *WTinyLFU  // W-TinyLFU eviction policy (when enabled)
}

// getShard returns the appropriate shard for a given key
func (sc *StrategicCache) getShard(key string) *cacheShard {
	var hash uint32

	// Fast path for short strings (most common case)
	if len(key) <= 8 {
		// Simple hash for short strings
		for i := 0; i < len(key); i++ {
			hash = hash*31 + uint32(key[i])
		}
	} else {
		// Use CRC32 for longer strings
		hash = crc32.ChecksumIEEE([]byte(key))
	}

	// Safe conversion since shardCount is validated in constructor
	shardIndex := int(hash % sc.shardCount)
	if shardIndex < 0 || shardIndex >= len(sc.shards) {
		// Fallback to first shard if index is out of bounds
		shardIndex = 0
	}
	return &sc.shards[shardIndex]
}

// NewStrategicCache creates a new strategic cache with the given configuration
func NewStrategicCache(config CacheConfig) *StrategicCache {
	// Set optimized defaults for maximum performance
	if config.CacheSize <= 0 {
		config.CacheSize = 10000 // Increased default cache size
	}
	if config.TTL <= 0 {
		config.TTL = 10 * time.Minute // Longer TTL for better hit rates
	}
	if config.CleanupInterval <= 0 {
		config.CleanupInterval = 2 * time.Minute // Less frequent cleanup
	}
	if config.ShardCount <= 0 {
		config.ShardCount = 32 // More shards for better concurrency
	}
	if config.ShardCount > 1<<30 {
		config.ShardCount = 1 << 30 // Limit to prevent overflow
	}
	if config.MaxShardSize <= 0 {
		config.MaxShardSize = config.CacheSize / config.ShardCount
	}

	// Create context for cleanup goroutines
	ctx, cancel := context.WithCancel(context.Background())

	// Ensure ShardCount is within uint32 range
	shardCount := config.ShardCount
	if shardCount > 1<<31-1 {
		shardCount = 1<<31 - 1
	}
	if shardCount < 1 || shardCount > int(^uint32(0)) {
		shardCount = 1
	}

	sc := &StrategicCache{
		config:     config,
		shards:     make([]cacheShard, config.ShardCount),
		ctx:        ctx,
		cancel:     cancel,
		shardCount: uint32(shardCount), // nosec G115 - Safe: shardCount is validated to be > 0 and <= MaxShardCount
	}

	// Initialize shards
	for i := 0; i < config.ShardCount; i++ {
		sc.shards[i] = cacheShard{
			data: make(map[string]*CacheEntry),
			ll:   list.New(),
		}
	}

	// Initialize EntryPool for CacheEntry reuse
	sc.entryPool = NewEntryPool()

	// Set eviction policy (W-TinyLFU is the best performing default for large caches)
	switch config.EvictionPolicy {
	case "lru":
		sc.policy = &LRUPolicy{}
	case "wtinylfu":
		// Initialize W-TinyLFU (highest priority - best performance)
		sc.wtinylfu = NewWTinyLFU(config.CacheSize, int(config.ShardCount))
		sc.wtinylfu.SetTTL(config.TTL) // Set TTL for W-TinyLFU
		sc.policy = &LRUPolicy{}       // W-TinyLFU handles its own eviction internally
	case "", "default":
		// For small caches (< 1000), use LRU instead of W-TinyLFU
		// W-TinyLFU works best with larger caches
		if config.CacheSize < 1000 {
			sc.policy = &LRUPolicy{}
		} else {
			// Initialize W-TinyLFU for large caches
			sc.wtinylfu = NewWTinyLFU(config.CacheSize, int(config.ShardCount))
			sc.wtinylfu.SetTTL(config.TTL) // Set TTL for W-TinyLFU
			sc.policy = &LRUPolicy{}       // W-TinyLFU handles its own eviction internally
		}
	default:
		// Default to LRU for maximum compatibility
		sc.policy = &LRUPolicy{}
	}

	// Set admission policy (always is the safest default)
	switch config.AdmissionPolicy {
	case "never":
		sc.admission = &NeverAdmitPolicy{}
	case "probabilistic":
		// Use probabilistic admission for better cache efficiency
		probability := config.AdmissionProbability
		if probability < 0 {
			probability = 0.5 // Only use default for negative values
		}
		sc.admission = &ProbabilisticAdmissionPolicy{Probability: probability}
	case "always", "":
		// Default to always for maximum compatibility
		sc.admission = &AlwaysAdmitPolicy{}
	default:
		// Default to always for maximum compatibility
		sc.admission = &AlwaysAdmitPolicy{}
	}

	// Start cleanup goroutines if TTL is enabled
	if config.TTL > 0 {
		for i := 0; i < config.ShardCount; i++ {
			sc.wg.Add(1)
			go sc.cleanupRoutine(i)
		}
	}

	return sc
}

// cleanupRoutine runs the cleanup loop for a specific shard
func (sc *StrategicCache) cleanupRoutine(shardIdx int) {
	defer sc.wg.Done()
	ticker := time.NewTicker(sc.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sc.cleanupExpired(shardIdx)
		case <-sc.ctx.Done():
			return
		}
	}
}

// cleanupExpired removes expired entries from a shard
func (sc *StrategicCache) cleanupExpired(shardIdx int) {
	shard := &sc.shards[shardIdx]
	shard.mu.Lock()
	defer shard.mu.Unlock()

	now := time.Now()
	for key, entry := range shard.data {
		if !entry.Timestamp.IsZero() && now.After(entry.Timestamp) {
			// Remove from linked list
			shard.ll.Remove(entry.llElem)
			delete(shard.data, key)
			// Return entry to pool for reuse
			sc.entryPool.Put(entry)
		}
	}
}

// Get retrieves a value from the cache
func (sc *StrategicCache) Get(key string) (interface{}, bool) {
	if !sc.config.EnableCaching {
		return nil, false
	}

	sc.closedMu.RLock()
	if sc.closed {
		sc.closedMu.RUnlock()
		return nil, false
	}
	sc.closedMu.RUnlock()

	// Ultra-aggressive fast path: Direct delegation when possible
	if sc.wtinylfu != nil && (sc.config.EvictionPolicy == "wtinylfu" || sc.config.EvictionPolicy == "") {
		return sc.wtinylfu.Get(key)
	}

	// Use sharded cache
	shard := sc.getShard(key)
	shard.mu.Lock()
	entry, exists := shard.data[key]
	if !exists {
		shard.misses++ // Increment misses counter
		shard.mu.Unlock()
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.Timestamp) {
		// Remove expired entry from linked list and map
		if entry.llElem != nil {
			shard.ll.Remove(entry.llElem)
		}
		delete(shard.data, key)
		// Return entry to pool for reuse
		sc.entryPool.Put(entry)
		shard.misses++ // Increment misses counter for expired entry
		shard.mu.Unlock()
		return nil, false
	}

	shard.hits++ // Increment hits counter
	// Update access count and timestamp using EntryPool (within lock)
	sc.entryPool.IncrementAccess(entry)
	// Update last access time for LRU policy
	entry.LastAccess = time.Now()

	// Move to front for LRU policy optimization - always move to front when accessed
	if _, ok := sc.policy.(*LRUPolicy); ok && entry.llElem != nil {
		shard.ll.MoveToFront(entry.llElem)
	}

	// Copy necessary data before releasing lock to avoid race conditions
	isCompressed := entry.Compressed
	isNil := entry.IsNil
	var dataCopy interface{}
	if isCompressed {
		if dataBytes, ok := entry.Data.([]byte); ok {
			// Make a copy of the compressed data
			dataCopy = make([]byte, len(dataBytes))
			copy(dataCopy.([]byte), dataBytes)
		} else {
			dataCopy = entry.Data
		}
	} else {
		dataCopy = entry.Data
	}

	shard.mu.Unlock()

	// Decompress if needed
	if isCompressed {
		if dataBytes, ok := dataCopy.([]byte); ok {
			_, payload, err := decompressGzipWithHeader(dataBytes)
			if err != nil {
				return nil, false
			}
			// The payload is already in the correct format (from toBytes)
			// Handle empty payload (for empty strings, nil values, etc.)
			if len(payload) == 0 {
				// Use the IsNil flag to distinguish between nil and empty string
				if isNil {
					return nil, true
				}
				return "", true
			}

			// Try to decode as gob first, if that fails, treat as string
			buf := getBuffer()
			buf.Write(payload)
			dec := gob.NewDecoder(buf)
			var decoded interface{}
			if err := dec.Decode(&decoded); err == nil {
				putBuffer(buf)
				return decoded, true
			}
			buf.Reset()
			buf.Write(payload)
			dec = gob.NewDecoder(buf)
			var box PrimitiveBox
			if err := dec.Decode(&box); err == nil {
				putBuffer(buf)
				return box.V, true
			}
			putBuffer(buf)

			// If all decoding fails, try to parse as primitive type
			// This handles the case where primitives were converted to strings by toBytes
			payloadStr := string(payload)
			if parsed, ok := parsePrimitiveFromString(payloadStr); ok {
				return parsed, true
			}

			// If all parsing fails, treat as string (common case)
			return payloadStr, true
		}
		return nil, false
	}

	return dataCopy, true
}

// Set stores a value in the cache
func (sc *StrategicCache) Set(key string, value interface{}) bool {
	if !sc.config.EnableCaching {
		return false
	}

	sc.closedMu.RLock()
	if sc.closed {
		sc.closedMu.RUnlock()
		return false
	}
	sc.closedMu.RUnlock()

	// Ultra-aggressive fast path: Direct delegation when possible
	if sc.wtinylfu != nil && (sc.config.EvictionPolicy == "wtinylfu" || sc.config.EvictionPolicy == "") {
		// Skip ALL validations for maximum performance
		if sc.config.MaxKeySize == 0 && sc.config.MaxValueSize == 0 && sc.config.MaxShardSize == 0 {
			// Skip admission policy check if it's "always" (most common case)
			if _, ok := sc.admission.(*AlwaysAdmitPolicy); ok {
				return sc.wtinylfu.Set(key, value)
			}
		}

		// Minimal validation path only if absolutely necessary
		if sc.config.MaxKeySize > 0 && len(key) > sc.config.MaxKeySize {
			return false
		}
		if sc.config.MaxValueSize > 0 {
			valueSize := calculateSize(value)
			if valueSize > sc.config.MaxValueSize {
				return false
			}
		}
		if _, ok := sc.admission.(*AlwaysAdmitPolicy); !ok {
			if !sc.admission.Allow(key, value) {
				return false
			}
		}
		return sc.wtinylfu.Set(key, value)
	}

	// Validate key size
	if sc.config.MaxKeySize > 0 && len(key) > sc.config.MaxKeySize {
		return false
	}

	// Validate value size and serializability
	if sc.config.MaxValueSize > 0 {
		valueSize := calculateSize(value)
		if valueSize > sc.config.MaxValueSize {
			return false
		}
	}

	// Reject non-serializable types (functions, channels, etc.)
	if value != nil {
		valueType := reflect.TypeOf(value)
		if valueType.Kind() == reflect.Func || valueType.Kind() == reflect.Chan {
			return false
		}
	}

	// Check admission policy
	if !sc.admission.Allow(key, value) {
		return false
	}

	// Use sharded cache
	shard := sc.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	// Check if key already exists
	if existingEntry, exists := shard.data[key]; exists {
		// Update existing entry
		existingEntry.Data = value
		existingEntry.AccessCount++
		existingEntry.Timestamp = time.Now().Add(sc.config.TTL) // Set expiration time
		existingEntry.LastAccess = time.Now()                   // Update last access time
		existingEntry.Size = calculateSize(value)

		// Move to front for LRU policy - always move to front when updated
		if _, ok := sc.policy.(*LRUPolicy); ok && existingEntry.llElem != nil {
			shard.ll.MoveToFront(existingEntry.llElem)
		}
		return true
	}

	// Create new entry
	entry := &CacheEntry{
		Key:         key,
		Data:        value,
		AccessCount: 1,
		Timestamp:   time.Now().Add(sc.config.TTL), // Set expiration time
		LastAccess:  time.Now(),                    // Set initial last access time
		Size:        calculateSize(value),
	}

	// Check if we need to evict
	maxShardSize := sc.config.CacheSize / int(sc.shardCount)
	if sc.config.MaxShardSize > 0 {
		maxShardSize = sc.config.MaxShardSize
	}

	if len(shard.data) >= maxShardSize {
		// Use the configured eviction policy
		if sc.policy != nil {
			evictKey := sc.policy.EvictKey(shard.data, shard.ll)
			if evictKey != "" {
				if evictEntry := shard.data[evictKey]; evictEntry != nil {
					// Remove from linked list if it exists
					if evictEntry.llElem != nil {
						shard.ll.Remove(evictEntry.llElem)
					}
					delete(shard.data, evictKey)
				}
			}
		} else {
			// Fallback to timestamp-based eviction
			var oldestKey string
			var oldestTime time.Time
			for k, e := range shard.data {
				if oldestKey == "" || e.Timestamp.Before(oldestTime) {
					oldestKey = k
					oldestTime = e.Timestamp
				}
			}
			if oldestKey != "" {
				if evictEntry := shard.data[oldestKey]; evictEntry != nil && evictEntry.llElem != nil {
					shard.ll.Remove(evictEntry.llElem)
				}
				delete(shard.data, oldestKey)
			}
		}
	}

	// Add to linked list for LRU policy - always add to front
	if _, ok := sc.policy.(*LRUPolicy); ok {
		entry.llElem = shard.ll.PushFront(entry)
	}

	shard.data[key] = entry
	return true
}

// Delete removes a key from the cache
func (sc *StrategicCache) Delete(key string) {
	sc.closedMu.RLock()
	if sc.closed {
		sc.closedMu.RUnlock()
		return
	}
	sc.closedMu.RUnlock()

	// If W-TinyLFU is enabled and no traditional eviction policy is specified, delegate to W-TinyLFU
	if sc.wtinylfu != nil && (sc.config.EvictionPolicy == "wtinylfu" || sc.config.EvictionPolicy == "") {
		sc.wtinylfu.Delete(key)
		return
	}

	shard := sc.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	if entry, exists := shard.data[key]; exists {
		// Remove from linked list
		if entry.llElem != nil {
			shard.ll.Remove(entry.llElem)
		}
		delete(shard.data, key)
		// Return entry to pool for reuse
		sc.entryPool.Put(entry)
	}
}

// Clear removes all entries from the cache
func (sc *StrategicCache) Clear() {
	sc.closedMu.RLock()
	if sc.closed {
		sc.closedMu.RUnlock()
		return
	}
	sc.closedMu.RUnlock()

	// If W-TinyLFU is enabled, clear W-TinyLFU
	if sc.wtinylfu != nil {
		sc.wtinylfu.Clear()
		return
	}

	for i := 0; i < int(sc.shardCount); i++ {
		shard := &sc.shards[i]
		shard.mu.Lock()
		// Return all entries to pool before clearing
		for _, entry := range shard.data {
			sc.entryPool.Put(entry)
		}
		shard.data = make(map[string]*CacheEntry)
		shard.ll.Init()
		shard.mu.Unlock()
	}
}

// CacheStats contains statistics about the cache performance
type CacheStats struct {
	Hits   int64
	Misses int64
	Size   int64
	Keys   int
}

// GetStats returns cache statistics
func (sc *StrategicCache) GetStats() CacheStats {
	sc.closedMu.RLock()
	if sc.closed {
		sc.closedMu.RUnlock()
		return CacheStats{}
	}
	sc.closedMu.RUnlock()

	// If W-TinyLFU is enabled, get stats from W-TinyLFU
	if sc.wtinylfu != nil {
		return sc.wtinylfu.GetStats()
	}

	// Calculate stats from shards
	var totalHits, totalMisses, totalSize int64
	var totalKeys int

	for i := range sc.shards {
		sc.shards[i].mu.RLock()
		shardSize := len(sc.shards[i].data)
		totalKeys += shardSize
		totalHits += sc.shards[i].hits
		totalMisses += sc.shards[i].misses
		sc.shards[i].mu.RUnlock()
	}

	// Calculate total size
	totalSize = int64(totalKeys)

	return CacheStats{
		Hits:   totalHits,
		Misses: totalMisses,
		Size:   totalSize,
		Keys:   totalKeys,
	}
}

// Close closes the cache and stops the cleanup goroutines
func (sc *StrategicCache) Close() {
	sc.closedMu.Lock()
	if sc.closed {
		sc.closedMu.Unlock()
		return
	}
	sc.closed = true
	sc.closedMu.Unlock()
	sc.cancel()
	done := make(chan struct{})
	go func() {
		sc.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
	}
	sc.Clear()
}

// Compression helpers
func compressGzipWithHeader(data []byte, header string) ([]byte, error) {
	// Skip compression for small data (compression overhead > benefit)
	if len(data) < 64 {
		var buf bytes.Buffer
		buf.WriteString(header)
		buf.Write(data)
		return buf.Bytes(), nil
	}

	var buf bytes.Buffer
	buf.WriteString(header)
	w := gzip.NewWriter(&buf)
	_, err := w.Write(data)
	if err != nil {
		if closeErr := w.Close(); closeErr != nil {
			return nil, fmt.Errorf("write error: %v, close error: %v", err, closeErr)
		}
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decompressGzipWithHeader(data []byte) (header string, payload []byte, err error) {
	if len(data) < 4 {
		return "", nil, fmt.Errorf("data too short for header")
	}
	header = string(data[:4])

	// If data is exactly 4 bytes, it's just a header
	if len(data) == 4 {
		return header, nil, nil
	}

	// Check if data is compressed (has gzip header)
	if len(data) >= 10 && data[4] == 0x1f && data[5] == 0x8b {
		// Compressed data - use gzip decompression
		r, err := gzip.NewReader(bytes.NewReader(data[4:]))
		if err != nil {
			return header, nil, err
		}
		defer r.Close()
		out, err := io.ReadAll(r)
		if err != nil {
			return header, nil, err
		}
		return header, out, nil
	}

	// For data that's not compressed but has a gzip-like header, return error
	if len(data) >= 6 && data[4] == 0x1f && data[5] == 0x8b {
		return header, nil, fmt.Errorf("invalid gzip data")
	}

	// Uncompressed data - return as-is
	return header, data[4:], nil
}
