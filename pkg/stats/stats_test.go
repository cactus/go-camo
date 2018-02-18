// Copyright (c) 2012-2018 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package stats

import (
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcurrentUpdate(t *testing.T) {
	t.Parallel()
	ps := &ProxyStats{}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for v := 0; v < 100000; v++ {
				ps.AddServed()
				ps.AddBytes(1024)
				runtime.Gosched()
			}
		}()
	}

	wg.Wait()
	c, b := ps.GetStats()
	assert.Equal(t, 10000000, int(c), "unexpected client count")
	assert.Equal(t, 10240000000, int(b), "unexpected bytes count")
}
