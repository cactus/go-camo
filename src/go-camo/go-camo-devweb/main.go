// go-camo daemon (go-camod)
package main

import (
	"code.google.com/p/rsc/devweb/slave"
	"encoding/json"
	"github.com/cactus/gologit"
	"go-camo/camoproxy"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Anonymous struct Container for holding configuration parameters parsed
	// from JSON config file.
	config := &struct {
		HmacKey   string
		Allowlist []string
		Denylist  []string
		MaxSize   int64
	}{}

	b, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal("Could not read configFile", err)
	}
	err = json.Unmarshal(b, &config)
	if err != nil {
		log.Fatal("Could not parse configFile", err)
	}

	// create logger and start toggle on signal handler
	logger := gologit.New(true)

	proxy := camoproxy.New(
		[]byte(config.HmacKey), config.Allowlist, config.Denylist,
		5120*1024, logger, true, 5)

	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.Handle("/", proxy)
	log.Println("starting up camoproxy")
	slave.Main()
}
