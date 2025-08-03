# Eviction Policies

Metis provides multiple eviction policies, each designed for different access patterns and performance requirements. Choosing the right policy is key to maximizing cache hit rates and application performance.

## WTinyLFU (Windowed TinyLFU)

**Default Policy**: `wtinylfu`

WTinyLFU is an advanced, frequency-based eviction policy that offers high hit rates for a wide variety of workloads. It is a hybrid approach that combines the strengths of both LFU (Least Frequently Used) and LRU (Least Recently Used) algorithms.

### How It Works

WTinyLFU divides the cache into two main sections:

1.  **Window Cache (LRU)**: A small, primary cache that holds newly added items. This section acts as a "window" of recent entries and follows a simple LRU policy. It gives new items a chance to be accessed before being considered for long-term storage.

2.  **Main Cache (Segmented LRU)**: A larger, secondary cache that stores items that have been accessed at least once while in the window cache. This section is managed by a Segmented LRU (SLRU) policy, which is an approximation of LFU. It protects frequently accessed items from being evicted by a sudden influx of new, infrequently accessed data.

An admission filter, based on a Count-Min Sketch, sits in front of the cache. It probabilistically decides whether a new item is worth admitting into the cache, preventing cache pollution from one-hit wonders.

### Use Cases

- **High-Throughput Systems**: Ideal for API gateways, web servers, and databases where access patterns are complex and varied.
- **Frequency-Based Caching**: Excels when some items are accessed much more frequently than others.
- **General-Purpose Caching**: Its balanced approach makes it a strong default choice for most applications.

## LRU (Least Recently Used)

**Policy Name**: `lru`

LRU is a classic eviction policy that discards the least recently used items first. It maintains a queue of items, and every time an item is accessed, it is moved to the front of the queue. When the cache is full, the item at the back of the queue is evicted.

### How It Works

Metis implements a standard, highly-optimized LRU using a doubly-linked list and a hash map for O(1) time complexity for `Get` and `Set` operations.

### Use Cases

- **Recency-Biased Workloads**: Effective when recently accessed items are very likely to be accessed again soon.
- **Simplicity**: A good choice when the access patterns are simple and a more complex policy like WTinyLFU is not necessary.
- **Debugging**: Its predictable behavior can make it easier to debug caching issues.

## Comparison

| Feature             | WTinyLFU                                       | LRU                                            |
| ------------------- | ---------------------------------------------- | ---------------------------------------------- |
| **Core Principle**  | Frequency and Recency                          | Recency                                        |
| **Complexity**      | Higher                                         | Lower                                          |
| **Memory Overhead** | Slightly higher due to the admission filter    | Lower                                          |
| **Hit Rate**        | Generally higher for varied workloads          | High for recency-focused workloads             |
| **Best For**        | Complex, real-world access patterns            | Simple, sequential, or looping access patterns |

To set the eviction policy, use the `EvictionPolicy` field in your `CacheConfig` or `metis.json` file.

---

Metis • an AGILira fragment
