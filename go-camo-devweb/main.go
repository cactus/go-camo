// go-camo daemon (go-camod)
package main

import (
	"code.google.com/p/rsc/devweb/slave"
	"encoding/json"
	"github.com/cactus/go-camo/camoproxy"
	"github.com/cactus/gologit"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
)

func main() {
	// Anonymous struct Container for holding configuration parameters parsed
	// from JSON config file.
	config := &camoproxy.ProxyConfig{
		MaxSize:         5120 * 1024,
		FollowRedirects: true,
		RequestTimeout:  5}

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

	proxy := camoproxy.New(config, logger)

	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.Handle("/", proxy)
	log.Println("starting up camoproxy")
	slave.Main()
}
