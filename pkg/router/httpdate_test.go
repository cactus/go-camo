// Copyright (c) 2012-2019 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package router

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestHTTPDateGoroutineUpdate(t *testing.T) {
	t.Parallel()
	d := newiHTTPDate()
	n := d.String()
	time.Sleep(2 * time.Second)
	l := d.String()
	assert.Check(t, n != l, "Date did not update as expected: %s == %s", n, l)
}

func TestHTTPDateManualUpdate(t *testing.T) {
	t.Parallel()
	d := &iHTTPDate{}
	d.Update()
	n := d.String()
	time.Sleep(2 * time.Second)
	d.Update()
	l := d.String()
	assert.Check(t, n != l, "Date did not update as expected: %s == %s", n, l)
}

func TestHTTPDateManualUpdateUninitialized(t *testing.T) {
	t.Parallel()
	d := &iHTTPDate{}

	n := d.String()
	time.Sleep(2 * time.Second)
	d.Update()
	l := d.String()
	assert.Check(t, n != l, "Date did not update as expected: %s == %s", n, l)
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
