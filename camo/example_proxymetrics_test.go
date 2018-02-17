// Copyright (c) 2012-2016 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package camo_test

import (
	"fmt"
	"os"

	"github.com/cactus/go-camo/camo"
	"github.com/cactus/go-camo/stats"
)

func ExampleProxyMetrics() {
	config := camo.Config{}
	proxy, err := camo.New(config)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}

	ps := &stats.ProxyStats{}
	proxy.SetMetricsCollector(ps)
}
