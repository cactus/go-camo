// Copyright (c) 2012-2019 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// go-camo daemon (go-camod)
package main

import (
	"bufio"
	"context"
	"expvar"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/cactus/go-camo/v2/pkg/camo"
	"github.com/cactus/go-camo/v2/pkg/htrie"
	"github.com/cactus/go-camo/v2/pkg/router"
	"github.com/quic-go/quic-go/http3"

	"github.com/cactus/mlog"
	flags "github.com/jessevdk/go-flags"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
)

const metricNamespace = "camo"

// ServerVersion holds the server version string
var ServerVersion = "no-version"

var (
	// configure histograms and counters
	responseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricNamespace,
			Name:      "response_size_bytes",
			Help:      "A histogram of sizes for proxy responses.",
			Buckets:   prometheus.ExponentialBuckets(1024, 2, 10),
		},
		[]string{},
	)
	responseDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricNamespace,
			Name:      "response_duration_seconds",
			Help:      "A histogram of latencies for proxy responses.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{},
	)
	responseCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "responses_total",
			Help:      "Total HTTP requests processed by the go-camo, excluding scrapes.",
		},
		[]string{"code", "method"},
	)
)

func loadFilterList(fname string) ([]camo.FilterFunc, error) {
	// #nosec
	file, err := os.Open(fname)
	if err != nil {
		return nil, fmt.Errorf("could not open filter-ruleset file: %s", err)
	}
	// #nosec
	defer file.Close()

	allowFilter := htrie.NewURLMatcher()
	denyFilter := htrie.NewURLMatcher()
	hasAllow := false
	hasDeny := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "allow|") {
			line = strings.TrimPrefix(line, "allow")
			err = allowFilter.AddRule(line)
			if err != nil {
				break
			}
			hasAllow = true
		} else if strings.HasPrefix(line, "deny|") {
			line = strings.TrimPrefix(line, "deny")
			err = denyFilter.AddRule(line)
			if err != nil {
				break
			}
			hasDeny = true
		} else {
			fmt.Println("ignoring line: ", line)
		}

		err = scanner.Err()
		if err != nil {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("error building filter ruleset: %s", err)
	}

	// append in order. allow first, then deny filters.
	// first false value aborts the request.
	filterFuncs := make([]camo.FilterFunc, 0)

	if hasAllow {
		filterFuncs = append(filterFuncs, allowFilter.CheckURL)
	}

	// denyFilter returns true on a match. we want to invert this for a deny rule, so
	// any deny rule match should return true, and anything _not_ matching should return false
	// so just wrap and invert the bool.
	if hasDeny {
		denyF := func(u *url.URL) (bool, error) {
			chk, err := denyFilter.CheckURL(u)
			return !chk, err
		}
		filterFuncs = append(filterFuncs, denyF)
	}

	if hasAllow && hasDeny {
		mlog.Print(
			strings.Join(
				[]string{
					"Warning! Allow and Deny rules both supplied.",
					"Having Allow rules means anything not matching an allow rule is denied.",
					"THEN deny rules are evaluated. Be sure this is what you want!",
				},
				" ",
			),
		)
	}

	return filterFuncs, nil
}

