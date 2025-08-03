# Metis Configuration Templates

Quick-start templates for common use cases.

## 🚀 **Ultra-Simple (Development)**
```json
{
  "cache_size": 1000
}
```

## 🏃 **Web Server (Production)**
```json
{
  "cache_size": 100000,
  "ttl": "30m",
  "eviction_policy": "wtinylfu",
  "shard_count": 64
}
```

## 🔥 **High-Performance (API Gateway)**
```json
{
  "cache_size": 1000000,
  "ttl": "0s",
  "eviction_policy": "wtinylfu",
  "shard_count": 128,
  "enable_compression": false
}
```

## 💾 **Memory-Efficient**
```json
{
  "cache_size": 50000,
  "ttl": "15m",
  "enable_compression": true,
  "max_value_size": 1048576
}
```

## ⚡ **Session Cache (Simple & Predictable)**
```json
{
  "cache_size": 10000,
  "ttl": "1h",
  "cleanup_interval": "10m",
  "eviction_policy": "lru"
}
```

## 🎯 **Debugging/Development (LRU)**
```json
{
  "cache_size": 1000,
  "eviction_policy": "lru",
  "ttl": "10m"
}
```

**Note**: We recommend **W-TinyLFU** (default) for production workloads and **LRU** for development/debugging.

Copy any template to your project as `metis.json` and you're ready to go!
