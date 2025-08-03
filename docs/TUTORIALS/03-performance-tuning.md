# Tutorial: Performance Tuning

Metis is designed for high performance out of the box, but understanding how to tune its configuration can help you extract maximum efficiency for your specific workload. This tutorial will guide you through some key configuration parameters and how they impact performance.

## Prerequisites

- Completed the [Basic Usage Tutorial](./01-basic-usage.md).
- Familiarity with the [Configuration Guide](../CONFIGURATION.md).

## The Scenario

We will explore how changing the number of shards and enabling compression affects the cache's throughput and memory usage. We will use Go's built-in benchmarking tools to measure these effects.

## Key Performance Levers

1.  **Shards**: The number of concurrent partitions in the cache.
2.  **Compression**: Gzip compression for stored values.
3.  **Eviction Policy**: `WTinyLFU` vs. `LRU`.

## Step 1: Create a Benchmark File

Let's create a new file named `tuning_test.go` to house our benchmarks.

```go
package main

import (
    "fmt"
    "testing"

    "github.com/agilira/metis"
)

// A sample struct to store in the cache.
type ComplexObject struct {
    ID      int
    Name    string
    Data    []byte
    Metrics map[string]float64
}

func newComplexObject(i int) ComplexObject {
    return ComplexObject{
        ID:   i,
        Name: fmt.Sprintf("Object-%d", i),
        Data: make([]byte, 1024), // 1KB of data
        Metrics: map[string]float64{
            "metric1": float64(i),
            "metric2": float64(i * 2),
        },
    }
}

// Benchmark for a cache with a low shard count.
func BenchmarkLowShards(b *testing.B) {
    config := metis.CacheConfig{
        CacheSize:      10000,
        EvictionPolicy: "wtinylfu",
        Shards:         2, // Low shard count
    }
    cache := metis.NewWithConfig(config)
    defer cache.Close()

    b.RunParallel(func(pb *testing.PB) {
        i := 0
        for pb.Next() {
            key := fmt.Sprintf("key-%d", i)
            cache.Set(key, newComplexObject(i))
            i++
        }
    })
}

// Benchmark for a cache with a high shard count.
func BenchmarkHighShards(b *testing.B) {
    config := metis.CacheConfig{
        CacheSize:      10000,
        EvictionPolicy: "wtinylfu",
        Shards:         32, // High shard count
    }
    cache := metis.NewWithConfig(config)
    defer cache.Close()

    b.RunParallel(func(pb *testing.PB) {
        i := 0
        for pb.Next() {
            key := fmt.Sprintf("key-%d", i)
            cache.Set(key, newComplexObject(i))
            i++
        }
    })
}

// Benchmark for a cache with compression enabled.
func BenchmarkWithCompression(b *testing.B) {
    config := metis.CacheConfig{
        CacheSize:      10000,
        EvictionPolicy: "wtinylfu",
        Shards:         16, // A balanced shard count
        Compress:       true,
    }
    cache := metis.NewWithConfig(config)
    defer cache.Close()

    b.RunParallel(func(pb *testing.PB) {
        i := 0
        for pb.Next() {
            key := fmt.Sprintf("key-%d", i)
            cache.Set(key, newComplexObject(i))
            i++
        }
    })
}
```

## Step 2: Run the Benchmarks

Execute the benchmarks from your terminal. We will run them with the `-benchmem` flag to see memory allocation statistics.

```bash
go test -bench=. -benchmem
```

## Step 3: Analyze the Results

You will see output similar to this (results will vary based on your machine):

```
goos: linux
goarch: amd64
pkg: github.com/agilira/metis
cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
BenchmarkLowShards-12             358336          3261 ns/op        1369 B/op          8 allocs/op
BenchmarkHighShards-12            684933          1675 ns/op        1372 B/op          8 allocs/op
BenchmarkWithCompression-12       183974          6433 ns/op        1584 B/op         13 allocs/op
```

### Analysis

1.  **`BenchmarkLowShards` vs. `BenchmarkHighShards`**:
    *   **`ns/op` (Nanoseconds per Operation)**: `BenchmarkHighShards` is significantly faster (1675 ns/op vs. 3261 ns/op). This is because more shards reduce lock contention, allowing the `b.RunParallel` goroutines to perform `Set` operations with less waiting.
    *   **Conclusion**: For highly concurrent workloads, increasing the number of shards is a critical performance tuning step. A good rule of thumb is to use a power of 2 near the number of available CPU cores.

2.  **`BenchmarkWithCompression`**:
    *   **`ns/op`**: This benchmark is the slowest (6433 ns/op). The overhead of Gzip compression adds significant latency to each `Set` operation.
    *   **`B/op` (Bytes per Operation)**: While the `B/op` here shows the allocations during the operation itself, the real benefit is in the total memory footprint of the cache, which is not directly measured by this benchmark metric. With compression, the 1KB `Data` field in our object would be stored much more compactly.
    *   **Conclusion**: Enable compression when your primary constraint is memory usage and you are storing large, compressible values. Be prepared to trade CPU cycles for memory savings.

## General Tuning Advice

-   **Start with Defaults**: Metis's defaults are sensible for a wide range of applications.
-   **Profile Your Application**: Before tuning, use tools like the Go profiler to identify bottlenecks. Don't optimize prematurely.
-   **Shards are Key for Concurrency**: If your application has many goroutines accessing the cache simultaneously, `Shards` is your most important tuning parameter.
-   **Use Compression for Memory Savings**: If your cache is consuming too much memory and your values are compressible (like JSON, text, or certain binary formats), enable `Compress`.
-   **`WTinyLFU` is Usually Best**: The `WTinyLFU` eviction policy generally provides a better hit ratio than `LRU` for workloads with mixed access patterns (some popular items, some one-hit-wonders). Stick with it unless you have a very specific, LRU-friendly access pattern.
