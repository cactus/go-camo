package stats

import (
	"fmt"
	"net/http"
	"sync"
)

type ProxyStats struct {
	sync.RWMutex
	clients uint64
	bytes   uint64
}

func (ps *ProxyStats) AddServed() {
	ps.Lock()
	ps.clients++
	ps.Unlock()
}

func (ps *ProxyStats) AddBytes(bc int64) {
	if bc <= 0 {
		return
	}
	ps.Lock()
	ps.bytes += uint64(bc)
	ps.Unlock()
}

func (ps *ProxyStats) GetStats() (uint64, uint64) {
	ps.RLock()
	defer ps.RUnlock()
	return ps.clients, ps.bytes
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
