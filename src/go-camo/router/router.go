// Copyright (c) 2012-2016 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package router

import (
	"io"
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

// RootHandler is a simple http hander for / that returns "Go-Camo"
func (dr *DumbRouter) RootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	// Status 200 is the default. No need to set explicitly here.
	io.WriteString(w, dr.ServerName)
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

	if dr.StatsHandler != nil && r.URL.Path == "/status" {
		dr.StatsHandler(w, r)
		return
	}

	if r.URL.Path == "/" {
		dr.RootHandler(w, r)
		return
	}

	http.Error(w, "404 Not Found", 404)
}
