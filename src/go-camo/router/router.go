// Copyright (c) 2012-2016 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package router

import (
	"io"
	"net/http"
	"strings"
)

type DumbRouter struct {
	ServerName   string
	CamoHandler  http.Handler
	StatsHandler http.HandlerFunc
	AddHeaders   map[string]string
}

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
	w.WriteHeader(200)
	io.WriteString(w, dr.ServerName)
}

func (dr *DumbRouter) HeadGet(w http.ResponseWriter, r *http.Request, handler http.HandlerFunc) {
	if r.Method == "HEAD" || r.Method == "GET" {
		handler(w, r)
	} else {
		http.Error(w, "Method Not Allowed", 405)
	}
	return
}

func (dr *DumbRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// set some default headers
	dr.SetHeaders(w)

	components := strings.Split(r.URL.Path, "/")
	if len(components) == 3 {
		dr.HeadGet(w, r, dr.CamoHandler.ServeHTTP)
		return
	}

	if r.URL.Path == "/status" && dr.StatsHandler != nil {
		dr.HeadGet(w, r, dr.StatsHandler)
		return
	}

	if r.URL.Path == "/" {
		dr.HeadGet(w, r, dr.RootHandler)
		return
	}

	http.Error(w, "404 Not Found", 404)
	return
}
