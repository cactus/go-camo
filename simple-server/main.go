// simple-server daemon
package main

import (
	"flag"
	"net/http"
	"runtime"
)

var serveDir = flag.String("d", ".", "Directory to serve from")

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	panic(http.ListenAndServe("127.0.0.1:8000", http.FileServer(
		http.Dir(*serveDir))))
}
