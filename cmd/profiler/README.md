# Metis Profiler Guide

Benchmark tool for the AGILira Metis Strategic Caching Library.

---

## Configuration Summary

All parameters are defined as constants:

```go
const (
  duration         = 30 * time.Second
  workers          = 16
  keySpaceSize     = 100_000
  valueSize        = 256
  workload         = "balanced" // or: read-heavy, write-heavy
  enableCompress   = true
  evictionPolicy   = "wtinylfu"
  admissionPolicy  = "probabilistic"
  shardCount       = 64
)
```

---

## Execution Flow

1. Warmup phase

   * Populates cache with entries from `key_0` to `key_99999`
2. Benchmark phase

   * Spawns N concurrent goroutines (workers)
   * Each loop randomly selects a key, and performs either Set or Get depending on the workload
   * Latency and operation count are recorded
3. After the benchmark duration ends, results are exported

---

## Output Example

```
--- Results ---
Total operations: 12340001
Set:  avg=104ns min=34ns max=612ns
Get:  avg=49ns  min=18ns max=307ns
Ops/sec: 1.23M
Heap alloc: 85 MB, GCs: 23, GC fraction: 0.32%
```

---

## CSV Export (metis\_results.csv)

Example content:

```
metric,value
total_ops,12345678
set_avg_ns,104
get_avg_ns,49
ops_per_sec,1230000
heap_alloc_mb,85
gc_count,23
gc_fraction,0.32
```

---

## JSON Export (metis\_results.json)

```json
{
  "total_ops": 12345678,
  "set_avg_ns": 104,
  "set_min_ns": 34,
  "set_max_ns": 612,
  "get_avg_ns": 49,
  "get_min_ns": 18,
  "get_max_ns": 307,
  "ops_per_sec": 1230000.00,
  "heap_alloc_mb": 85,
  "gc_count": 23,
  "gc_fraction": 0.32
}
```

---

## Profiling with `pprof`

To enable CPU profiling:

```bash
METIS_PROFILE=true go run main.go
```

Then analyze with:

```bash
go tool pprof -http=:6060 cpu.prof
```

---

## Comparing Results with `benchstat`

```bash
benchstat results_v1.txt results_v2.txt
```

Where `results_vX.txt` are manually saved outputs from different runs of the profiler.

---

## Future Improvements (Optional)

* Support for Prometheus metrics export
* HDR histogram support for latency distribution
* Integration with time series databases (e.g., KairosDB, InfluxDB)

---

## Notes

This profiler is designed to simulate realistic access patterns using configurable read-write distributions. It incorporates warmup phases, latency measurement, concurrent workloads, memory stats, and output in CSV and JSON formats for further analysis.

The goal is to provide a precise yet flexible benchmark harness for evaluating changes to the Metis cache engine.

---

AGILira © 2025
Licensed under MPL-2.0
