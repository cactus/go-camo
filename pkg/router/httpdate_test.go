// Copyright (c) 2012-2023 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package router

import (
	"bytes"
	"flag"
	"os"
	"testing"
	"time"

	"github.com/cactus/mlog"
	"github.com/dropwhile/assert"
)

var logBuffer = &bytes.Buffer{}

func TestHTTPDateGoroutineUpdate(t *testing.T) {
	t.Parallel()
	d := newiHTTPDate()
	n := d.String()
	time.Sleep(2 * time.Second)
	l := d.String()
	assert.NotEqual(t, n, l, "Date did not update as expected: %s == %s", n, l)
}

func TestHTTPDateManualUpdate(t *testing.T) {
	t.Parallel()
	d := &iHTTPDate{}
	d.Update()
	n := d.String()
	time.Sleep(2 * time.Second)
	d.Update()
	l := d.String()
	assert.NotEqual(t, n, l, "Date did not update as expected: %s == %s", n, l)
}

func TestHTTPDateManualUpdateUninitialized(t *testing.T) {
	t.Parallel()
	d := &iHTTPDate{}

	n := d.String()
	time.Sleep(2 * time.Second)
	d.Update()
	l := d.String()
	assert.NotEqual(t, n, l, "Date did not update as expected: %s == %s", n, l)
}

func BenchmarkDataString(b *testing.B) {
	d := newiHTTPDate()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = d.String()
		}
	})
}

func TestMain(m *testing.M) {
	flag.Parse()

	debug := os.Getenv("DEBUG")
	// now configure a standard logger
	mlog.SetFlags(mlog.Lstd)

	if debug != "" {
		mlog.SetFlags(mlog.Flags() | mlog.Ldebug)
		mlog.Debug("debug logging enabled")
	}

	mlog.DefaultLogger = mlog.New(logBuffer, mlog.Lstd)

	os.Exit(m.Run())
}
