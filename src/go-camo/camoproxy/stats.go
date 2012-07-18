package camoproxy

import (
	"sync"
)

type proxyStats struct {
	sync.Mutex
	clientsServed uint64
	bytesServed   uint64
	Enable        bool
}

func (ps *proxyStats) AddServed() {
	ps.Lock()
	defer ps.Unlock()
	ps.clientsServed += 1
}

func (ps *proxyStats) AddBytes(bc int64) {
	ps.Lock()
	defer ps.Unlock()
	if bc <= 0 {
		return
	}
	ps.bytesServed += uint64(bc)
}

func (ps *proxyStats) GetStats() (b uint64, c uint64) {
	return ps.clientsServed, ps.bytesServed
}
