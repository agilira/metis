// example_metis_config.go: Example Go configuration for power users
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"time"

	"github.com/agilira/metis"
)

func init() {
	// Power user configuration - this overrides metis.json if present
	config := metis.CacheConfig{
		EnableCaching:        true,
		CacheSize:            50000,
		TTL:                  15 * time.Minute,
		CleanupInterval:      2 * time.Minute,
		EnableCompression:    true,
		EvictionPolicy:       "wtinylfu",
		ShardCount:           64,
		AdmissionPolicy:      "probabilistic",
		AdmissionProbability: 0.5,
		MaxKeySize:           2048,
		MaxValueSize:         2097152, // 2MB
	}

	metis.SetGlobalConfig(config)
}
