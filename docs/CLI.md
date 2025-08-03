# Metis CLI Tools

Metis provides two command-line interfaces to help you work with caching configurations and debug cache behavior during development.

## CLI Tools Overview

### 1. Configuration Generator (`metis-cli`)
The primary CLI tool for generating `metis.json` configuration files interactively.

### 2. Debug & Analysis Tool (`metis-debug`)
A developer-focused tool for debugging cache behavior, analyzing performance, and validating configurations.

---

## Configuration Generator (`metis-cli`)

The Metis Configuration CLI (`metis-cli`) is an interactive utility designed to help you quickly generate a `metis.json` configuration file for your cache.

This tool is especially useful when you want to fine-tune your cache settings without writing code, or when you want to create a baseline configuration that can be checked into your project's source control.

### Location

The CLI tool is located in the `cmd/metis-cli` directory of the project.

### How to Run

You can run the CLI directly using `go run`:

```bash
go run ./cmd/metis-cli/main.go
```

## Usage

When you run the command, the CLI will prompt you with a series of questions about your desired cache configuration. It provides sensible defaults for most questions, so you can simply press `Enter` to accept them or provide your own values.

### Interactive Questions

1.  **Cache Size**: The maximum number of items the cache can hold.
2.  **Eviction Policy**: The strategy to use when the cache is full. You can choose between `wtinylfu` (recommended) and `lru`.
3.  **Time-to-Live (TTL)**: The default duration for how long an item should stay in the cache (e.g., `10m`, `1h`).
4.  **Enable Compression**: Whether to enable Gzip compression for cached values. This is useful if you are storing large, compressible objects.
5.  **Number of Shards**: The number of concurrent shards to use. A good starting point is a power of 2 close to the number of CPU cores.
6.  **Output File Name**: The name of the configuration file to be created. The default is `metis.json`.

### Example Session

```
$ go run ./cmd/metis-cli/main.go
? Enter cache size: 1000
? Select eviction policy: WTinyLFU
? Enter default TTL (e.g., 5m, 1h): 10m
? Enable compression? Yes
? Enter number of shards: 4
? Enter the output file name: metis.json

Configuration file 'metis.json' created successfully.
```

### Generated Output (`metis.json`)

The session above would generate the following `metis.json` file in your project's root directory:

```json
{
  "cache_size": 1000,
  "eviction_policy": "wtinylfu",
  "ttl": "10m0s",
  "compress": true,
  "shards": 4
}
```

---

## Debug & Analysis Tool (`metis-debug`)

The Metis Debug CLI (`metis-debug`) is a developer-focused tool designed to help you analyze cache behavior, debug performance issues, and validate your cache configuration during development.

This tool provides real-time insights into cache performance metrics, memory usage, and configuration validation without requiring modifications to your application code.

### Location

The debug tool is located in the `cmd/metis-debug` directory of the project.

### How to Run

You can run the debug CLI directly using `go run`:

```bash
go run ./cmd/metis-debug/main.go [command] [flags]
```

### Available Commands

#### 1. `inspect` - Analyze Cache Performance

Provides comprehensive analysis of cache performance and memory usage.

```bash
# Basic inspection (estimated performance)
go run ./cmd/metis-debug/main.go inspect

# Real cache measurements
go run ./cmd/metis-debug/main.go inspect -real

# JSON output for automation
go run ./cmd/metis-debug/main.go inspect -json

# Verbose output with detailed metrics
go run ./cmd/metis-debug/main.go inspect -v
```

**Sample Output (Estimated Mode):**
```
=== Cache Performance Analysis ===
Health Check: ✓ PASSED

Runtime Information:
- Go Version: go1.24.5
- Architecture: amd64
- OS: linux
- CPUs: 8

Memory Statistics:
- Allocated Memory: 0.2 MB
- Total Allocations: 177960
- Garbage Collections: 0
- Next GC Target: 4.0 MB

Cache Performance Metrics:
- Estimated Operations/sec: 3,500,000
- Average Set Latency: 133 ns
- Average Get Latency: 80 ns
- Estimated Hit Rate: 92.5%
- Cache Type: W-TinyLFU with admission filter
```

**Sample Output (Real Mode with `-real` flag):**
```
=== Cache Performance Analysis ===
Health Check: ✓ PASSED

=== REAL Metis Cache Analysis ===

Cache Configuration:
- Policy: wtinylfu
- Size: 1000 entries
- Shards: 16
- TTL: 5m0s
- Compression: false

Real Performance Measurements:
- Operations/sec: 1,017,327
- Set Latency: 1380 ns
- Get Latency: 585 ns
- Hit Rate: 24.3%
- Cache Utilization: 1000/1000 entries

Runtime Information:
- Go Version: go1.24.5
- Architecture: amd64
- OS: linux
- CPUs: 8

Memory Statistics:
- Allocated Memory: 1.7 MB
- Total Allocations: 1805680
- Garbage Collections: 0
- Next GC Target: 4.0 MB
```

