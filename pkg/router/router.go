// Copyright (c) 2012-2018 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package router

import (
	"net/http"
	"strings"
)

// DumbRouter is a basic, special purpose, http router
type DumbRouter struct {
	ServerName   string
	CamoHandler  http.Handler
	StatsHandler http.HandlerFunc
	AddHeaders   map[string]string
}

// SetHeaders sets the headers on the response
func (dr *DumbRouter) SetHeaders(w http.ResponseWriter) {
	h := w.Header()
	for k, v := range dr.AddHeaders {
		h.Set(k, v)
	}
	h.Set("Date", formattedDate.String())
	h.Set("Server", dr.ServerName)
}

// HealthCheckHandler is HTTP handler for confirming the backend service
// is available from an external client, such as a load balancer.
func (dr *DumbRouter) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

// ServeHTTP fulfills the http server interface
func (dr *DumbRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// set some default headers
	dr.SetHeaders(w)

	if r.Method != "HEAD" && r.Method != "GET" {
		http.Error(w, "Method Not Allowed", 405)
	}

	components := strings.Split(r.URL.Path, "/")
	if len(components) == 3 {
		dr.CamoHandler.ServeHTTP(w, r)
		return
	}

	if r.URL.Path == "/healthcheck" {
		dr.HealthCheckHandler(w, r)
		return
	}

	if dr.StatsHandler != nil && r.URL.Path == "/status" {
		dr.StatsHandler(w, r)
		return
	}

	http.Error(w, "404 Not Found", 404)
}
