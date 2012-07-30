// go-camo daemon (go-camod)
package main

import (
	"code.google.com/p/gorilla/mux"
	"code.google.com/p/rsc/devweb/slave"
	"encoding/json"
	"github.com/cactus/go-camo/camoproxy"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func main() {
	// Anonymous struct Container for holding configuration parameters parsed
	// from JSON config file.
	config := camoproxy.Config{
		MaxSize:         5120 * 1024,
		FollowRedirects: true,
		RequestTimeout:  5 * time.Second}

	b, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal("Could not read configFile", err)
	}
	err = json.Unmarshal(b, &config)
	if err != nil {
		log.Fatal("Could not parse configFile", err)
	}

	proxy, err := camoproxy.New(config)
	if err != nil {
		log.Fatal(err)
	}
	logger := camoproxy.Logger
	logger.Set(true)
	logger.Debugln("Debug logging enabled")

	router := mux.NewRouter()
	router.Handle("/favicon.ico", http.NotFoundHandler())
	router.Handle("/status", proxy.StatsHandler())
	router.Handle("/{sigHash}/{encodedUrl}", proxy).Methods("GET")
	http.Handle("/", router)
	log.Println("starting up camoproxy")
	slave.Main()
}
