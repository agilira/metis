// wtinylfu.go: Ultra-optimized W-TinyLFU implementation for Metis strategic caching library
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package metis

import (
	"hash"
	"hash/fnv"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// WTinyLFU implements the W-TinyLFU (Windowed TinyLFU) eviction policy
type WTinyLFU struct {
	shardMask  uint32
	shardCount int
	shards     []*WTinyLFUShard
	disableTTL bool
	hashPool   *sync.Pool
	ttl        time.Duration
}

// WTinyLFUShard contains cache components
type WTinyLFUShard struct {
	windowCache     *FastLRU
	mainCache       *FastSLRU
	admissionFilter *FastTinyLFU
	hits            atomic.Int64
	misses          atomic.Int64
	readMu          sync.RWMutex
	writeMu         sync.Mutex
	windowSize      int
	mainSize        int
	ttl             time.Duration
}

// FastLRU is the LRU implementation
type FastLRU struct {
	data    map[string]*fastNode
	head    *fastNode
	tail    *fastNode
	size    int
	maxSize int
	mu      sync.RWMutex
}

type fastNode struct {
	key   string
	value interface{}
	prev  *fastNode
	next  *fastNode
}

// FastSLRU implements Segmented LRU
type FastSLRU struct {
	probation *FastLRU
	protected *FastLRU
	hits      atomic.Int64
}

// FastTinyLFU implements TinyLFU admission filter with Count-Min Sketch
type FastTinyLFU struct {
	enabled   bool
	size      int
	sketch    [][]uint32 // Count-Min Sketch for frequency estimation
	hashCount int        // Number of hash functions
	resetAt   uint32     // Reset threshold
	counter   uint32     // Global counter for aging
}

// NewWTinyLFU creates an optimized W-TinyLFU cache
func NewWTinyLFU(maxSize, shardCount int) *WTinyLFU {
	if shardCount <= 0 || shardCount > int(^uint32(0)) {
		shardCount = 16 // fallback di sicurezza
	}
	shardCount = nextPowerOf2(shardCount)

	if maxSize <= 0 {
		maxSize = 1000
	}

	wt := &WTinyLFU{
		shardCount: shardCount,
		shardMask:  uint32(shardCount - 1),
		shards:     make([]*WTinyLFUShard, shardCount),
		disableTTL: true,
		hashPool: &sync.Pool{
			New: func() interface{} {
				return fnv.New32a()
			},
		},
	}

	shardSize := maxSize / shardCount
	if shardSize == 0 {
		shardSize = 1 // Minimum shard size
	}

	for i := 0; i < shardCount; i++ {
		// For very small shard sizes, ensure total doesn't exceed maxSize
		var windowSize, mainSize int
		if shardSize == 1 {
			// For tiny shards, use only window cache
			windowSize = 1
			mainSize = 0
		} else {
			windowSize = max(1, shardSize/10)       // Use 10% for window
			mainSize = max(0, shardSize-windowSize) // Remaining for main
		}

		wt.shards[i] = &WTinyLFUShard{
			windowCache:     NewFastLRU(windowSize),
			mainCache:       NewFastSLRU(max(1, mainSize)), // Ensure at least 1 for SLRU
			admissionFilter: NewFastTinyLFU(max(1, shardSize/10)),
			windowSize:      windowSize,
			mainSize:        mainSize,
		}
	}

	return wt
}

// SetTTL sets the time-to-live for cache entries
func (wt *WTinyLFU) SetTTL(ttl time.Duration) {
	wt.ttl = ttl
	wt.disableTTL = (ttl <= 0)
	for _, shard := range wt.shards {
		shard.ttl = ttl
	}
}

// Get retrieves a value from the cache
func (wt *WTinyLFU) Get(key string) (interface{}, bool) {
	if key == "" {
		return nil, false
	}

	h := wt.hashPool.Get().(hash.Hash32)
	h.Reset()
	// nosec
	if _, err := h.Write(*(*[]byte)(unsafe.Pointer(&key))); err != nil {
		wt.hashPool.Put(h)
		return nil, false
	}
	shardIndex := h.Sum32() & wt.shardMask
	wt.hashPool.Put(h)

	shard := wt.shards[shardIndex]
	return shard.Get(key)
}

// Get retrieves a value from the shard
func (shard *WTinyLFUShard) Get(key string) (interface{}, bool) {
	shard.readMu.RLock()

	if value, exists := shard.windowCache.FastGet(key); exists {
		shard.readMu.RUnlock()
		shard.hits.Add(1)
		return value, true
	}

	if value, exists := shard.mainCache.FastGet(key); exists {
		shard.readMu.RUnlock()
		shard.hits.Add(1)
		return value, true
	}

	shard.readMu.RUnlock()
	shard.misses.Add(1)
	return nil, false
}

// Set stores a value in the cache
func (wt *WTinyLFU) Set(key string, value interface{}) bool {
	if key == "" {
		return false
	}

	h := wt.hashPool.Get().(hash.Hash32)
	h.Reset()
	if _, err := h.Write(*(*[]byte)(unsafe.Pointer(&key))); err != nil { // nosec G103
		wt.hashPool.Put(h)
		return false
	}
	shardIndex := h.Sum32() & wt.shardMask
	wt.hashPool.Put(h)

	shard := wt.shards[shardIndex]
	return shard.Set(key, value)
}

// SetGet combines Set and Get operations
func (wt *WTinyLFU) SetGet(key string, value interface{}) (interface{}, bool) {
	wt.Set(key, value)
	return wt.Get(key)
}

// Set stores a value in the shard with admission filter
func (shard *WTinyLFUShard) Set(key string, value interface{}) bool {
	shard.writeMu.Lock()
	defer shard.writeMu.Unlock()

	// Record access in admission filter
	shard.admissionFilter.Record(key)

	// Check if key already exists in window cache
	if shard.windowCache.Exists(key) {
		shard.windowCache.FastSet(key, value)
		return true
	}

	// Check if key already exists in main cache
	if shard.mainCache.Exists(key) {
		shard.mainCache.FastSet(key, value)
		return true
	}

	// Key doesn't exist, decide where to place it
	// For new keys, always try window cache first
	if shard.windowCache.Size() < shard.windowSize {
		shard.windowCache.FastSet(key, value)
		return true
	}

	// Window cache is full, check if main cache has space
	if shard.mainSize > 0 && shard.mainCache.Size() < shard.mainSize {
		shard.mainCache.FastSet(key, value)
		return true
	}

	// Both caches are full, use admission policy with TinyLFU filter
	totalSize := shard.windowCache.Size() + shard.mainCache.Size()
	maxTotal := shard.windowSize + shard.mainSize

	if totalSize >= maxTotal {
		// At capacity, use admission filter to decide
		if maxTotal <= 1 {
			return false // For very small caches, don't exceed capacity
		}

		// Get victim from window cache (LRU victim)
		victimKey := shard.getWindowVictim()
		if victimKey != "" {
			// Use admission filter to decide
			if shard.admissionFilter.ShouldAdmit(key, victimKey) {
				shard.windowCache.FastSet(key, value) // This will evict the victim
				return true
			}
			return false // Admission filter rejected
		}

		// Fallback: evict from window and add new item
		shard.windowCache.FastSet(key, value)
		return true
	}

	// Not at full capacity yet, add to window
	shard.windowCache.FastSet(key, value)
	return true
}

// SetGet combines Set and Get for shard
func (shard *WTinyLFUShard) SetGet(key string, value interface{}) (interface{}, bool) {
	shard.Set(key, value)
	return shard.Get(key)
}

// Delete removes a key from the cache
func (wt *WTinyLFU) Delete(key string) bool {
	if key == "" {
		return false
	}

	h := wt.hashPool.Get().(hash.Hash32)
	h.Reset()
	if _, err := h.Write(*(*[]byte)(unsafe.Pointer(&key))); err != nil { // nosec G103
		wt.hashPool.Put(h)
		return false
	}
	shardIndex := h.Sum32() & wt.shardMask
	wt.hashPool.Put(h)

	shard := wt.shards[shardIndex]
	return shard.Delete(key)
}

// Delete removes a key from the shard
func (shard *WTinyLFUShard) Delete(key string) bool {
	shard.writeMu.Lock()
	defer shard.writeMu.Unlock()

	deleted := false
	if shard.windowCache.Delete(key) {
		deleted = true
	}
	if shard.mainCache.Delete(key) {
		deleted = true
	}
	return deleted
}

// getWindowVictim returns the LRU key from window cache for admission decisions
func (shard *WTinyLFUShard) getWindowVictim() string {
	shard.windowCache.mu.RLock()
	defer shard.windowCache.mu.RUnlock()

	if shard.windowCache.tail.prev != shard.windowCache.head {
		return shard.windowCache.tail.prev.key
	}
	return ""
}

// Clear removes all entries
func (wt *WTinyLFU) Clear() {
	for _, shard := range wt.shards {
		shard.Clear()
	}
}

// Clear removes all entries from shard
func (shard *WTinyLFUShard) Clear() {
	shard.writeMu.Lock()
	defer shard.writeMu.Unlock()

	shard.windowCache.Clear()
	shard.mainCache.Clear()
	shard.hits.Store(0)
	shard.misses.Store(0)
}

// Exists checks if a key exists
func (wt *WTinyLFU) Exists(key string) bool {
	_, exists := wt.Get(key)
	return exists
}

// Size returns total cache size
func (wt *WTinyLFU) Size() int {
	total := 0
	for _, shard := range wt.shards {
		total += shard.Size()
	}
	return total
}

// Size returns shard size
func (shard *WTinyLFUShard) Size() int {
	shard.readMu.RLock()
	defer shard.readMu.RUnlock()
	return shard.windowCache.Size() + shard.mainCache.Size()
}

// MaxSize returns maximum cache size
func (wt *WTinyLFU) MaxSize() int {
	total := 0
	for _, shard := range wt.shards {
		total += shard.windowSize + shard.mainSize
	}
	return total
}

// Hits returns total cache hits
func (wt *WTinyLFU) Hits() int64 {
	total := int64(0)
	for _, shard := range wt.shards {
		total += shard.hits.Load()
	}
	return total
}

// Stats returns cache statistics
func (wt *WTinyLFU) Stats() map[string]interface{} {
	hits := wt.Hits()
	misses := int64(0)
	for _, shard := range wt.shards {
		misses += shard.misses.Load()
	}

	total := hits + misses
	hitRate := float64(0)
	if total > 0 {
		hitRate = float64(hits) / float64(total) * 100
	}

	return map[string]interface{}{
		"size":     wt.Size(),
		"max_size": wt.MaxSize(),
		"hits":     hits,
		"misses":   misses,
		"hit_rate": hitRate,
		"ttl":      wt.ttl,
		"shards":   len(wt.shards),
		// Additional keys expected by tests
		"window_size":     wt.WindowSize(),
		"main_size":       wt.MainSize(),
		"total_size":      wt.Size(),
		"shard_count":     len(wt.shards),
		"total_hits":      hits,
		"admission_stats": wt.shards[0].admissionFilter.Stats(),
	}
}

// GetStats returns cache statistics in CacheStats format for compatibility
func (wt *WTinyLFU) GetStats() CacheStats {
	hits := wt.Hits()
	misses := int64(0)
	for _, shard := range wt.shards {
		misses += shard.misses.Load()
	}

	return CacheStats{
		Hits:   hits,
		Misses: misses,
		Size:   int64(wt.Size()),
		Keys:   wt.Size(),
	}
}

// HealthCheck performs health check
func (wt *WTinyLFU) HealthCheck() map[string]interface{} {
	stats := wt.Stats()
	health := "healthy"

	if wt.Size() >= wt.MaxSize() {
		health = "full"
	}

	stats["health"] = health
	stats["type"] = "wtinylfu"
	return stats
}

// WindowSize returns the window size of the first shard for test compatibility
func (wt *WTinyLFU) WindowSize() int {
	if len(wt.shards) > 0 {
		return wt.shards[0].windowSize
	}
	return 0
}

// MainSize returns the main size of the first shard for test compatibility
func (wt *WTinyLFU) MainSize() int {
	if len(wt.shards) > 0 {
		return wt.shards[0].mainSize
	}
	return 0
}

// AdmissionFilter returns the admission filter of the first shard for test compatibility
func (wt *WTinyLFU) AdmissionFilter() interface{} {
	if len(wt.shards) > 0 {
		return wt.shards[0].admissionFilter
	}
	return nil
}

// NewFastLRU creates a new FastLRU cache with the specified maximum size
func NewFastLRU(maxSize int) *FastLRU {
	lru := &FastLRU{
		data:    make(map[string]*fastNode),
		maxSize: maxSize,
		head:    &fastNode{},
		tail:    &fastNode{},
	}
	lru.head.next = lru.tail
	lru.tail.prev = lru.head
	return lru
}

// FastGet retrieves a value from the cache and moves it to the front
func (lru *FastLRU) FastGet(key string) (interface{}, bool) {
	lru.mu.RLock()
	node, exists := lru.data[key]
	if !exists {
		lru.mu.RUnlock()
		return nil, false
	}
	value := node.value
	lru.mu.RUnlock()

	lru.mu.Lock()
	lru.moveToFront(node)
	lru.mu.Unlock()

	return value, true
}

// FastSet adds or updates a key-value pair in the cache
func (lru *FastLRU) FastSet(key string, value interface{}) bool {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if node, exists := lru.data[key]; exists {
		node.value = value
		lru.moveToFront(node)
		return true
	}

	if lru.size >= lru.maxSize && lru.maxSize > 0 {
		oldest := lru.tail.prev
		if oldest != lru.head && oldest != nil {
			delete(lru.data, oldest.key)
			lru.removeNode(oldest)
			lru.size--
		}
	}

	newNode := &fastNode{
		key:   key,
		value: value,
	}
	lru.data[key] = newNode
	lru.addToFront(newNode)
	lru.size++
	return true // Return true for successful insertion
}

// Delete removes a key-value pair from the cache
func (lru *FastLRU) Delete(key string) bool {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if node, exists := lru.data[key]; exists {
		delete(lru.data, key)
		lru.removeNode(node)
		lru.size--
		return true
	}
	return false
}

// Clear removes all items from the cache
func (lru *FastLRU) Clear() {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	lru.data = make(map[string]*fastNode)
	lru.head.next = lru.tail
	lru.tail.prev = lru.head
	lru.size = 0
}

// Get is an alias for FastGet for test compatibility
func (lru *FastLRU) Get(key string) (interface{}, bool) {
	return lru.FastGet(key)
}

// Set is an alias for FastSet for test compatibility
func (lru *FastLRU) Set(key string, value interface{}) bool {
	return lru.FastSet(key, value)
}

// Exists checks if a key exists in the LRU
func (lru *FastLRU) Exists(key string) bool {
	lru.mu.RLock()
	defer lru.mu.RUnlock()
	_, exists := lru.data[key]
	return exists
}

// Size returns the current number of items in the LRU
func (lru *FastLRU) Size() int {
	lru.mu.RLock()
	defer lru.mu.RUnlock()
	return lru.size
}

func (lru *FastLRU) moveToFront(node *fastNode) {
	lru.removeNode(node)
	lru.addToFront(node)
}

func (lru *FastLRU) removeNode(node *fastNode) {
	node.prev.next = node.next
	node.next.prev = node.prev
}

func (lru *FastLRU) addToFront(node *fastNode) {
	node.prev = lru.head
	node.next = lru.head.next
	lru.head.next.prev = node
	lru.head.next = node
}

// NewFastSLRU creates a new FastSLRU cache with the specified size
func NewFastSLRU(size int) *FastSLRU {
	probationSize := int(float64(size) * 0.8)
	protectedSize := size - probationSize

	return &FastSLRU{
		probation: NewFastLRU(probationSize),
		protected: NewFastLRU(protectedSize),
	}
}

// FastGet retrieves a value from the cache, promoting it to protected if found in probation
func (slru *FastSLRU) FastGet(key string) (interface{}, bool) {
	// Check protected first
	if value, exists := slru.protected.FastGet(key); exists {
		slru.hits.Add(1)
		return value, true
	}

	// Check probation and promote if found
	if value, exists := slru.probation.FastGet(key); exists {
		// Remove from probation and add to protected (promotion)
		slru.probation.Delete(key)
		slru.protected.FastSet(key, value)
		slru.hits.Add(1)
		return value, true
	}

	return nil, false
}

// FastSet adds or updates a key-value pair in the appropriate segment
func (slru *FastSLRU) FastSet(key string, value interface{}) bool {
	// Check if key already exists in protected and update
	slru.protected.mu.RLock()
	_, existsInProtected := slru.protected.data[key]
	slru.protected.mu.RUnlock()

	if existsInProtected {
		return slru.protected.FastSet(key, value)
	}

	// Check if key already exists in probation and update
	slru.probation.mu.RLock()
	_, existsInProbation := slru.probation.data[key]
	slru.probation.mu.RUnlock()

	if existsInProbation {
		return slru.probation.FastSet(key, value)
	}

	// New key: add to probation
	return slru.probation.FastSet(key, value)
}

// Delete removes a key-value pair from both segments
func (slru *FastSLRU) Delete(key string) bool {
	deleted := false
	if slru.protected.Delete(key) {
		deleted = true
	}
	if slru.probation.Delete(key) {
		deleted = true
	}
	return deleted
}

// Clear removes all items from both segments
func (slru *FastSLRU) Clear() {
	slru.protected.Clear()
	slru.probation.Clear()
}

// Size returns the total number of items in both segments
func (slru *FastSLRU) Size() int {
	return slru.protected.Size() + slru.probation.Size()
}

// Get is an alias for FastGet for test compatibility
func (slru *FastSLRU) Get(key string) (interface{}, bool) {
	return slru.FastGet(key)
}

// Set is an alias for FastSet for test compatibility
func (slru *FastSLRU) Set(key string, value interface{}) bool {
	return slru.FastSet(key, value)
}

// Exists checks if a key exists in either segment
func (slru *FastSLRU) Exists(key string) bool {
	if slru.protected.Exists(key) {
		return true
	}
	return slru.probation.Exists(key)
}

// Hits returns total hits from both segments
func (slru *FastSLRU) Hits() int64 {
	return slru.hits.Load()
}

// EvictProbation evicts the oldest item from probation segment
func (slru *FastSLRU) EvictProbation() (string, interface{}) {
	// Find and remove oldest item from probation using atomic operations
	slru.probation.mu.Lock()
	defer slru.probation.mu.Unlock()

	if slru.probation.size > 0 && slru.probation.tail.prev != slru.probation.head {
		oldest := slru.probation.tail.prev
		key := oldest.key
		value := oldest.value
		// Use internal deletion (already have lock)
		delete(slru.probation.data, key)
		slru.probation.removeNode(oldest)
		slru.probation.size--
		return key, value
	}
	return "", nil
}

// PromoteToProtected promotes a key-value pair to the protected segment
func (slru *FastSLRU) PromoteToProtected(key string, value interface{}) bool {
	return slru.protected.Set(key, value)
}

// NewFastTinyLFU creates a new FastTinyLFU admission filter with Count-Min Sketch
func NewFastTinyLFU(size int) *FastTinyLFU {
	if size <= 0 {
		size = 1000
	}

	// Use 4 hash functions for good distribution
	hashCount := 4
	sketchWidth := size * 4 // Width proportional to expected items

	// Initialize Count-Min Sketch
	sketch := make([][]uint32, hashCount)
	for i := range sketch {
		sketch[i] = make([]uint32, sketchWidth)
	}

	return &FastTinyLFU{
		enabled:   true, // ✅ ENABLED!
		size:      size,
		sketch:    sketch,
		hashCount: hashCount,
		resetAt:   uint32(size * 10), // Reset every 10x size accesses
		counter:   0,
	}
}

// Utility functions
func nextPowerOf2(n int) int {
	if n <= 1 {
		return 1
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n++
	return n
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Admission Filter Methods

// Record records an access to a key in the frequency sketch
func (filter *FastTinyLFU) Record(key string) {
	if !filter.enabled {
		return
	}

	// Increment global counter
	filter.counter++

	// Check if we need to reset (aging mechanism)
	if filter.counter >= filter.resetAt {
		filter.reset()
	}

	// Record in all hash functions
	for i := 0; i < filter.hashCount; i++ {
		hash := filter.hash(key, uint32(i))
		index := hash % uint32(len(filter.sketch[i]))
		filter.sketch[i][index]++
	}
}

// Estimate estimates the frequency of a key
func (filter *FastTinyLFU) Estimate(key string) uint32 {
	if !filter.enabled {
		return 1 // Always admit if disabled
	}

	minFreq := uint32(^uint32(0)) // Max uint32

	// Take minimum across all hash functions (Count-Min Sketch property)
	for i := 0; i < filter.hashCount; i++ {
		hash := filter.hash(key, uint32(i))
		index := hash % uint32(len(filter.sketch[i]))
		freq := filter.sketch[i][index]
		if freq < minFreq {
			minFreq = freq
		}
	}

	return minFreq
}

// ShouldAdmit decides if a new key should be admitted to the cache
func (filter *FastTinyLFU) ShouldAdmit(newKey string, victimKey string) bool {
	if !filter.enabled {
		return true
	}

	newFreq := filter.Estimate(newKey)
	victimFreq := filter.Estimate(victimKey)

	// Admit if new item has higher or equal frequency
	return newFreq >= victimFreq
}

// reset halves all counters (aging mechanism)
func (filter *FastTinyLFU) reset() {
	for i := range filter.sketch {
		for j := range filter.sketch[i] {
			filter.sketch[i][j] = filter.sketch[i][j] / 2
		}
	}
	filter.counter = 0
}

// hash generates a hash for the given key and salt
func (filter *FastTinyLFU) hash(key string, salt uint32) uint32 {
	// Simple hash function (FNV-1a variant with salt)
	hash := uint32(2166136261) ^ salt
	for _, b := range []byte(key) {
		hash ^= uint32(b)
		hash *= 16777619
	}
	return hash
}

// Stats returns admission filter statistics
func (filter *FastTinyLFU) Stats() map[string]interface{} {
	return map[string]interface{}{
		"enabled":    filter.enabled,
		"size":       filter.size,
		"counter":    filter.counter,
		"reset_at":   filter.resetAt,
		"hash_count": filter.hashCount,
	}
}
