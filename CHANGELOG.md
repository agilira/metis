# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- [Future features will be listed here]

### Changed
- [Future changes will be listed here]

### Fixed
- [Future fixes will be listed here]

### Security
- [Future security updates will be listed here]

## [0.1.0] - 2025-08-03

### Added
- High-performance sharded caching library with 95% test coverage
- WTinyLFU (Windowed TinyLFU) eviction policy for optimal hit rates
- LRU (Least Recently Used) eviction policy for general-purpose caching
- Multi-level configuration system: programmatic, JSON file, and environment variables
- Intelligent configuration validator with performance optimization suggestions
- Built-in Gzip compression for memory optimization
- Comprehensive observability with detailed statistics (hits, misses, hit rate)
- CLI tools for debugging and profiling cache performance
- Robust serialization using Go's `gob` package
- Use-case specific constructors (WebServer, APIGateway, Development)
- Extensive benchmarking and performance testing suite

### Changed
- Enhanced error handling and edge case management
- Improved memory efficiency and garbage collection patterns
- Optimized sharding algorithm for better concurrency

### Fixed
- Resolved CI/CD pipeline issues and test reliability
- Fixed compression header handling in edge cases
- Improved configuration validation logic for varying CPU core counts
- Enhanced security scanning and dependency management

### Security
- Implemented comprehensive security scanning with gosec
- Added input validation and sanitization for all external inputs
- Enhanced CLI argument validation to prevent command injection

---

Metis • an AGILira fragment
