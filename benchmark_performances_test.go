// benchmark_performances_test.go: Performance benchmarks for Metis
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package metis

import (
	"testing"
)

// BenchmarkPerformances tests Metis performance with optimized configuration
func BenchmarkPerformances(b *testing.B) {
	const keySpaceSize = 1000

	b.Run("Metis_Optimized_Set", func(b *testing.B) {
		cache := NewWTinyLFU(10000, 64)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			key := string(rune(i % keySpaceSize))
			value := i
			cache.Set(key, value)
		}
	})

	b.Run("Metis_Optimized_Get", func(b *testing.B) {
		cache := NewWTinyLFU(10000, 64)

		// Pre-populate cache
		for i := 0; i < keySpaceSize; i++ {
			key := string(rune(i))
			value := i
			cache.Set(key, value)
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			key := string(rune(i % keySpaceSize))
			_, _ = cache.Get(key)
		}
	})

	b.Run("Metis_Optimized_Concurrent", func(b *testing.B) {
		cache := NewWTinyLFU(10000, 64)

		b.ResetTimer()
		b.ReportAllocs()

		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := string(rune(i % keySpaceSize))
				value := i
				cache.Set(key, value)
				_, _ = cache.Get(key)
				i++
			}
		})
	})
}
