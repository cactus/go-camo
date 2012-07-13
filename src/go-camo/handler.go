package main

import (
	"io"
	"net/http"
)

// A simple http hander for / that returns "Go-Camo"
func RootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	io.WriteString(w, "Go-Camo")
}

