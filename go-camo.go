// go-camo daemon (go-camod)
package main

import (
	"encoding/json"
	"fmt"
	"github.com/cactus/go-camo/camoproxy"
	"github.com/cactus/gologit"
	"github.com/gorilla/mux"
	flags "github.com/jessevdk/go-flags"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"syscall"
	"time"
)

const (
	ServerName    = "go-camo"
	ServerVersion = "0.2.0"
)

// Server Name with version
var ServerNameVer = fmt.Sprintf("%s %s", ServerName, ServerVersion)

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
		ConfigFile     string        `short:"c" long:"config" description:"JSON Config File"`
		HmacKey        string        `short:"k" long:"key" description:"HMAC key"`
		Stats          bool          `long:"stats" description:"Enable Stats"`
		MaxSize        int64         `long:"max-size" default:"5120" description:"Max response image size (KB)"`
		ReqTimeout     time.Duration `long:"timeout" default:"4s" description:"Upstream request timeout"`
		MaxRedirects   int           `long:"max-redirects" default:"3" description:"Maximum number of redirects to follow"`
		BindAddress    string        `long:"listen" default:"0.0.0.0:8080" description:"Address:Port to bind to for HTTP"`
		BindAddressSSL string        `long:"ssl-listen" description:"Address:Port to bind to for HTTPS/SSL/TLS"`
		SSLKey         string        `long:"ssl-key" description:"ssl private key (key.pem) path"`
		SSLCert        string        `long:"ssl-cert" description:"ssl cert (cert.pem) path"`
		Verbose        bool          `short:"v" long:"verbose" description:"Show verbose (debug) log level output"`
		Version        bool          `short:"V" long:"version" description:"print version and exit"`
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
		fmt.Printf("%s (%s,%s-%s)\n", ServerNameVer, runtime.Version(), runtime.Compiler, runtime.GOARCH)
		os.Exit(0)
	}

	config := camoproxy.Config{}

	if opts.ConfigFile != "" {
		b, err := ioutil.ReadFile(opts.ConfigFile)
		if err != nil {
			log.Fatal("Could not read configFile", err)
		}
		err = json.Unmarshal(b, &config)
		if err != nil {
			log.Fatal("Could not parse configFile", err)
		}
	}

	// env var overrides config file
	if hmacKey := os.Getenv("GOCAMO_HMAC"); hmacKey != "" {
		config.HmacKey = hmacKey
	}

	// flags override config file and env var
	if opts.HmacKey != "" {
		config.HmacKey = opts.HmacKey
	}

	if config.MaxSize == 0 {
		config.MaxSize = opts.MaxSize
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

	// convert from KB to Bytes
	config.MaxSize = config.MaxSize * 1024
	config.RequestTimeout = opts.ReqTimeout
	config.MaxRedirects = opts.MaxRedirects
	config.ServerName = ServerName

	// set logger debug level and start toggle on signal handler
	logger := gologit.Logger
	logger.Set(opts.Verbose)
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
