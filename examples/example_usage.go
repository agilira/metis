// example_usage.go: Example usage of simplified Metis API
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"time"

	"github.com/agilira/metis"
)

func main() {
	fmt.Println("=== Metis Simplified API Examples ===")

	// Show configuration info
	fmt.Println("Configuration Info:")
	fmt.Println(metis.GetConfigInfo())
	fmt.Println()

	// Example 1: Basic usage with automatic configuration
	fmt.Println("1. Basic Cache Usage:")
	cache := metis.New()
	defer cache.Close()

	// Store different types of data
	cache.Set("user:123", map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	})

	cache.Set("session:abc", "active")
	cache.Set("counter", 42)

	// Retrieve data
	if user, exists := cache.Get("user:123"); exists {
		fmt.Printf("   Found user: %v\n", user)
	}

	if session, exists := cache.Get("session:abc"); exists {
		fmt.Printf("   Session status: %v\n", session)
	}

	if counter, exists := cache.Get("counter"); exists {
		fmt.Printf("   Counter value: %v\n", counter)
	}

	// Get statistics
	stats := cache.Stats()
	fmt.Printf("   Cache stats: %s\n\n", stats.String())

	// Example 2: Cache operations
	fmt.Println("2. Cache Operations:")

	// Set multiple items
	for i := 1; i <= 5; i++ {
		key := fmt.Sprintf("item:%d", i)
		value := fmt.Sprintf("value:%d", i)
		cache.Set(key, value)
	}

	// Check size
	fmt.Printf("   Cache size after adding 5 items: %d\n", cache.Size())

	// Delete an item
	cache.Delete("item:3")
	fmt.Printf("   Cache size after deleting item:3: %d\n", cache.Size())

	// Check if deleted item exists
	if _, exists := cache.Get("item:3"); !exists {
		fmt.Println("   item:3 successfully deleted")
	}

	// Clear all items
	cache.Clear()
	fmt.Printf("   Cache size after clear: %d\n", cache.Size())

	// Example 3: Performance demonstration
	fmt.Println("\n3. Performance Demonstration:")

	// Add many items
	start := time.Now()
	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("perf:%d", i)
		cache.Set(key, fmt.Sprintf("data:%d", i))
	}
	setDuration := time.Since(start)

	// Read many items
	start = time.Now()
	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("perf:%d", i)
		cache.Get(key)
	}
	getDuration := time.Since(start)

	fmt.Printf("   Set 10,000 items in: %v (%.0f ops/sec)\n",
		setDuration, float64(10000)/setDuration.Seconds())
	fmt.Printf("   Get 10,000 items in: %v (%.0f ops/sec)\n",
		getDuration, float64(10000)/getDuration.Seconds())

	finalStats := cache.Stats()
	fmt.Printf("   Final stats: %s\n", finalStats.String())

	fmt.Println("\n=== Examples completed successfully! ===")
	fmt.Println("\nConfiguration Options:")
	fmt.Println("1. Create 'metis.json' in your project root for simple configuration:")
	fmt.Println(`   {
     "cache_size": 10000,
     "ttl": "10m"
   }`)
	fmt.Println("\n2. Create 'metis_config.go' for advanced configuration:")
	fmt.Println(`   package main
   
   import "github.com/agilira/metis"
   
   func init() {
       config := metis.CacheConfig{
           CacheSize: 50000,
           TTL: 15 * time.Minute,
           // ... all advanced options
       }
       metis.SetGlobalConfig(config)
   }`)
}
