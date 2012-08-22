package main

import (
	"sync"
)

type ProxyStats struct {
	sync.Mutex
	clientsServed uint64
	bytesServed   uint64
	Enable        bool
}

func (ps *ProxyStats) AddServed() {
	ps.Lock()
	defer ps.Unlock()
	ps.clientsServed += 1
}

func (ps *ProxyStats) AddBytes(bc int64) {
	ps.Lock()
	defer ps.Unlock()
	if bc <= 0 {
		return
	}
	ps.bytesServed += uint64(bc)
}

func (ps *ProxyStats) GetStats() (b uint64, c uint64) {
	return ps.clientsServed, ps.bytesServed
}
