// Copyright (c) 2012-2016 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
//
// some parts from: https://brandur.org/t-parallel

package mlog

import (
	"fmt"
	"testing"
)

// TestoingLogWriter is an adapter between mlog and Go's testing package,
// which lets us send all output to `t.Log` so that it's correctly
// collated with the test that emitted it. This helps especially when
// using parallel testing where output would otherwise be interleaved
// and make debugging extremely difficult.
type TestingLogWriter struct {
	tb testing.TB
}

func (lw *TestingLogWriter) Write(p []byte) (n int, err error) {
	// Unfortunately, even with this call to `t.Helper()` there's no
	// way to correctly attribute the log location to where it's
	// actually emitted in our code (everything shows up under
	// `logger.go`). A good explanation of this problem and possible
	// future solutions here:
	//
	// https://github.com/neilotoole/slogt#deficiency
	if lw == nil {
		return 0, nil
	}
	if lw.tb == nil {
		fmt.Println("got nil testing.TB")
		return 0, fmt.Errorf("got a nil testing.TBf")
	}
	lw.tb.Helper()
	lw.tb.Log(string(p))
	return len(p), nil
}
