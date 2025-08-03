// utils.go: Utility functions for Metis strategic caching library
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package metis

import (
	"encoding/gob"
	"reflect"
	"strconv"
)

// calculateSize estimates the memory size of a value in bytes
func calculateSize(value interface{}) int {
	if value == nil {
		return 0
	}

	switch v := value.(type) {
	case string:
		return len(v)
	case []byte:
		return len(v)
	case int:
		// Fast path for common integer types
		return len(strconv.AppendInt(nil, int64(v), 10))
	case int32:
		return len(strconv.AppendInt(nil, int64(v), 10))
	case int64:
		return len(strconv.AppendInt(nil, v, 10))
	case uint:
		return len(strconv.AppendUint(nil, uint64(v), 10))
	case uint32:
		return len(strconv.AppendUint(nil, uint64(v), 10))
	case uint64:
		return len(strconv.AppendUint(nil, v, 10))
	case float32:
		return len(strconv.AppendFloat(nil, float64(v), 'g', -1, 32))
	case float64:
		return len(strconv.AppendFloat(nil, v, 'g', -1, 64))
	case bool:
		if v {
			return 4 // "true"
		}
		return 5 // "false"
	case PrimitiveBox:
		return calculateSize(v.V)
	default:
		if v := reflect.ValueOf(value); v.Kind() == reflect.Ptr && v.IsNil() {
			return 8 // pointer size
		}
		// Fallback to gob encoding for complex types
		buf := getBuffer()
		defer putBuffer(buf)
		enc := gob.NewEncoder(buf)
		if err := enc.Encode(value); err != nil {
			return 0
		}
		return buf.Len()
	}
}

// toBytes converts a value to []byte for compression
func toBytes(value interface{}) ([]byte, error) {
	if value == nil {
		return []byte{}, nil
	}

	switch v := value.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	case int:
		// Fast path: convert directly to bytes without string conversion
		return strconv.AppendInt(nil, int64(v), 10), nil
	case int32:
		return strconv.AppendInt(nil, int64(v), 10), nil
	case int64:
		return strconv.AppendInt(nil, v, 10), nil
	case uint:
		return strconv.AppendUint(nil, uint64(v), 10), nil
	case uint32:
		return strconv.AppendUint(nil, uint64(v), 10), nil
	case uint64:
		return strconv.AppendUint(nil, v, 10), nil
	case float32:
		return strconv.AppendFloat(nil, float64(v), 'g', -1, 32), nil
	case float64:
		return strconv.AppendFloat(nil, v, 'g', -1, 64), nil
	case bool:
		if v {
			return []byte("true"), nil
		}
		return []byte("false"), nil
	case PrimitiveBox:
		return toBytes(v.V)
	default:
		// Fallback to gob encoding for complex types
		box := PrimitiveBox{V: value}
		buf := getBuffer()
		defer putBuffer(buf)
		enc := gob.NewEncoder(buf)
		if err := enc.Encode(box); err != nil {
			return nil, err
		}
		// Make a copy of the bytes to avoid buffer reuse issues
		result := make([]byte, buf.Len())
		copy(result, buf.Bytes())
		return result, nil
	}
}

// parsePrimitiveFromString attempts to parse a string back to its original primitive type
func parsePrimitiveFromString(s string) (interface{}, bool) {
	// Try to parse as boolean first
	if s == "true" {
		return true, true
	}
	if s == "false" {
		return false, true
	}

	// Try to parse as integer
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		// Return as int64 to preserve the original type
		return i, true
	}

	// Try to parse as unsigned integer
	if u, err := strconv.ParseUint(s, 10, 64); err == nil {
		// Return as uint64 to preserve the original type
		return u, true
	}

	// Try to parse as float
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		// Return as float64 to preserve the original type
		return f, true
	}

	// If all parsing fails, return the original string
	return s, true
}
