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
	"regexp"
	"runtime"
	"time"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// command line flags
	hmacKeyFlag := flag.String("hmacKey", "", "HMAC Key")
	configFileFlag := flag.String("configFile", "", "JSON Config File")
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

	if *configFileFlag != "" {
		b, err := ioutil.ReadFile(*configFileFlag)
		if err != nil {
			log.Fatal("Could not read configFile", err)
		}
		err = json.Unmarshal(b, &config)
		if err != nil {
			log.Fatal("Could not parse configFile", err)
		}
	}

	// flags override config file
	if *hmacKeyFlag != "" {
		config.HmacKey = *hmacKeyFlag
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

	proxy := &camoproxy.ProxyHandler{
		Transport: tr,
		HMacKey:   []byte(config.HmacKey),
		MaxSize:   config.MaxSize * 1024}

	// build/compile regex
	proxy.RegexpAllowlist = make([]*regexp.Regexp, 0)
	proxy.RegexpDenylist = make([]*regexp.Regexp, 0)

	var c *regexp.Regexp
	var err error
	for _, v := range config.Denylist {
		c, err = regexp.Compile(v)
		if err != nil {
			log.Fatal(err)
		}
		proxy.RegexpDenylist = append(proxy.RegexpDenylist, c)
	}
	for _, v := range config.Allowlist {
		c, err = regexp.Compile(v)
		if err != nil {
			log.Fatal(err)
		}
		proxy.RegexpAllowlist = append(proxy.RegexpAllowlist, c)
	}

	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.Handle("/", proxy)
	log.Println("Starting server on", *bindAddress)
	log.Fatal(http.ListenAndServe(*bindAddress, nil))
}
