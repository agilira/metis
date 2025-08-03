// pool.go: Pool implementation for Metis strategic caching library
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package metis

import (
	"bytes"
	"sync"
)

// bufferPool provides pooled *bytes.Buffer instances for serialization/deserialization.
var bufferPool = sync.Pool{
	New: func() interface{} { return new(bytes.Buffer) },
}

// getBuffer retrieves a *bytes.Buffer from the pool.
func getBuffer() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

// putBuffer resets and returns a *bytes.Buffer to the pool.
func putBuffer(buf *bytes.Buffer) {
	buf.Reset()
	bufferPool.Put(buf)
}
