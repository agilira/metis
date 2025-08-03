// entrypool.go: EntryPool implementation for Metis strategic caching library
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

// EntryPool manages a pool of CacheEntry objects for reuse
type EntryPool struct {
	pool sync.Pool
}

// NewEntryPool creates a new EntryPool
func NewEntryPool() *EntryPool {
	return &EntryPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &CacheEntry{
					Data:        nil,
					Timestamp:   time.Time{},
					AccessCount: 0,
					llElem:      nil,
					Key:         "",
					IsNil:       false,
				}
			},
		},
	}
}

// Get retrieves a CacheEntry from the pool
func (ep *EntryPool) Get() *CacheEntry {
	entry := ep.pool.Get().(*CacheEntry)
	return entry
}

// Put returns a CacheEntry to the pool after resetting its fields
func (ep *EntryPool) Put(entry *CacheEntry) {
	if entry == nil {
		return
	}
	// Reset all fields to their zero values (EXISTING CORRECT LOGIC)
	entry.Data = nil
	entry.Timestamp = time.Time{}
	entry.AccessCount = 0
	entry.llElem = nil
	entry.Key = ""
	entry.IsNil = false

	ep.pool.Put(entry) // Return the *same* entry to the pool
}

// CreateEntry creates a new CacheEntry with the given parameters
func (ep *EntryPool) CreateEntry(key string, data interface{}, ttl time.Duration, llElem *list.Element) *CacheEntry {
	entry := ep.Get()
	entry.Key = key
	entry.Data = data
	entry.llElem = llElem

	if ttl > 0 {
		entry.Timestamp = time.Now().Add(ttl)
	} else if ttl < 0 {
		// For negative TTL, set timestamp to past
		entry.Timestamp = time.Now().Add(ttl)
	}

	// Set IsNil flag for nil values
	if data == nil {
		entry.IsNil = true
	}

	return entry
}

// UpdateEntry updates an existing entry with new data
func (ep *EntryPool) UpdateEntry(entry *CacheEntry, data interface{}, ttl time.Duration) {
	entry.Data = data
	entry.AccessCount = 0 // Reset access count on update

	if ttl > 0 {
		entry.Timestamp = time.Now().Add(ttl)
	} else if ttl < 0 {
		// For negative TTL, set timestamp to past
		entry.Timestamp = time.Now().Add(ttl)
	} else {
		entry.Timestamp = time.Time{}
	}

	// Update IsNil flag
	entry.IsNil = (data == nil)
}

// IncrementAccess increments the access count for an entry
func (ep *EntryPool) IncrementAccess(entry *CacheEntry) {
	entry.AccessCount++
}

// IsExpired checks if an entry has expired
func (ep *EntryPool) IsExpired(entry *CacheEntry) bool {
	return !entry.Timestamp.IsZero() && time.Now().After(entry.Timestamp)
}

// ResetEntry resets an entry to its initial state
func (ep *EntryPool) ResetEntry(entry *CacheEntry) {
	entry.Data = nil
	entry.Timestamp = time.Time{}
	entry.AccessCount = 0
	entry.llElem = nil
	entry.Key = ""
	entry.IsNil = false
}
