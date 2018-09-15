// Copyright (c) 2012-2018 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package stats

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

// ProxyStats is the counter container
type ProxyStats struct {
	clients uint64
	bytes   uint64
}

// AddServed increments the number of clients served counter
func (ps *ProxyStats) AddServed() {
	atomic.AddUint64(&ps.clients, 1)
}

// AddBytes increments the number of bytes served counter
func (ps *ProxyStats) AddBytes(bc int64) {
	if bc <= 0 {
		return
	}
	atomic.AddUint64(&ps.bytes, uint64(bc))
}

// GetStats returns the stats: clients, bytes
func (ps *ProxyStats) GetStats() (uint64, uint64) {
	psClients := atomic.LoadUint64(&ps.clients)
	psBytes := atomic.LoadUint64(&ps.bytes)
	return psClients, psBytes
}

// Handler returns an http.HandlerFunc that returns running totals and
// stats about the server.
func Handler(ps *ProxyStats) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, b := ps.GetStats()
		if r.URL.Query().Get("format") == "json" {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			fmt.Fprintf(w, "{\"ClientsServed\": %d, \"BytesServed\": %d}\n", c, b)
		} else {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			fmt.Fprintf(w, "ClientsServed, BytesServed\n%d, %d\n", c, b)
		}
	}
}
