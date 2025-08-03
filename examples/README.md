# Metis Examples

This directory contains complete examples demonstrating how to use the Metis caching library.

## Available Eviction Policies

The simplified Metis library now supports two high-performance eviction policies:

- **`wtinylfu`** (Default): W-TinyLFU algorithm optimized for production use with superior hit rates
- **`lru`**: Least Recently Used algorithm for simple, predictable behavior and debugging

## Examples

### Basic Usage

```go
// examples/basic/main.go
package main

import (
    "fmt"
    "github.com/agilira/metis"
)

func main() {
    cache := metis.New()
    defer cache.Close()

    cache.Set("key1", "value1")
    cache.Set("key2", 42)

    if value, exists := cache.Get("key1"); exists {
        fmt.Printf("Found: %v\n", value)
    }

    stats := cache.Stats()
    fmt.Printf("Cache stats: %s\n", stats.String())
}
```

### Configuration Examples

#### JSON Configuration

```json
// metis.json
{
  "cache_size": 5000,
  "ttl": "5m",
  "eviction_policy": "wtinylfu"
}
```

#### Go Configuration

```go
// metis_config.go
package main

import (
    "time"
    "github.com/agilira/metis"
)

func init() {
    config := metis.CacheConfig{
        CacheSize: 50000,
        TTL: 15 * time.Minute,
        EvictionPolicy: "wtinylfu",
        ShardCount: 64,
        EnableCompression: true,
    }
    metis.SetGlobalConfig(config)
}
```

### Advanced Usage

```go
// examples/advanced/main.go
package main

import (
    "fmt"
    "time"
    "github.com/agilira/metis"
)

func main() {
    // Custom configuration
    config := metis.CacheConfig{
        CacheSize: 1000,
        TTL: 1 * time.Minute,
        EvictionPolicy: "lru",
        ShardCount: 16,
    }
    
    cache := metis.NewWithConfig(config)
    defer cache.Close()

    // Performance test
    start := time.Now()
    for i := 0; i < 10000; i++ {
        key := fmt.Sprintf("key_%d", i)
        cache.Set(key, fmt.Sprintf("value_%d", i))
    }
    duration := time.Since(start)
    
    fmt.Printf("Set 10,000 items in: %v\n", duration)
    fmt.Printf("Operations/sec: %.0f\n", float64(10000)/duration.Seconds())
    
    stats := cache.Stats()
    fmt.Printf("Final stats: %s\n", stats.String())
}
```

## Running Examples

```bash
# Run basic example
cd examples/basic
go run main.go

# Run advanced example
cd examples/advanced
go run main.go
```

## Configuration Files

Place configuration files in your project root:

- `metis.json` for simple configuration
- `metis_config.go` for advanced configuration

The examples demonstrate both approaches. 