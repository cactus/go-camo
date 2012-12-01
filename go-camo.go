// go-camo daemon (go-camod)
package main

import (
	"code.google.com/p/gorilla/mux"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/cactus/go-camo/camoproxy"
	"github.com/cactus/gologit"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"syscall"
	"time"
)

func main() {
	var gmx int
	if gmxEnv := os.Getenv("GOMAXPROCS"); gmxEnv != "" {
		gmx, _ = strconv.Atoi(gmxEnv)
	} else {
		gmx = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(gmx)

	// command line flags
	debug := flag.Bool("debug", false, "Enable Debug Logging")
	stats := flag.Bool("stats", false, "Enable Stats")
	hmacKey := flag.String("hmac-key", "", "HMAC Key")
	configFile := flag.String("config-file", "", "JSON Config File")
	maxSize := flag.Int64("max-size", 5120, "Max response image size (KB)")
	reqTimeout := flag.Duration("timeout", 4*time.Second,
		"Upstream request timeout")
	noFollow := flag.Bool("no-follow-redirects", false,
		"Disable following upstream redirects")
	bindAddress := flag.String("bind-address", "0.0.0.0:8080",
		"Address:Port to bind to for HTTP")
	bindAddressSSL := flag.String("bind-address-ssl", "",
		"Address:Port to bind to for HTTPS/SSL/TLS")
	sslKey := flag.String("ssl-key", "", "ssl private key (key.pem) path")
	sslCert := flag.String("ssl-cert", "", "ssl cert (cert.pem) path")
	version := flag.Bool("version", false, "print version and exit")
	// parse said flags
	flag.Parse()

	if *version {
		fmt.Println(camoproxy.ServerNameVer)
		os.Exit(0)
	}

	config := camoproxy.Config{}

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

	if *bindAddress == "" && *bindAddressSSL == "" {
		log.Fatal("One of bind-address or bind-ssl-address required")
	}

	if *bindAddressSSL != "" && *sslKey == "" {
		log.Fatal("ssl-key is required when specifying bind-ssl-address")
	}
	if *bindAddressSSL != "" && *sslCert == "" {
		log.Fatal("ssl-cert is required when specifying bind-ssl-address")
	}

	// convert from KB to Bytes
	config.MaxSize = config.MaxSize * 1024
	config.RequestTimeout = *reqTimeout
	config.NoFollowRedirects = *noFollow

	// set logger debug level and start toggle on signal handler
	logger := gologit.Logger
	logger.Set(*debug)
	logger.Debugln("Debug logging enabled")
	logger.ToggleOnSignal(syscall.SIGUSR1)

	proxy, err := camoproxy.New(config)
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	router.Handle("/favicon.ico", http.NotFoundHandler())
	router.Handle("/{sigHash}/{encodedUrl}", proxy).Methods("GET")
	router.HandleFunc("/", RootHandler)
	http.Handle("/", router)

	if *stats {
		ps := &ProxyStats{}
		proxy.SetMetricsCollector(ps)
		log.Println("Enabling stats at /status")
		router.Handle("/status", StatsHandler(ps))
	}

	if *bindAddress != "" {
		log.Println("Starting server on", *bindAddress)
		go func() {
			log.Fatal(http.ListenAndServe(*bindAddress, nil))
		}()
	}
	if *bindAddressSSL != "" {
		log.Println("Starting TLS server on", *bindAddressSSL)
		go func() {
			log.Fatal(http.ListenAndServeTLS(
				*bindAddressSSL, *sslCert, *sslKey, nil))
		}()
	}

	// just block. listen and serve will exit the program if they fail/return
	// so we just need to block to prevent main from exiting.
	select {}
}
