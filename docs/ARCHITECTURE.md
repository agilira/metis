# Architecture Overview

This document provides a high-level overview of the internal architecture of Metis. Understanding these components can help you make more informed configuration choices and contribute to the project more effectively.

## High-Level Diagram

```
+-----------------------+
|      Public API       |
| (metis.Cache)         |
+-----------------------+
           |
           v
+-----------------------+
|   Strategic Cache     |
| (Routing & Middleware)|
+-----------------------+
           |
           v
+-----------------------+      +-----------------------+      +-----------------------+
|       Shard 0         |      |       Shard 1         |      |       Shard N         |
|-----------------------|      |-----------------------|      |-----------------------|
| - RWMutex             |      | - RWMutex             |      | - RWMutex             |
| - Data (map)          |      | - Data (map)          |      | - Data (map)          |
| - Eviction Policy     |      | - Eviction Policy     |      | - Eviction Policy     |
+-----------------------+      +-----------------------+      +-----------------------+
```

## Core Components

### 1. Public API (`Cache`)

This is the primary entry point for users of the library. It provides a simplified, user-friendly interface that abstracts away the underlying complexity. It is responsible for exposing methods like `Set`, `Get`, `Delete`, etc.

### 2. Strategic Cache

This layer acts as a central router and middleware. Its main responsibilities are:
- **Configuration Management**: It holds the `CacheConfig` and applies settings like TTL and compression.
- **Sharding Logic**: It determines which shard a key belongs to.
- **Aggregation**: It aggregates stats from all shards to provide a global view of the cache's performance.

### 3. Shards (`cacheShard`)

Metis is a sharded cache, meaning the data is partitioned across multiple independent data structures called shards.

- **Concurrency**: Each shard has its own `sync.RWMutex`, which means that operations on different shards can occur in parallel without lock contention. This is the key to Metis's high concurrency.
- **Key Distribution**: When an operation is performed, the `Strategic Cache` calculates a hash of the key and uses a bitwise mask (`keyHash & shardMask`) to quickly determine the target shard. This ensures an even distribution of keys.
- **Isolation**: Each shard manages its own data map and its own eviction policy instance (e.g., its own LRU list).

## Request Flow

### `Set(key, value)`

1.  The `Public API` receives the request.
2.  The `Strategic Cache` calculates the hash of the `key` to select a shard.
3.  If compression is enabled, the `value` is compressed using Gzip.
4.  The request is forwarded to the target `cacheShard`.
5.  The shard acquires a write lock (`mu.Lock()`).
6.  The key-value pair is stored in the shard's internal `map`.
7.  The eviction policy is updated (e.g., the item is added to the front of an LRU list).
8.  The lock is released.

### `Get(key)`

1.  The `Public API` receives the request.
2.  The `Strategic Cache` calculates the hash of the `key` to select a shard.
3.  The request is forwarded to the target `cacheShard`.
4.  The shard acquires a read lock (`mu.RLock()`), allowing other reads to proceed concurrently.
5.  The value is retrieved from the shard's `map`.
6.  If the value was compressed, it is decompressed.
7.  The eviction policy is updated (e.g., the item is moved to the front of the LRU list).
8.  The lock is released and the value is returned.

## Eviction and Cleanup

- **Eviction**: Eviction is triggered when a shard's `CacheSize` limit is reached during a `Set` operation. The chosen `EvictionPolicy` (e.g., LRU) determines which item is removed.
- **TTL Cleanup**: A background goroutine runs periodically (defined by `CleanupInterval`) to scan for and remove expired items. This is a lazy process to avoid performance overhead on critical paths.

---

Metis • an AGILira fragment
