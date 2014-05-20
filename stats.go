package main

import (
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
	return ps.bytes, ps.clients
}