#### 2. `version` - Show Version Information

Displays version information and build details.

```bash
go run ./cmd/metis-debug/main.go version
```

**Output:**
```
metis-debug version 1.0.0, Go version: go1.24.5
```

#### 3. `help` - Show Available Commands

Shows usage information and available commands.

```bash
go run ./cmd/metis-debug/main.go help
```

**Output:**
```
🔧 Metis Debug CLI v1.0.0

USAGE: metis-debug <command> [flags]
COMMANDS:
  inspect     Show cache statistics and performance analysis
  version     Show version information
  help        Show this help

INSPECT FLAGS:
  -json       Output in JSON format
  -v          Enable verbose output
  -real       Use real Metis cache measurements (default: estimated)
```

### Command Flags

- `-json`: Output results in JSON format for automation and integration
- `-v`: Enable verbose output with additional metrics and details
- `-real`: Use real Metis cache instance instead of estimated performance data

### JSON Output Format

When using the `-json` flag, the tool outputs machine-readable JSON:

**Estimated Mode (default):**
```json
{
  "cache": {
    "estimated_ops_per_sec": 3500000,
    "avg_set_latency_ns": 133,
    "avg_get_latency_ns": 80,
    "hit_rate_percent": 92.5,
    "type": "W-TinyLFU (Estimated)"
  },
  "memory": {
    "alloc_mb": 0.167,
    "total_alloc": 175360,
    "num_gc": 0,
    "next_gc_mb": 4.0
  },
  "runtime": {
    "go_version": "go1.24.5",
    "arch": "amd64",
    "os": "linux",
    "num_cpu": 8
  },
  "health": {
    "status": "passed",
    "timestamp": "2025-08-03T14:30:15Z"
  }
}
```

**Real Mode (with `-real` flag):**
```json
{
  "cache": {
    "type": "W-TinyLFU (Real)",
    "real_ops_per_sec": 1017327,
    "real_set_latency_ns": 1380,
    "real_get_latency_ns": 585,
    "cache_size": 1000,
    "hit_rate_percent": 24.3,
    "total_operations": 20000,
    "eviction_policy": "wtinylfu"
  },
  "memory": {
    "alloc_mb": 1.7,
    "total_alloc": 1805680,
    "num_gc": 0,
    "next_gc_mb": 4.0
  },
  "runtime": {
    "go_version": "go1.24.5",
    "arch": "amd64",
    "os": "linux",
    "num_cpu": 8
  },
  "config": {
    "cache_size": 1000,
    "eviction_policy": "wtinylfu",
    "shard_count": 16,
    "enable_compression": false,
    "ttl_minutes": 5
  }
}
```

### Use Cases

#### Estimated Mode (Default)
- **Performance Baselines**: Compare against theoretical performance targets
- **CI/CD Integration**: Fast validation without creating actual cache instances
- **Documentation**: Consistent output for guides and benchmarks
- **Quick Reference**: Immediate performance estimates without overhead

#### Real Mode (`-real` flag)
- **Performance Debugging**: Identify actual bottlenecks in cache operations
- **Memory Analysis**: Monitor real memory usage and garbage collection behavior
- **Configuration Tuning**: Test different cache configurations with real workloads
- **Regression Testing**: Validate actual performance changes between versions
- **Development**: Real-time analysis during cache integration development

---

## Next Steps

### For Configuration Generation
Once you have your `metis.json` file, the Metis library will automatically detect and use it when you create a new cache instance with `metis.New()`, as long as the file is in the same directory where the application is run.

### For Performance Analysis
Use the `metis-debug` tool during development to monitor cache behavior and validate performance:

```bash
# Quick estimated analysis
go run ./cmd/metis-debug/main.go inspect

# Real performance measurement
go run ./cmd/metis-debug/main.go inspect -real

# JSON output for CI/CD integration
go run ./cmd/metis-debug/main.go inspect -json
```

The CLI provides both estimated baselines (fast, consistent) and real measurements (accurate, overhead-inclusive) to support different development and testing needs.

See the [Configuration Guide](./CONFIGURATION.md) for more details on how Metis loads configuration, and the [Benchmarking Guide](./BENCHMARKING.md) for advanced performance analysis techniques.

---

Metis • an AGILira fragment