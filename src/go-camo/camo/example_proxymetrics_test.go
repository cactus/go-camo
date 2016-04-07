// Copyright (c) 2012-2016 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package camo_test

import (
	"log"

	"go-camo/camo"
	"go-camo/stats"
)

func ExampleProxyMetrics() {
	config := camo.Config{}
	proxy, err := camo.New(config)
	if err != nil {
		log.Fatal("Error: ", err)
	}

	ps := &stats.ProxyStats{}
	proxy.SetMetricsCollector(ps)
}
