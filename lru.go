// lru.go: LRU implementation for Metis strategic caching library
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package metis

import (
	"container/list"
	"sync"
)

// LRU implements a thread-safe Least Recently Used cache
type LRU struct {
	capacity int
	list     *list.List
	cache    map[string]*list.Element
	mu       sync.RWMutex
	hits     int64
}

// LRUEntry represents an entry in the LRU cache
type LRUEntry struct {
	Key   string
	Value interface{}
}

// NewLRU creates a new LRU cache with the specified capacity
func NewLRU(capacity int) *LRU {
	return &LRU{
		capacity: capacity,
		list:     list.New(),
		cache:    make(map[string]*list.Element),
	}
}

// Get retrieves a value from the cache
func (lru *LRU) Get(key string) (interface{}, bool) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if elem, exists := lru.cache[key]; exists {
		lru.list.MoveToFront(elem)
		lru.hits++
		return elem.Value.(*LRUEntry).Value, true
	}

	return nil, false
}

// Set adds or updates a value in the cache
func (lru *LRU) Set(key string, value interface{}) bool {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if elem, exists := lru.cache[key]; exists {
		// Update existing entry
		elem.Value.(*LRUEntry).Value = value
		lru.list.MoveToFront(elem)
		return true
	}

	// Create new entry
	entry := &LRUEntry{Key: key, Value: value}
	elem := lru.list.PushFront(entry)
	lru.cache[key] = elem

	// Evict if necessary
	if lru.list.Len() > lru.capacity {
		lru.evict()
	}

	return true
}

// Exists checks if a key exists in the cache
func (lru *LRU) Exists(key string) bool {
	lru.mu.RLock()
	defer lru.mu.RUnlock()

	_, exists := lru.cache[key]
	return exists
}

// Delete removes a key from the cache
func (lru *LRU) Delete(key string) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if elem, exists := lru.cache[key]; exists {
		lru.list.Remove(elem)
		delete(lru.cache, key)
	}
}

// Clear removes all items from the cache
func (lru *LRU) Clear() {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	lru.list.Init()
	lru.cache = make(map[string]*list.Element)
}

// Size returns the current number of items in the cache
func (lru *LRU) Size() int {
	lru.mu.RLock()
	defer lru.mu.RUnlock()

	return lru.list.Len()
}

// MaxSize returns the maximum capacity of the cache
func (lru *LRU) MaxSize() int {
	return lru.capacity
}

// Hits returns the number of cache hits
func (lru *LRU) Hits() int64 {
	lru.mu.RLock()
	defer lru.mu.RUnlock()

	return lru.hits
}

// Evict removes the least recently used item and returns it
func (lru *LRU) Evict() (string, interface{}) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if lru.list.Len() == 0 {
		return "", nil
	}

	elem := lru.list.Back()
	entry := elem.Value.(*LRUEntry)
	lru.list.Remove(elem)
	delete(lru.cache, entry.Key)

	return entry.Key, entry.Value
}

// evict removes the least recently used item
func (lru *LRU) evict() {
	if lru.list.Len() == 0 {
		return
	}

	elem := lru.list.Back()
	entry := elem.Value.(*LRUEntry)
	lru.list.Remove(elem)
	delete(lru.cache, entry.Key)
}
