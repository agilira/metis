# Makefile for Metis Profiler
# Copyright (c) 2025 AGILira

.PHONY: help build clean profiler test-profile test-core coverage-core

# Default target
help:
	@echo "Metis Profiler - Available targets:"
	@echo "  build        - Build the profiler executable"
	@echo "  profiler     - Build and run a quick test"
	@echo "  test-profile - Run profiling with all profiles enabled"
	@echo "  test-core    - Run tests for core library only"
	@echo "  coverage-core- Generate coverage report for core library"
	@echo "  clean        - Remove generated files"
	@echo "  help         - Show this help message"

# Build the profiler
build:
	@echo "Building profiler..."
	go build -o profiler.exe ./cmd/profiler
	@echo "Profiler built successfully: profiler.exe"

# Quick test run
profiler: build
	@echo "Running quick profiler test..."
	./profiler.exe -duration 5s -workers 4 -ops 500 -keyspace 2000 -valuesize 512

# Full profiling test
test-profile: build
	@echo "Running full profiling test..."
	./profiler.exe \
		-duration 30s \
		-workers 8 \
		-ops 1000 \
		-keyspace 10000 \
		-valuesize 1024 \
		-cpuprofile cpu.prof \
		-memprofile memory.prof \
		-blockprofile block.prof \
		-mutexprofile mutex.prof \
		-eviction wtinylfu \
		-admission probabilistic \
		-shards 16 \
		-compression

# Clean generated files
clean:
	@echo "Cleaning generated files..."
	@if exist profiler.exe del profiler.exe
	@if exist *.prof del *.prof
	@if exist *.prof del *.prof
	@echo "Clean completed"

# Stress test
stress: build
	@echo "Running stress test..."
	./profiler.exe \
		-duration 60s \
		-workers 16 \
		-ops 2000 \
		-keyspace 50000 \
		-valuesize 2048 \
		-cpuprofile stress_cpu.prof \
		-memprofile stress_memory.prof

# Light load test
light: build
	@echo "Running light load test..."
	./profiler.exe \
		-duration 10s \
		-workers 2 \
		-ops 100 \
		-keyspace 1000 \
		-valuesize 256

# Medium load test
medium: build
	@echo "Running medium load test..."
	./profiler.exe \
		-duration 30s \
		-workers 8 \
		-ops 1000 \
		-keyspace 10000 \
		-valuesize 1024

# Test core library only (excludes CLI tools)
test-core:
	@echo "Running tests for core library only..."
	go test -v .

# Generate coverage report for core library only  
coverage-core:
	@echo "Generating coverage report for core library..."
	go test -cover -coverprofile=coverage-core.out .
	go tool cover -html=coverage-core.out -o coverage-core.html
	@echo "Core library coverage: $$(go tool cover -func=coverage-core.out | grep total | awk '{print $$3}')"
	@echo "Coverage report saved to coverage-core.html"

# Clean generated files  
clean:
	rm -f profiler.exe profiler coverage.out coverage.html coverage-core.out coverage-core.html
	@echo "Cleaned generated files"
