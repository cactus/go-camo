// Copyright (c) 2012-2023 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package router

import (
	"fmt"
	"net/http"
	"strings"
)

// DumbRouter is a basic, special purpose, http router
type DumbRouter struct {
	CamoHandler   http.Handler
	dateGenerator fmt.Stringer
	AddHeaders    map[string]string
	ServerName    string
}

// HealthCheckHandler is HTTP handler for confirming the backend service
// is available from an external client, such as a load balancer.
func (dr *DumbRouter) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// ServeHTTP fulfills the http server interface
func (dr *DumbRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// set some default headers
	h := w.Header()
	for k, v := range dr.AddHeaders {
		h.Set(k, v)
	}
	h.Set("Date", dr.dateGenerator.String())
	h.Set("Server", dr.ServerName)

	if r.Method != "HEAD" && r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.URL.Path == "/healthcheck" {
		dr.HealthCheckHandler(w, r)
		return
	}

	// Ever so slightly faster to check first char to ensure it is a slash,
	// then check remaining string for remaining slash, than to just
	// count the whole string.
	if r.URL.Path[0] == '/' && strings.Count(r.URL.Path[1:], "/") == 1 {
		dr.CamoHandler.ServeHTTP(w, r)
		return
	}

	http.Error(w, "404 Not Found", http.StatusNotFound)
}

func NewDumbRouter(
	serverName string,
	headers map[string]string,
	camoHandler http.Handler,
) *DumbRouter {
	return &DumbRouter{
		ServerName:    serverName,
		AddHeaders:    headers,
		CamoHandler:   camoHandler,
		dateGenerator: newiHTTPDate(),
	}
}
