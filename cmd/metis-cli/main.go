// /cmd/metis-cli/main.go: CLI tool for easy Metis configuration generation
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// SimpleConfig represents a basic configuration for Metis cache
type SimpleConfig struct {
	CacheSize         int    `json:"cache_size"`
	TTL               string `json:"ttl,omitempty"`
	EvictionPolicy    string `json:"eviction_policy,omitempty"`
	ShardCount        int    `json:"shard_count,omitempty"`
	EnableCompression bool   `json:"enable_compression,omitempty"`
	MaxValueSize      int    `json:"max_value_size,omitempty"`
}

func main() {
	fmt.Println("🚀 Metis Configuration Generator")
	fmt.Println("===================================")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	var config SimpleConfig

	// Ask about use case
	fmt.Println("What's your primary use case?")
	fmt.Println("1. Development/Testing (small, fast)")
	fmt.Println("2. Web Application (balanced)")
	fmt.Println("3. High-Performance API (maximum speed)")
	fmt.Println("4. Memory-Constrained (efficient)")
	fmt.Println("5. Custom configuration")
	fmt.Println("6. Exit")
	fmt.Print("Choose (1-6): ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		config = SimpleConfig{
			CacheSize: 1000,
			TTL:       "10m",
		}
	case "2":
		config = SimpleConfig{
			CacheSize:      50000,
			TTL:            "30m",
			EvictionPolicy: "wtinylfu",
			ShardCount:     32,
		}
	case "3":
		config = SimpleConfig{
			CacheSize:      1000000,
			TTL:            "0s",
			EvictionPolicy: "wtinylfu",
			ShardCount:     128,
		}
	case "4":
		config = SimpleConfig{
			CacheSize:         10000,
			TTL:               "1h",
			EnableCompression: true,
			MaxValueSize:      524288, // 512KB
		}
	case "5":
		config = customConfig(reader)
	case "6":
		fmt.Println("👋 Goodbye!")
		os.Exit(0)
	default:
		fmt.Println("Invalid choice, using development defaults")
		config = SimpleConfig{CacheSize: 1000}
	}

	// Generate metis.json
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Printf("Error generating config: %v\n", err)
		return
	}

	err = os.WriteFile("metis.json", data, 0600)
	if err != nil {
		fmt.Printf("Error writing metis.json: %v\n", err)
		return
	}

	fmt.Println("\n✅ Generated metis.json successfully!")
	fmt.Println("📝 Content:")
	fmt.Println(string(data))
	fmt.Println("\n🚀 You can now use metis.New() in your code!")
}

func customConfig(reader *bufio.Reader) SimpleConfig {
	var config SimpleConfig

	fmt.Print("Cache size (number of entries): ")
	if sizeStr, _ := reader.ReadString('\n'); sizeStr != "" {
		if size, err := strconv.Atoi(strings.TrimSpace(sizeStr)); err == nil {
			config.CacheSize = size
		}
	}

	fmt.Print("TTL (e.g., '30m', '1h', '0s' for no expiration): ")
	if ttl, _ := reader.ReadString('\n'); ttl != "" {
		config.TTL = strings.TrimSpace(ttl)
	}

	fmt.Print("Enable compression? (y/n): ")
	if comp, _ := reader.ReadString('\n'); comp != "" {
		config.EnableCompression = strings.ToLower(strings.TrimSpace(comp)) == "y"
	}

	return config
}
