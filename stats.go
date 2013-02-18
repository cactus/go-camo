package main

import (
	"sync/atomic"
)

type ProxyStats struct {
	clientsServed uint64
	bytesServed   uint64
}

func (ps *ProxyStats) AddServed() {
	ps.clientsServed += 1
	atomic.AddUint64(&ps.clientsServed, 1)
}

func (ps *ProxyStats) AddBytes(bc int64) {
	if bc <= 0 {
		return
	}
	atomic.AddUint64(&ps.bytesServed, uint64(bc))
}

func (ps *ProxyStats) GetStats() (b uint64, c uint64) {
	b = atomic.LoadUint64(&ps.clientsServed)
	c = atomic.LoadUint64(&ps.bytesServed)
	return b, c
}