func main() {
	// command line flags
	var opts struct {
		HMACKey             string        `short:"k" long:"key" description:"HMAC key"`
		AddHeaders          []string      `short:"H" long:"header" description:"Add additional header to each response. This option can be used multiple times to add multiple headers"`
		BindAddress         string        `long:"listen" default:"0.0.0.0:8080" description:"Address:Port to bind to for HTTP"`
		BindAddressSSL      string        `long:"ssl-listen" description:"Address:Port to bind to for HTTPS/SSL/TLS"`
		BindSocket          string        `long:"socket-listen" description:"Path for unix domain socket to bind to for HTTP"`
		EnableQuic          bool          `long:"quic" description:"Enable http3/quic. Binds to the same port number as ssl-listen but udp+quic."`
		SSLKey              string        `long:"ssl-key" description:"ssl private key (key.pem) path"`
		SSLCert             string        `long:"ssl-cert" description:"ssl cert (cert.pem) path"`
		MaxSize             int64         `long:"max-size" description:"Max allowed response size (KB)"`
		ReqTimeout          time.Duration `long:"timeout" default:"4s" description:"Upstream request timeout"`
		MaxRedirects        int           `long:"max-redirects" default:"3" description:"Maximum number of redirects to follow"`
		Metrics             bool          `long:"metrics" description:"Enable Prometheus compatible metrics endpoint"`
		NoDebugVars         bool          `long:"no-debug-vars" description:"Disable the /debug/vars/ metrics endpoint. This option has no effects when the metrics are not enabled"`
		NoLogTS             bool          `long:"no-log-ts" description:"Do not add a timestamp to logging"`
		LogJson             bool          `long:"log-json" description:"Log in JSON format"`
		DisableKeepAlivesFE bool          `long:"no-fk" description:"Disable frontend http keep-alive support"`
		DisableKeepAlivesBE bool          `long:"no-bk" description:"Disable backend http keep-alive support"`
		AllowContentVideo   bool          `long:"allow-content-video" description:"Additionally allow 'video/*' content"`
		AllowContentAudio   bool          `long:"allow-content-audio" description:"Additionally allow 'audio/*' content"`
		AllowCredentialURLs bool          `long:"allow-credential-urls" description:"Allow urls to contain user/pass credentials"`
		FilterRuleset       string        `long:"filter-ruleset" description:"Text file containing filtering rules (one per line)"`
		ServerName          string        `long:"server-name" default:"go-camo" description:"Value to use for the HTTP server field"`
		ExposeServerVersion bool          `long:"expose-server-version" description:"Include the server version in the HTTP server response header"`
		EnableXFwdFor       bool          `long:"enable-xfwd4" description:"Enable x-forwarded-for passthrough/generation"`
		Verbose             bool          `short:"v" long:"verbose" description:"Show verbose (debug) log level output"`
		Version             []bool        `short:"V" long:"version" description:"Print version and exit; specify twice to show license information"`
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

	// set the server name
	ServerName := opts.ServerName

	// setup the server response field
	ServerResponse := opts.ServerName

	// expand/override server response value if showing version is desired
	if opts.ExposeServerVersion {
		ServerResponse = fmt.Sprintf("%s %s", ServerName, ServerVersion)
	}

	// setup -V version output
	if len(opts.Version) > 0 {
		fmt.Printf("%s %s (%s,%s-%s)\n", "go-camo", ServerVersion, runtime.Version(), runtime.Compiler, runtime.GOARCH)
		if len(opts.Version) > 1 {
			fmt.Printf("\n%s\n", strings.TrimSpace(licenseText))
		}
		os.Exit(0)
	}

	// start out with a very bare logger that only prints
	// the message (no special format or log elements)
	mlog.SetFlags(0)

	config := camo.Config{}
	if hmacKey := os.Getenv("GOCAMO_HMAC"); hmacKey != "" {
		config.HMACKey = []byte(hmacKey)
	}

	// flags override env var
	if opts.HMACKey != "" {
		config.HMACKey = []byte(opts.HMACKey)
	}

	if len(config.HMACKey) == 0 {
		mlog.Fatal("HMAC key required")
	}

	if opts.BindAddress == "" && opts.BindAddressSSL == "" && opts.BindSocket == "" {
		mlog.Fatal("One of listen or ssl-listen required")
	}

	if opts.BindAddressSSL != "" && opts.SSLKey == "" {
		mlog.Fatal("ssl-key is required when specifying ssl-listen")
	}
	if opts.BindAddressSSL != "" && opts.SSLCert == "" {
		mlog.Fatal("ssl-cert is required when specifying ssl-listen")
	}
	if opts.EnableQuic && opts.BindAddressSSL == "" {
		mlog.Fatal("ssl-listen is required when specifying quic")
	}

	// set keepalive options
	config.DisableKeepAlivesBE = opts.DisableKeepAlivesBE
	config.DisableKeepAlivesFE = opts.DisableKeepAlivesFE

	// other options
	config.EnableXFwdFor = opts.EnableXFwdFor
	config.AllowCredentialURLs = opts.AllowCredentialURLs

	// additional content types to allow
	config.AllowContentVideo = opts.AllowContentVideo
	config.AllowContentAudio = opts.AllowContentAudio

	var filters []camo.FilterFunc
	if opts.FilterRuleset != "" {
		filters, err = loadFilterList(opts.FilterRuleset)
		if err != nil {
			mlog.Fatal("Could not read filter-ruleset", err)
		}

	}

	AddHeaders := map[string]string{
		"X-Content-Type-Options":  "nosniff",
		"X-XSS-Protection":        "1; mode=block",
		"Content-Security-Policy": "default-src 'none'; img-src data:; style-src 'unsafe-inline'",
	}

	for _, v := range opts.AddHeaders {
		s := strings.SplitN(v, ":", 2)
		if len(s) != 2 {
			mlog.Printf("ignoring bad header: '%s'", v)
			continue
		}

		s0 := strings.TrimSpace(s[0])
		s1 := strings.TrimSpace(s[1])

		if len(s0) == 0 || len(s1) == 0 {
			mlog.Printf("ignoring bad header: '%s'", v)
			continue
		}
		AddHeaders[s[0]] = s[1]
	}

	// now configure a standard logger
	mlog.SetFlags(mlog.Lstd)
	if opts.NoLogTS {
		mlog.SetFlags(mlog.Flags() ^ mlog.Ltimestamp)
	}

	if opts.Verbose {
		mlog.SetFlags(mlog.Flags() | mlog.Ldebug)
		mlog.Debug("debug logging enabled")
	}

	if opts.LogJson {
		mlog.SetEmitter(&mlog.FormatWriterJSON{})
	}

	// convert from KB to Bytes
	config.MaxSize = opts.MaxSize * 1024
	config.RequestTimeout = opts.ReqTimeout
	config.MaxRedirects = opts.MaxRedirects
	config.ServerName = ServerName

	// configure metrics collection in camo
	if opts.Metrics {
		config.CollectMetrics = true
	}

	proxy, err := camo.NewWithFilters(config, filters)
	if err != nil {
		mlog.Fatal("Error creating camo", err)
	}

	var router http.Handler = &router.DumbRouter{
		ServerName:  ServerResponse,
		AddHeaders:  AddHeaders,
		CamoHandler: proxy,
	}

	mux := http.NewServeMux()

	// configure router endpoint for rendering metrics
	if opts.Metrics {
		mlog.Printf("Enabling metrics at /metrics")
		// Register a version info metric.
		verOverride := os.Getenv("APP_INFO_VERSION")
		if verOverride != "" {
			version.Version = verOverride
		} else {
			version.Version = ServerVersion
		}
		version.Revision = os.Getenv("APP_INFO_REVISION")
		version.Branch = os.Getenv("APP_INFO_BRANCH")
		version.BuildDate = os.Getenv("APP_INFO_BUILD_DATE")
		prometheus.MustRegister(version.NewCollector(metricNamespace))

		// Wrap the dumb router in instrumentation.
		router = promhttp.InstrumentHandlerDuration(responseDuration, router)
		router = promhttp.InstrumentHandlerCounter(responseCount, router)
		router = promhttp.InstrumentHandlerResponseSize(responseSize, router)

		// also configure expvars. this is usually a side effect of importing
		// exvar, as it auto-adds it to the default servemux. Since we want
		// to avoid it being available that when metrics is not enabled, we add
		// it in manually only if metrics IS enabled.
		if !opts.NoDebugVars {
			mux.Handle("/debug/vars", expvar.Handler())
		}
		mux.Handle("/metrics", promhttp.Handler())
	}

	mux.Handle("/", router)

	var httpSrv *http.Server
	var tlsSrv *http.Server
	var quicSrv *http3.Server

	if opts.BindAddress != "" {
		httpSrv = &http.Server{
			Addr:        opts.BindAddress,
			ReadTimeout: 30 * time.Second,
			Handler:     mux,
		}
	}

	if opts.BindAddressSSL != "" {
		tlsSrv = &http.Server{
			Addr:        opts.BindAddressSSL,
			ReadTimeout: 30 * time.Second,
			Handler:     mux,
		}

		if opts.EnableQuic {
			quicSrv = &http3.Server{
				Addr:    opts.BindAddressSSL,
				Handler: mux,
			}
			// wrap default mux to set some default quic reference headers on tls responses
			tlsSrv.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				quicSrv.SetQuicHeaders(w.Header()) // #nosec G104 - ignore error. should only happen if server.Port isn't discoverable
				mux.ServeHTTP(w, r)
			})
		}
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		// we need to reserve to buffer size 1, so the notifier are not blocked
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		s := <-sigint
		mlog.Info("Handling signal:", s)
		mlog.Info("Starting graceful shutdown")

		closeWait := 200 * time.Millisecond

		ctx, cancel := context.WithTimeout(context.Background(), closeWait)
		// Even though ctx may be expired by then, it is good practice to call its
		// cancellation function in any case. Failure to do so may keep the
		// context and its parent alive longer than necessary.
		defer cancel()
		if httpSrv != nil {
			if err := httpSrv.Shutdown(ctx); err != nil {
				mlog.Info("Error gracefully shutting down HTTP server:", err)
			}
		}

		ctx, cancel = context.WithTimeout(context.Background(), closeWait)
		defer cancel()
		if tlsSrv != nil {
			if err := tlsSrv.Shutdown(ctx); err != nil {
				mlog.Info("Error gracefully shutting down HTTP/TLS server:", err)
			}
		}

		if quicSrv != nil {
			if err := quicSrv.CloseGracefully(closeWait); err != nil {
				mlog.Info("Error gracefully shutting down HTTP3/QUIC server:", err)
			}
		}

		close(idleConnsClosed)
	}()

	if opts.BindSocket != "" {
		if _, err := os.Stat(opts.BindSocket); err == nil {
			mlog.Fatal("Cannot bind to unix socket, file aready exists.")
		}

		mlog.Printf("Starting HTTP server on: unix:%s", opts.BindSocket)
		go func() {
			ln, err := net.Listen("unix", opts.BindSocket)
			if err != nil {
				mlog.Fatal("Error listening on unix socket", err)
			}

			if err := httpSrv.Serve(ln); err != http.ErrServerClosed {
				mlog.Fatal(err)
			}
		}()
	}

	if httpSrv != nil {
		mlog.Printf("Starting HTTP server on: tcp:%s", opts.BindAddress)
		go func() {
			if err := httpSrv.ListenAndServe(); err != http.ErrServerClosed {
				mlog.Fatal(err)
			}
		}()
	}

	if tlsSrv != nil {
		mlog.Printf("Starting HTTP/TLS server on: tcp:%s", opts.BindAddressSSL)
		go func() {
			if err := tlsSrv.ListenAndServeTLS(opts.SSLCert, opts.SSLKey); err != http.ErrServerClosed {
				mlog.Fatal(err)
			}
		}()
	}

	if quicSrv != nil {
		mlog.Printf("Starting HTTP3/QUIC server on: udp:%s", opts.BindAddressSSL)
		go func() {
			if err := quicSrv.ListenAndServeTLS(opts.SSLCert, opts.SSLKey); err != http.ErrServerClosed {
				mlog.Fatal(err)
			}
		}()
	}

	// just block waiting for closure
	<-idleConnsClosed
}
