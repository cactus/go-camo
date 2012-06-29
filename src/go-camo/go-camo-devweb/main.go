// go-camo daemon (go-camod)
package main

import (
	"encoding/json"
	"code.google.com/p/rsc/devweb/slave"
	"go-camo/camoproxy"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"github.com/cactus/gologit"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Anonymous struct Container for holding configuration parameters parsed
	// from JSON config file.
	config := &struct {
		HmacKey   string
		Allowlist []string
		Denylist  []string
		MaxSize   int64}{}

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
	logger.Debugln("Debug logging enabled")

	proxy := camoproxy.New(
		[]byte(config.HmacKey), config.Allowlist, config.Denylist,
		5120 * 1024, logger, true)

	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.Handle("/", proxy)
	slave.Main()
}
