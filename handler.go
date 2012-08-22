package main

import (
	"fmt"
	"github.com/cactus/go-camo/camoproxy"
	"io"
	"net/http"
)

// A simple http hander for / that returns "Go-Camo"
func RootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	io.WriteString(w, "Go-Camo")
}

// StatsHandler returns an http.Handler that returns running totals and stats
// about the server.
func StatsHandler(pm camoproxy.ProxyMetrics) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(200)
			c, b := pm.GetStats()
			fmt.Fprintf(w, "ClientsServed, BytesServed\n%d, %d\n", c, b)
		})
}
