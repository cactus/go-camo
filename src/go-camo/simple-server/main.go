// simple-server daemon
package main

import (
	"log"
	"net/http"
	"os"
	"runtime"

	flags "github.com/jessevdk/go-flags"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// command line flags
	var opts struct {
		BindAddress string `long:"listen" short:"l" default:"0.0.0.0:8000" description:"Address:Port to bind to for HTTP"`
	}

	// parse said flags
	parser := flags.NewParser(&opts, flags.Default)
	parser.Usage = "[OPTIONS] DIR"
	args, err := parser.Parse()
	if err != nil {
		if e, ok := err.(*flags.Error); ok {
			if e.Type == flags.ErrHelp {
				os.Exit(0)
			}
		}
		os.Exit(1)
	}

	var dirname string
	alen := len(args)
	switch {
	case alen < 1:
		dirname = "."
	case alen == 1:
		dirname = args[0]
	case alen > 1:
		log.Fatal("Too many arguments")
	}

	panic(http.ListenAndServe(opts.BindAddress, http.FileServer(http.Dir(dirname))))
}
