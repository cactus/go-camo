// go-camo daemon (go-camod)
package main

import (
	"encoding/json"
	"flag"
	"go-camo/camoproxy"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"syscall"
	"github.com/cactus/gologit"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// command line flags
	debug := flag.Bool("debug", false, "Enable Debug Logging")
	hmacKey := flag.String("hmac-key", "", "HMAC Key")
	configFile := flag.String("config-file", "", "JSON Config File")
	maxSize := flag.Int64("max-size", 5120, "Max size in KB to allow")
	follow := flag.Bool("follow-redirects", false,
		"Enable following upstream redirects")
	bindAddress := flag.String("bind-address", "0.0.0.0:8080",
		"Address:Port to bind to")
	// parse said flags
	flag.Parse()

	// Anonymous struct Container for holding configuration parameters parsed
	// from JSON config file.
	config := &struct {
		HmacKey   string
		Allowlist []string
		Denylist  []string
		MaxSize   int64}{}

	if *configFile != "" {
		b, err := ioutil.ReadFile(*configFile)
		if err != nil {
			log.Fatal("Could not read configFile", err)
		}
		err = json.Unmarshal(b, &config)
		if err != nil {
			log.Fatal("Could not parse configFile", err)
		}
	}

	// flags override config file
	if *hmacKey != "" {
		config.HmacKey = *hmacKey
	}
	if config.MaxSize == 0 {
		config.MaxSize = *maxSize
	}

	// create logger and start toggle on signal handler
	logger := gologit.New(*debug)
	logger.Debugln("Debug logging enabled")
	logger.ToggleOnSignal(syscall.SIGUSR1)

	proxy := camoproxy.New(
		[]byte(config.HmacKey), config.Allowlist, config.Denylist,
		config.MaxSize * 1024, logger, *follow)

	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.Handle("/", proxy)
	log.Println("Starting server on", *bindAddress)
	log.Fatal(http.ListenAndServe(*bindAddress, nil))
}
