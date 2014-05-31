// go-camo daemon (go-camod)
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/cactus/go-camo/camo"
	"github.com/cactus/gologit"
	"github.com/gorilla/mux"
	flags "github.com/jessevdk/go-flags"
)

var (
	ServerName    = "go-camo"
	ServerVersion = "no-version"
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
	var opts struct {
		HMACKey             string        `short:"k" long:"key" description:"HMAC key"`
		AddHeaders          []string      `short:"H" long:"header" description:"Extra header to return for each response. This option can be used multiple times to add multiple headers"`
		Stats               bool          `long:"stats" description:"Enable Stats"`
		AllowList           string        `long:"allow-list" description:"Text file of hostname allow regexes (one per line)"`
		MaxSize             int64         `long:"max-size" default:"5120" description:"Max response image size (KB)"`
		ReqTimeout          time.Duration `long:"timeout" default:"4s" description:"Upstream request timeout"`
		MaxRedirects        int           `long:"max-redirects" default:"3" description:"Maximum number of redirects to follow"`
		DisableKeepAlivesFE bool          `long:"no-fk" description:"Disable frontend http keep-alive support"`
		DisableKeepAlivesBE bool          `long:"no-bk" description:"Disable backend http keep-alive support"`
		BindAddress         string        `long:"listen" default:"0.0.0.0:8080" description:"Address:Port to bind to for HTTP"`
		BindAddressSSL      string        `long:"ssl-listen" description:"Address:Port to bind to for HTTPS/SSL/TLS"`
		SSLKey              string        `long:"ssl-key" description:"ssl private key (key.pem) path"`
		SSLCert             string        `long:"ssl-cert" description:"ssl cert (cert.pem) path"`
		Verbose             bool          `short:"v" long:"verbose" description:"Show verbose (debug) log level output"`
		Version             bool          `short:"V" long:"version" description:"print version and exit"`
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

	if opts.Version {
		fmt.Printf("%s %s (%s,%s-%s)\n", ServerName, ServerVersion, runtime.Version(), runtime.Compiler, runtime.GOARCH)
		os.Exit(0)
	}

	config := camo.Config{}
	if hmacKey := os.Getenv("GOCAMO_HMAC"); hmacKey != "" {
		config.HMACKey = []byte(hmacKey)
	}

	// flags override env var
	if opts.HMACKey != "" {
		config.HMACKey = []byte(opts.HMACKey)
	}

	if len(config.HMACKey) == 0 {
		log.Fatal("HMAC key required")
	}

	if opts.BindAddress == "" && opts.BindAddressSSL == "" {
		log.Fatal("One of bind-address or bind-ssl-address required")
	}

	if opts.BindAddressSSL != "" && opts.SSLKey == "" {
		log.Fatal("ssl-key is required when specifying bind-ssl-address")
	}
	if opts.BindAddressSSL != "" && opts.SSLCert == "" {
		log.Fatal("ssl-cert is required when specifying bind-ssl-address")
	}

	// set keepalive options
	opts.DisableKeepAlivesBE = config.DisableKeepAlivesBE
	opts.DisableKeepAlivesFE = config.DisableKeepAlivesFE

	if opts.AllowList != "" {
		b, err := ioutil.ReadFile(opts.AllowList)
		if err != nil {
			log.Fatal("Could not read alllow-list. ", err)
		}
		config.AllowList = strings.Split(string(b), "\n")
	}

	config.AddHeaders = map[string]string{
		"X-Content-Type-Options":  "nosniff",
		"X-XSS-Protection":        "1; mode=block",
		"Content-Security-Policy": "default-src 'none'",
	}

	for _, v := range opts.AddHeaders {
		s := strings.Split(v, ":")
		if len(s) == 2 && len(s[0]) > 0 && len(s[1]) > 0 {
			config.AddHeaders[s[0]] = s[1]
		} else {
			log.Printf("ignoring bad header: '%s'\n", v)
		}
	}

	// convert from KB to Bytes
	config.MaxSize = opts.MaxSize * 1024
	config.RequestTimeout = opts.ReqTimeout
	config.MaxRedirects = opts.MaxRedirects
	config.ServerName = ServerName

	// set logger debug level and start toggle on signal handler
	logger := gologit.Logger
	logger.Set(opts.Verbose)
	logger.Debugln("Debug logging enabled")
	logger.ToggleOnSignal(syscall.SIGUSR1)

	proxy, err := camo.New(config)
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	router.NotFoundHandler = proxy.NotFoundHandler()
	router.Handle("/{sigHash}/{encodedURL}", proxy).Methods("HEAD", "GET")
	router.HandleFunc("/", RootHandler)
	http.Handle("/", router)

	if opts.Stats {
		ps := &ProxyStats{}
		proxy.SetMetricsCollector(ps)
		log.Println("Enabling stats at /status")
		router.Handle("/status", StatsHandler(ps))
	}

	if opts.BindAddress != "" {
		log.Println("Starting server on", opts.BindAddress)
		go func() {
			srv := &http.Server{
				Addr:        opts.BindAddress,
				ReadTimeout: 30 * time.Second}
			log.Fatal(srv.ListenAndServe())
		}()
	}
	if opts.BindAddressSSL != "" {
		log.Println("Starting TLS server on", opts.BindAddressSSL)
		go func() {
			srv := &http.Server{
				Addr:        opts.BindAddressSSL,
				ReadTimeout: 30 * time.Second}
			log.Fatal(srv.ListenAndServeTLS(opts.SSLCert, opts.SSLKey))
		}()
	}

	// just block. listen and serve will exit the program if they fail/return
	// so we just need to block to prevent main from exiting.
	select {}
}
