package camoproxy_test

import (
	"sync"
	"github.com/cactus/go-camo/camoproxy"
)

type ProxyStats struct {
	sync.Mutex
	clientsServed uint64
	bytesServed   uint64
}

func (ps *ProxyStats) AddServed() {
	ps.Lock()
	defer ps.Unlock()
	ps.clientsServed += 1
}

func (ps *ProxyStats) AddBytes(bc int64) {
	ps.Lock()
	defer ps.Unlock()
	ps.bytesServed += uint64(bc)
}

func ExampleProxyMetrics() {
	proxy := &camoproxy.Proxy{}
	config := camoproxy.Config{}
	ps := &ProxyStats{}
	proxy.SetMetricsCollector(ps)
}
