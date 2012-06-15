package main
import (
	"net/http"
	"flag"
	"runtime"
)

var serveDir = flag.String("serverDir", ".", "Directory to serve from")

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	panic(http.ListenAndServe("127.0.0.1:8000", http.FileServer(
			http.Dir(*serveDir))))
}
