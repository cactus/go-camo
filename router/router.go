package router

import (
	"net/http"
	"strings"
)

type DumbRouter struct {
	ServerName   string
	AddHeaders   map[string]string
	RootHandler  http.HandlerFunc
	StatsHandler http.HandlerFunc
	CamoHandler  http.Handler
}

func (dr *DumbRouter) SetHeaders(w http.ResponseWriter) {
	h := w.Header()
	for k, v := range dr.AddHeaders {
		h.Set(k, v)
	}
	h.Set("Date", formattedDate.String())
	h.Set("Server", dr.ServerName)
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
