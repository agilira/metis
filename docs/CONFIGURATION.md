# Advanced Configuration

Metis provides a flexible and layered configuration system. This document details all available options and the precedence rules that govern how they are applied.

## Configuration Precedence

Metis loads its configuration from multiple sources. The sources are prioritized as follows, with the first one found taking precedence:

1.  **Programmatic Configuration**: A `CacheConfig` struct passed directly to `metis.NewWithConfig()`. This offers the highest level of control and overrides all other sources.
2.  **JSON File**: A `metis.json` file located in the application's root directory. This is automatically loaded by `metis.New()`.
3.  **Default Values**: If no other configuration is provided, Metis falls back to a set of sensible default values suitable for general-purpose use.

## `CacheConfig` Parameters

The `CacheConfig` struct is the primary way to configure the cache programmatically.

| Parameter           | Type          | Description                                                                                                | Default      |
| ------------------- | ------------- | ---------------------------------------------------------------------------------------------------------- | ------------ |
| `CacheSize`         | `int`         | The maximum number of items the cache can hold.                                                            | `1000`       |
| `ShardCount`        | `int`         | The number of shards to distribute the cache across. A power of 2 is recommended for optimal performance.  | `16`         |
| `EvictionPolicy`    | `string`      | The eviction policy to use. Supported values: `"wtinylfu"`, `"lru"`.                                       | `"wtinylfu"` |
| `TTL`               | `time.Duration` | The default time-to-live for cache items. A zero value disables expiration.                                | `0` (none)   |
| `EnableCompression` | `bool`        | If `true`, cache values are compressed using Gzip to save memory.                                          | `false`      |
| `MaxValueSize`      | `int`         | The maximum size (in bytes) of a value before it is rejected. Helps prevent large items from polluting the cache. | `0` (none)   |
| `AdmissionPolicy`   | `string`      | The admission policy to use. Currently supports `"always"`.                                                | `"always"`   |
| `Logger`            | `Logger`      | An optional logger interface for debugging and monitoring.                                                 | `nil`        |

### Example: Programmatic Configuration

```go
import "time"

config := metis.CacheConfig{
    CacheSize:         100000,
    ShardCount:        32,
    EvictionPolicy:    "wtinylfu",
    TTL:               15 * time.Minute,
    EnableCompression: true,
}

cache := metis.NewWithConfig(config)
```

## JSON Configuration

The `metis.json` file maps directly to the `CacheConfig` struct fields. Note that field names use `snake_case`.

### Example `metis.json`

```json
{
  "cache_size": 100000,
  "shard_count": 32,
  "eviction_policy": "wtinylfu",
  "ttl": "15m",
  "enable_compression": true,
  "max_value_size": 2048
}
```

## Configuration Validator

Metis includes a smart validator (`ValidateConfig`) that analyzes a given `CacheConfig` and provides warnings and suggestions. This is used internally but can also be called directly for debugging.

### Example Usage

```go
config := metis.CacheConfig{CacheSize: 10} // Intentionally small
validationResult := metis.ValidateConfig(config)

if !validationResult.IsValid {
    fmt.Println("Configuration is invalid.")
}

for _, warning := range validationResult.Warnings {
    fmt.Println("Warning:", warning)
}

for _, suggestion := range validationResult.Suggestions {
    fmt.Println("Suggestion:", suggestion)
}
```
