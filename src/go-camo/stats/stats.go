// Copyright (c) 2012-2016 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package stats

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type ProxyStats struct {
	clients uint64
	bytes   uint64
}

func (ps *ProxyStats) AddServed() {
	atomic.AddUint64(&ps.clients, 1)
}

func (ps *ProxyStats) AddBytes(bc int64) {
	if bc <= 0 {
		return
	}
	atomic.AddUint64(&ps.bytes, uint64(bc))
}

func (ps *ProxyStats) GetStats() (uint64, uint64) {
	psClients := atomic.LoadUint64(&ps.clients)
	psBytes := atomic.LoadUint64(&ps.bytes)
	return psClients, psBytes
}

// StatsHandler returns an http.HandlerFunc that returns running totals and
// stats about the server.
func StatsHandler(ps *ProxyStats) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		c, b := ps.GetStats()
		fmt.Fprintf(w, "ClientsServed, BytesServed\n%d, %d\n", c, b)
	}
}
