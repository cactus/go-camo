// go-camo daemon (go-camod)
package main

import (
	"encoding/json"
	"flag"
	"go-camo/camoproxy"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"runtime"
	"time"
	"syscall"
	"github.com/cactus/gologit"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// command line flags
	debug := flag.Bool("debug", false, "Enable Debug Logging")
	hmacKey := flag.String("hmacKey", "", "HMAC Key")
	configFile := flag.String("configFile", "", "JSON Config File")
	maxSize := flag.Int64("maxSize", 5120, "Max size in KB to allow")
	bindAddress := flag.String("bindAddress", "0.0.0.0:8080",
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

	tr := &http.Transport{
		Dial: func(netw, addr string) (net.Conn, error) {
			// 2 second timeout on requests
			timeout := time.Second * 2
			c, err := net.DialTimeout(netw, addr, timeout)
			if err != nil {
				return nil, err
			}
			// also set time limit on reading
			c.SetDeadline(time.Now().Add(timeout))
			return c, nil
		}}

	// spawn an idle conn trimmer
	go func() {
		time.Sleep(5 * time.Minute)
		tr.CloseIdleConnections()
	}()

	// create logger and start toggle on signal handler
	logger := gologit.New(*debug)
	logger.Debugln("Debug logging enabled")
	logger.ToggleOnSignal(syscall.SIGUSR1)

	proxy := camoproxy.New(
		tr, []byte(config.HmacKey), config.Allowlist, config.Denylist,
		config.MaxSize * 1024, logger)

	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.Handle("/", proxy)
	log.Println("Starting server on", *bindAddress)
	log.Fatal(http.ListenAndServe(*bindAddress, nil))
}
