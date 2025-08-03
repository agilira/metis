// api.go: Simplified API layer for Metis strategic caching library
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package metis

import (
	"fmt"
)

// Cache provides a simple interface to the full StrategicCache
type Cache struct {
	strategic *StrategicCache
}

// Stats provides simplified cache statistics
type Stats struct {
	Size    int     `json:"size"`
	Hits    int64   `json:"hits"`
	Misses  int64   `json:"misses"`
	HitRate float64 `json:"hit_rate"`
}

// New creates a new cache with automatic configuration loading
// Priority: Go config > JSON config > defaults
func New() *Cache {
	config := loadConfig()
	return &Cache{
		strategic: NewStrategicCache(config),
	}
}

// Set stores a value in the cache
func (c *Cache) Set(key string, value interface{}) {
	c.strategic.Set(key, value)
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	return c.strategic.Get(key)
}

// Delete removes a key from the cache
func (c *Cache) Delete(key string) {
	c.strategic.Delete(key)
}

// Clear removes all items from the cache
func (c *Cache) Clear() {
	c.strategic.Clear()
}

// Size returns the current number of items in the cache
func (c *Cache) Size() int {
	stats := c.strategic.GetStats()
	return stats.Keys
}

// Stats returns simplified cache statistics
func (c *Cache) Stats() Stats {
	s := c.strategic.GetStats()

	total := s.Hits + s.Misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(s.Hits) / float64(total) * 100.0
	}

	return Stats{
		Size:    s.Keys,
		Hits:    s.Hits,
		Misses:  s.Misses,
		HitRate: hitRate,
	}
}

// Close closes the cache and frees resources
func (c *Cache) Close() {
	c.strategic.Close()
}

// String returns a human-readable representation of cache stats
func (s Stats) String() string {
	return fmt.Sprintf("Cache Stats: %d items, %d hits, %d misses, %.1f%% hit rate",
		s.Size, s.Hits, s.Misses, s.HitRate)
}

// NewWithConfig creates a cache with custom configuration for advanced users
func NewWithConfig(config CacheConfig) *Cache {
	return &Cache{
		strategic: NewStrategicCache(config),
	}
}

// GetConfigInfo returns information about the current configuration
func GetConfigInfo() string {
	config := LoadConfig()
	source := GetConfigSource()

	return fmt.Sprintf("Configuration Source: %s\nCache Size: %d\nTTL: %v\nEviction Policy: %s\nShard Count: %d",
		source, config.CacheSize, config.TTL, config.EvictionPolicy, config.ShardCount)
}

// NewHighPerformanceCache creates a high-performance cache with minimal overhead
// This is the fastest way to use Metis - direct WTinyLFU without StrategicCache overhead
func NewHighPerformanceCache(size int) *WTinyLFU {
	return NewWTinyLFU(size, 64) // 64 shards for optimal concurrency
}

// NewForUseCase creates a cache optimized for specific use cases
func NewForUseCase(useCase string) *Cache {
	config := GetConfigRecommendation(useCase)
	return NewWithConfig(config)
}

// NewWebServerCache creates a cache optimized for web servers
func NewWebServerCache() *Cache {
	return NewForUseCase("web-server")
}

// NewAPIGatewayCache creates a cache optimized for API gateways
func NewAPIGatewayCache() *Cache {
	return NewForUseCase("api-gateway")
}

// NewDevelopmentCache creates a cache optimized for development
func NewDevelopmentCache() *Cache {
	return NewForUseCase("development")
}

// NewHighPerformanceCacheWithShards creates a high-performance cache with custom shard count
func NewHighPerformanceCacheWithShards(size, shardCount int) *WTinyLFU {
	return NewWTinyLFU(size, shardCount)
}
