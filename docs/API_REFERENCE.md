# API Reference

This document provides a practical reference for the public API of the Metis caching library. For more detailed configuration options, see [Advanced Configuration](./CONFIGURATION.md).

## Constructors

### `metis.New()`

Creates a new cache instance by automatically loading the configuration from the environment.

- **Signature**: `func New() *Cache`
- **Details**: It searches for a `metis.json` file in the root directory. If not found, it falls back to default settings. This is the simplest way to get started.

**Example:**
```go
// Automatically loads metis.json or uses defaults
cache := metis.New()
defer cache.Close()
```

### `metis.NewWithConfig()`

Creates a new cache instance with a specific, programmatically defined configuration.

- **Signature**: `func NewWithConfig(config CacheConfig) *Cache`
- **Details**: This function gives you full control over the cache's behavior. The provided `CacheConfig` struct overrides all other configuration sources.

**Example:**
```go
import "time"

config := metis.CacheConfig{
    CacheSize:      10000,
    EvictionPolicy: "lru",
    TTL:            5 * time.Minute,
}
cache := metis.NewWithConfig(config)
defer cache.Close()
```

## Cache Methods

All methods on the `*Cache` object are thread-safe.

### `Set()`

Adds or updates an item in the cache.

- **Signature**: `func (c *Cache) Set(key string, value interface{})`
- **Details**: The `value` can be of any type. Metis uses `gob` internally to serialize the data. For custom types, you may need to register them using `gob.Register()`.

**Example:**
```go
type User struct {
    ID   int
    Name string
}
gob.Register(User{}) // Register custom type

cache.Set("user:1", User{ID: 1, Name: "Alice"})
cache.Set("session:token", "xyz-123")
```

### `Get()`

Retrieves an item from the cache.

- **Signature**: `func (c *Cache) Get(key string) (interface{}, bool)`
- **Returns**:
    - `value (interface{})`: The cached item. It will be `nil` if the key is not found.
    - `found (bool)`: `true` if the key exists, `false` otherwise.
- **Details**: Always check the `found` boolean, as a `nil` value could be a legitimate cached value. You must perform a type assertion to convert the `interface{}` back to its original type.

**Example:**
```go
value, found := cache.Get("user:1")
if found {
    user := value.(User)
    fmt.Printf("User found: %s\n", user.Name)
} else {
    fmt.Println("User not found in cache.")
}
```

### `Delete()`

Removes an item from the cache.

- **Signature**: `func (c *Cache) Delete(key string)`

**Example:**
```go
cache.Delete("session:token")
```

### `Clear()`

Removes all items from the cache across all shards.

- **Signature**: `func (c *Cache) Clear()`

### `Stats()`

Returns statistics about the cache's performance.

- **Signature**: `func (c *Cache) Stats() Stats`
- **Returns**: A `Stats` struct containing `Hits`, `Misses`, `Size`, and `HitRate`.

**Example:**
```go
stats := cache.Stats()
fmt.Printf("Hit Rate: %.2f%%\n", stats.HitRate)
fmt.Printf("Items in cache: %d\n", stats.Size)
```

### `Close()`

Releases any resources used by the cache, such as background cleanup goroutines.

- **Signature**: `func (c *Cache) Close()`
- **Details**: It is crucial to call this method when the cache is no longer needed, typically using `defer`.

**Example:**
```go
cache := metis.New()
defer cache.Close()
// ... use the cache
```
