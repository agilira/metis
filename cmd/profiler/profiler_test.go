package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agilira/metis"
)

func TestOpStat(t *testing.T) {
	stat := &opStat{}

	// Test recording operations
	stat.Record(time.Millisecond)
	stat.Record(2 * time.Millisecond)
	stat.Record(3 * time.Millisecond)

	if stat.Count != 3 {
		t.Errorf("Expected count 3, got %d", stat.Count)
	}

	if stat.Min != time.Millisecond {
		t.Errorf("Expected min %v, got %v", time.Millisecond, stat.Min)
	}

	if stat.Max != 3*time.Millisecond {
		t.Errorf("Expected max %v, got %v", 3*time.Millisecond, stat.Max)
	}

	expected := 2 * time.Millisecond // (1+2+3)/3 = 2
	if stat.Avg() != expected {
		t.Errorf("Expected average %v, got %v", expected, stat.Avg())
	}

	// Test zero case
	emptyStat := &opStat{}
	if emptyStat.Avg() != 0 {
		t.Errorf("Expected zero average for empty opStat, got %v", emptyStat.Avg())
	}
}

func TestCacheConfig(t *testing.T) {
	// Test that we can create a valid cache config like main() does
	config := metis.CacheConfig{
		EnableCaching:     true,
		CacheSize:         1000,
		TTL:               5 * time.Minute,
		EnableCompression: false,
		EvictionPolicy:    "wtinylfu",
		AdmissionPolicy:   "always",
		ShardCount:        16,
		CleanupInterval:   30 * time.Second,
		MaxKeySize:        256,
		MaxValueSize:      1024,
	}

	// Test that we can create a cache with this config
	cache := metis.NewStrategicCache(config)
	if cache == nil {
		t.Error("Expected non-nil cache")
	}

	// Test basic operations
	success := cache.Set("test", "value")
	if !success {
		t.Error("Expected successful set operation")
	}

	value, found := cache.Get("test")
	if !found {
		t.Error("Expected to find value")
	}

	if value != "value" {
		t.Errorf("Expected 'value', got %s", value)
	}
}

func TestStatisticsCollection(t *testing.T) {
	// Test the statistics collection logic that would be used in main()

	setStat := &opStat{}
	getStat := &opStat{}
	deleteStat := &opStat{}

	// Simulate operations
	setStat.Record(time.Microsecond * 100)
	setStat.Record(time.Microsecond * 150)
	setStat.Record(time.Microsecond * 200)

	getStat.Record(time.Microsecond * 50)
	getStat.Record(time.Microsecond * 75)

	deleteStat.Record(time.Microsecond * 80)

	// Verify statistics
	if setStat.Count != 3 {
		t.Errorf("Expected 3 set operations, got %d", setStat.Count)
	}

	if getStat.Count != 2 {
		t.Errorf("Expected 2 get operations, got %d", getStat.Count)
	}

	if deleteStat.Count != 1 {
		t.Errorf("Expected 1 delete operation, got %d", deleteStat.Count)
	}

	// Test averages
	expectedSetAvg := time.Microsecond * 150 // (100+150+200)/3
	if setStat.Avg() != expectedSetAvg {
		t.Errorf("Expected set average %v, got %v", expectedSetAvg, setStat.Avg())
	}

	expectedGetAvg := time.Microsecond * 62 // (50+75)/2 = 62.5 -> 62
	actualGetAvg := getStat.Avg()
	if actualGetAvg < time.Microsecond*60 || actualGetAvg > time.Microsecond*65 {
		t.Errorf("Expected get average around %v, got %v", expectedGetAvg, actualGetAvg)
	}
}

func TestConfigFileHandling(t *testing.T) {
	// Test config file functionality that would be used in main()
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.json")

	// Create a test config
	config := map[string]interface{}{
		"maxSize":              1000,
		"shards":               4,
		"compressionEnabled":   false,
		"compressionThreshold": 1024,
		"ttl":                  "1h",
		"evictionPolicy":       "w-tinylfu",
	}

	// Write config to file
	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	err = os.WriteFile(configFile, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Read and validate config
	readData, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var readConfig map[string]interface{}
	err = json.Unmarshal(readData, &readConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	// Validate config contents
	if readConfig["maxSize"].(float64) != 1000 {
		t.Errorf("Expected maxSize 1000, got %v", readConfig["maxSize"])
	}

	if readConfig["evictionPolicy"].(string) != "w-tinylfu" {
		t.Errorf("Expected evictionPolicy 'w-tinylfu', got %v", readConfig["evictionPolicy"])
	}
}

func TestWorkloadPatterns(t *testing.T) {
	// Test different workload patterns that the profiler supports
	cache := metis.NewStrategicCache(metis.CacheConfig{
		EnableCaching:     true,
		CacheSize:         100,
		TTL:               time.Minute,
		EnableCompression: false,
		EvictionPolicy:    "wtinylfu",
		ShardCount:        4,
	})

	// Test read-heavy workload simulation
	readStat := &opStat{}
	for i := 0; i < 10; i++ {
		start := time.Now()
		_, _ = cache.Get("nonexistent")
		readStat.Record(time.Since(start))
	}

	if readStat.Count != 10 {
		t.Errorf("Expected 10 read operations, got %d", readStat.Count)
	}

	// Test write-heavy workload simulation
	writeStat := &opStat{}
	for i := 0; i < 5; i++ {
		start := time.Now()
		_ = cache.Set("key"+string(rune('0'+i)), "value")
		writeStat.Record(time.Since(start))
	}

	if writeStat.Count != 5 {
		t.Errorf("Expected 5 write operations, got %d", writeStat.Count)
	}

	// Test balanced workload
	mixedStat := &opStat{}
	for i := 0; i < 6; i++ {
		if i%2 == 0 {
			start := time.Now()
			_ = cache.Set("mixed"+string(rune('0'+i)), "value")
			mixedStat.Record(time.Since(start))
		} else {
			start := time.Now()
			_, _ = cache.Get("mixed" + string(rune('0'+i-1)))
			mixedStat.Record(time.Since(start))
		}
	}

	if mixedStat.Count != 6 {
		t.Errorf("Expected 6 mixed operations, got %d", mixedStat.Count)
	}
}

func TestMainConstants(t *testing.T) {
	// Test that the constants used in main() are reasonable
	if duration <= 0 {
		t.Error("Duration should be positive")
	}

	if workers <= 0 {
		t.Error("Workers should be positive")
	}

	if keySpaceSize <= 0 {
		t.Error("Key space size should be positive")
	}

	if valueSize <= 0 {
		t.Error("Value size should be positive")
	}

	validWorkloads := []string{"read-heavy", "write-heavy", "balanced"}
	found := false
	for _, w := range validWorkloads {
		if workload == w {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Workload '%s' is not valid", workload)
	}

	if shardCount <= 0 {
		t.Error("Shard count should be positive")
	}
}

func BenchmarkOpStatRecord(b *testing.B) {
	stat := &opStat{}
	duration := time.Microsecond

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stat.Record(duration)
	}
}

func BenchmarkOpStatAvg(b *testing.B) {
	stat := &opStat{}
	// Pre-populate with some data
	for i := 0; i < 1000; i++ {
		stat.Record(time.Microsecond)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = stat.Avg()
	}
}
