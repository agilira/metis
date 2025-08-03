# Metis: A High-Performance Go Caching Library
### an AGILira fragment

Metis is a sharded, high-performance caching library for Go, designed for applications that demand speed, scalability, and observability. It provides precise control over eviction policies, memory usage, and configuration, making it suitable for a wide range of use cases, from web servers to high-throughput API gateways.

[![CI/CD Pipeline](https://github.com/agilira/metis/workflows/CI/CD%20Pipeline/badge.svg)](https://github.com/agilira/metis/actions?query=workflow%3A%22CI%2FCD+Pipeline%22)
[![Security Scan](https://github.com/agilira/metis/workflows/Security%20Scan/badge.svg)](https://github.com/agilira/metis/actions?query=workflow%3A%22Security+Scan%22)
[![Coverage](https://img.shields.io/badge/dynamic/json?url=https://raw.githubusercontent.com/agilira/metis/main/coverage-badge.json&label=coverage&query=$.message&color=brightgreen)](https://github.com/agilira/metis)

## Core Features

- **Advanced Eviction Policies**: Implements **WTinyLFU** (Windowed TinyLFU) for high hit rates and **LRU** (Least Recently Used) for general-purpose caching.
- **Sharded Architecture**: Distributes data across multiple shards to minimize lock contention and maximize concurrency in multi-threaded applications.
- **Flexible Configuration**: Configure the cache programmatically, via a `metis.json` file, or rely on environment variables. Configuration loading follows a clear precedence: programmatic > `metis.json` > defaults.
- **Smart Validator**: Includes a configuration validator that analyzes your settings and provides actionable suggestions to optimize performance and memory usage.
- **Built-in Compression**: Optionally compresses cache values using Gzip to reduce memory footprint.
- **Robust Serialization**: Uses Go's `gob` package to reliably serialize a wide range of data types, including complex nested structures.
- **Observability**: Exposes detailed statistics (hits, misses, hit rate) and includes tools for performance benchmarking and profiling.

## Installation

To add Metis to your project, run:
```bash
go get github.com/agilira/metis
```

## Quick Start

Create a `metis.json` file in your project's root directory:

```json
{
  "cache_size": 10000,
  "ttl": "15m",
  "eviction_policy": "wtinylfu"
}
```

Then, use `metis.New()` to create a cache instance that automatically uses this configuration:

```go
package main

import (
    "fmt"
    "github.com/agilira/metis"
)

func main() {
    // Metis automatically loads metis.json if present
    cache := metis.New()
    defer cache.Close()

    cache.Set("user:123", "premium_user")
    value, found := cache.Get("user:123")

    if found {
        fmt.Printf("Found value: %s\n", value)
    }

    stats := cache.Stats()
    fmt.Printf("Cache Stats: %+v\n", stats)
}
```

## Configuration

Metis offers multiple ways to configure your cache.

### 1. Programmatic Configuration

For the highest level of control, create a `CacheConfig` struct and pass it to `NewWithConfig`. This will override any other configuration source.

```go
import "time"

config := metis.CacheConfig{
    CacheSize:         50000,
    ShardCount:        16, // Recommended: a power of 2
    EvictionPolicy:    "wtinylfu",
    TTL:               30 * time.Minute,
    EnableCompression: true,
    MaxValueSize:      1024, // 1 KB
}

cache := metis.NewWithConfig(config)
```

### 2. JSON Configuration

Place a `metis.json` file in your application's root. `metis.New()` will automatically detect and load it.

**Example `metis.json`:**
```json
{
  "cache_size": 50000,
  "ttl": "30m",
  "eviction_policy": "wtinylfu",
  "shard_count": 32,
  "enable_compression": true
}
```

### 3. Use-Case Based Presets

Metis provides convenient constructors for common use cases:

- `metis.NewWebServerCache()`
- `metis.NewAPIGatewayCache()`
- `metis.NewDevelopmentCache()`

## Benchmarking

To run the built-in benchmarks and evaluate Metis's performance on your system:

```bash
go test -bench=. -benchmem
```

## Contributing

Contributions are highly welcome. Please read our [CONTRIBUTING.md](./CONTRIBUTING.md) guide for details on the process.

## License

Metis is licensed under the [Mozilla Public License 2.0](./LICENSE).

## Documentation
For detailed documentation, visit the [docs](./docs/) folder.

---

Metis • an AGILira fragment