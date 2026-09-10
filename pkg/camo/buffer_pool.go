// Copyright (c) 2012-2023 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package camo

import (
	"sync"
)

const bufSize = 32 * 1024

var bufPool = sync.Pool{
	New: func() any {
		// note: 32 * 1024 is the size used by io.Copy by default.
		// Seems like a good starting point, just with a bit less garbage
		// (using a sync pool) to reduce some GC work.
		// ref: https://golang.org/src/io/io.go?s=13136:13214#L391
		buf := make([]byte, bufSize)
		return &buf
	},
}
