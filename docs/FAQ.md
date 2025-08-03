# Frequently Asked Questions (FAQ)

### Q: What is the default eviction policy?

**A:** The default eviction policy is `WTinyLFU` (Windowed TinyLFU). It generally provides superior hit ratios compared to traditional `LRU` for most real-world workloads, as it is resistant to cache pollution from infrequent, one-hit-wonder items.

### Q: Do I need to configure the number of shards?

**A:** Not necessarily. If you don't specify the `Shards` parameter, Metis will default to the next power of two of the number of CPU cores detected (`runtime.NumCPU()`). This is a sensible default for many applications. You should only need to tune this if you have a very high-concurrency workload and have identified lock contention as a bottleneck through profiling.

### Q: Why is my custom `struct` causing a `panic: gob: type not registered`?

**A:** This happens when you enable a feature that requires serialization (currently, `Compress: true`) and try to store a custom struct that hasn't been registered with Go's `gob` package.

**Solution:** You must register your type in an `init()` function in your application. See the [Custom Serialization Tutorial](./TUTORIALS/04-custom-serialization.md) for a detailed guide.

```go
import "encoding/gob"

type MyStruct struct {
    // ...
}

func init() {
    gob.Register(MyStruct{})
}
```

### Q: What is the difference between `WTinyLFU` and `LRU`?

**A:**
*   **LRU (Least Recently Used):** Evicts the item that hasn't been accessed for the longest time. It's simple and effective for workloads where recently accessed items are very likely to be accessed again soon. However, it can be "polluted" by scans or infrequent items that fill the cache and evict valuable, frequently-used items.
*   **WTinyLFU (Windowed TinyLFU):** A more advanced, frequency-based policy. It maintains a compact summary of access frequencies to make more intelligent eviction decisions. It protects frequently accessed items from being evicted by a burst of one-time accesses. For most general-purpose caching, `WTinyLFU` will provide a better cache hit ratio.

See the [Eviction Policies Guide](./EVICTION_POLICIES.md) for more details.

### Q: How does Metis handle configuration?

**A:** Metis uses a layered configuration system with the following order of precedence (from highest to lowest):

1.  **Programmatic Configuration:** Settings provided directly in a `metis.CacheConfig` struct when calling `metis.NewWithConfig()`.
2.  **`metis.json` File:** If no programmatic config is provided, Metis looks for a `metis.json` file in the application's working directory.
3.  **Default Values:** If neither of the above is found, Metis uses its built-in default values (e.g., `WTinyLFU`, 1024 shards based on CPU, etc.).

See the [Configuration Guide](./CONFIGURATION.md) for a full breakdown.

### Q: Can I disable the TTL for certain items?

**A:** Yes. When you call `cache.Set(key, value)`, it uses the default TTL set in the cache's configuration. To use a different TTL (or no TTL), use the `cache.SetWithTTL` method.

```go
// Set an item with a custom TTL of 30 seconds
cache.SetWithTTL("short-lived-key", myValue, 30*time.Second)

// Set an item with no expiration (it will only be evicted by cache pressure)
cache.SetWithTTL("permanent-key", myValue, metis.NoExpiration)
```

### Q: Is Metis thread-safe?

**A:** Yes. All public methods on the `metis.Cache` object are thread-safe. The cache is internally sharded, with each shard protected by its own read-write mutex, allowing for a high degree of concurrency. You can safely call `Get`, `Set`, `Delete`, etc., from multiple goroutines.

---

Metis • an AGILira fragment
