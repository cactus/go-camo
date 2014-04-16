// simple-server daemon
package main

import (
	flags "github.com/jessevdk/go-flags"
	"net/http"
	"os"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// command line flags
	var opts struct {
			BindAddress string `long:"listen" short:"l" default:"0.0.0.0:8000" description:"Address:Port to bind to for HTTP"`
			ServeDir    string `long:"serve-dir" short:"d" default:"." description:"Directory to serve from"`
	}

	// parse said flags
	_, err := flags.Parse(&opts)
	if err != nil {
			if e, ok := err.(*flags.Error); ok {
					if e.Type == flags.ErrHelp {
							os.Exit(0)
					}
			}
			os.Exit(1)
	}

	panic(http.ListenAndServe(opts.BindAddress, http.FileServer(http.Dir(opts.ServeDir))))
}
