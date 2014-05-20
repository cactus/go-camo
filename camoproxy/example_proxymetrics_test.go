package camoproxy_test

import (
	"log"
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
	config := camoproxy.Config{}
	proxy, err := camoproxy.New(config)
	if err != nil {
		log.Fatal("Error: ", err)
	}
	ps := &ProxyStats{}
	proxy.SetMetricsCollector(ps)
}
