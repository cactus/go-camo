// Copyright (c) 2012-2023 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// go-camo daemon (go-camod)
package main

import (
	"context"
	"expvar"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"codeberg.org/dropwhile/mlog"
	"github.com/alecthomas/kong"
	"github.com/cactus/go-camo/v2/pkg/camo"
	"github.com/cactus/go-camo/v2/pkg/router"

	"github.com/prometheus/client_golang/prometheus"
	vcoll "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/quic-go/quic-go/http3"
	gomaxecs "github.com/rdforte/gomaxecs/maxprocs"
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

type CLI struct { // betteralign:ignore
	HMACKey      string `name:"key" short:"k" group:"general" env:"GOCAMO_HMAC" help:"HMAC key"`
	AutoMaxProcs bool   `name:"automaxprocs" group:"general" help:"Set GOMAXPROCS automatically to match Linux container CPU quota/limits."`

	BindSocket     string `name:"socket-listen" placeholder:"PATH" group:"listeners" help:"Path for unix domain socket to bind to for HTTP"`
	BindAddress    string `name:"listen" default:"0.0.0.0:8080" group:"listeners" help:"Address:Port to bind to for HTTP"`
	BindAddressSSL string `name:"ssl-listen" placeholder:"HOST_PORT" group:"listeners" help:"Address:Port to bind to for HTTPS/SSL/TLS"`
	EnableQuic     bool   `name:"quic" group:"listeners" help:"Enable http3/quic. Binds to the same port number as ssl-listen but udp+quic."`
	SSLKey         string `name:"ssl-key" placeholder:"PATH" group:"listeners" help:"ssl private key (key.pem) path"`
	SSLCert        string `name:"ssl-cert" placeholder:"PATH" group:"listeners" help:"ssl cert (cert.pem) path"`

	MaxSize             int64         `name:"max-size" placeholder:"INT" group:"proxy" help:"Max allowed response size, in KB"`
	MaxSizeRedirect     string        `name:"max-size-redirect" placeholder:"URL" group:"proxy" help:"redirect to URL when max-size is exceeded"`
	MaxRedirects        int           `name:"max-redirects" default:"3" group:"proxy" help:"Maximum number of redirects to follow"`
	EnableXFwdFor       bool          `name:"xfwd4" env:"GOCAMO_XFWD_FOR" group:"proxy" help:"Enable x-forwarded-for passthrough/generation"`
	DisableKeepAlivesFE bool          `name:"no-fk" group:"proxy" help:"Disable frontend http keep-alive support (frontend)"`
	DisableKeepAlivesBE bool          `name:"no-bk" group:"proxy" help:"Disable backend http keep-alive support (backend)"`
	AllowContentVideo   bool          `name:"allow-content-video" group:"proxy" help:"Additionally allow 'video/*' content"`
	AllowContentAudio   bool          `name:"allow-content-audio" group:"proxy" help:"Additionally allow 'audio/*' content"`
	AllowCredentialURLs bool          `name:"allow-credential-urls" group:"proxy" help:"Allow urls to contain user/pass credentials"`
	ReqTimeout          time.Duration `name:"timeout" default:"4s" group:"proxy" help:"Upstream request timeout (backend)"`
	IdleTimeout         time.Duration `name:"idletimeout" default:"30s" group:"proxy" help:"Maximum amount of time to wait for the next request when keep-alive is enabled (frontend)"`
	ReadTimeout         time.Duration `name:"readtimeout" default:"30s" group:"proxy" help:"Maximum duration for reading the entire request, including the body (frontend)"`
	UserAgent           string        `name:"user-agent" default:"go-camo" group:"proxy" help:"user-agent for outgoing requests"`
	AddHeaders          []string      `name:"header" short:"H" group:"response" help:"Add additional header to each response. This option can be used multiple times to add multiple headers."`
	FilterRuleset       string        `name:"filter-ruleset" group:"proxy" placeholder:"PATH" help:"Text file containing filtering rules (one per line)"`

	ServerName          string `name:"server-name" group:"response" default:"go-camo" help:"Value to use for the HTTP server field"`
	ExposeServerVersion bool   `name:"expose-server-version" group:"response" help:"Include the server version in the HTTP server response header"`

	Metrics     bool `name:"metrics" group:"logmetrics" help:"Enable Prometheus compatible metrics endpoint"`
	NoDebugVars bool `name:"no-debug-vars" group:"logmetrics" help:"Disable the /debug/vars/ metrics endpoint. This option has no effects when the metrics are not enabled."`
	NoLogTS     bool `name:"no-log-ts" group:"logmetrics" help:"Do not add a timestamp to logging"`
	LogJson     bool `name:"log-json" group:"logmetrics" help:"Log in JSON format"`
	Profile     bool `name:"prof" group:"logmetrics" help:"Enable go http profiler endpoint"`

	Verbose bool `name:"verbose" short:"v" help:"Show verbose (debug) log level output"`
	Version int  `name:"version" short:"V" type:"counter" env:"-" help:"Print version and exit; specify twice to show license information."`
}

func (cli *CLI) Run() {
	// set the server name
	ServerName := cli.ServerName

	// setup the server response field
	ServerResponse := cli.ServerName

	// expand/override server response value if showing version is desired
	if cli.ExposeServerVersion {
		ServerResponse = fmt.Sprintf("%s %s", ServerName, ServerVersion)
	}

	// setup -V version output
	if cli.Version > 0 {
		fmt.Printf("%s %s (%s,%s-%s)\n", "go-camo", ServerVersion, runtime.Version(), runtime.Compiler, runtime.GOARCH)
		if cli.Version > 1 {
			fmt.Printf("\n%s\n", strings.TrimSpace(licenseText))
		}
		os.Exit(0)
	}

	// start out with a very bare logger that only prints
	// the message (no special format or log elements)
	mlog.SetFlags(0)

	config := camo.Config{}

	config.HMACKey = []byte(cli.HMACKey)
	if len(config.HMACKey) == 0 {
		mlog.Fatal("HMAC key required")
	}

	if cli.BindAddress == "" && cli.BindAddressSSL == "" && cli.BindSocket == "" {
		mlog.Fatal("One of listen or ssl-listen required")
	}

	if cli.BindAddressSSL != "" && cli.SSLKey == "" {
		mlog.Fatal("ssl-key is required when specifying ssl-listen")
	}
	if cli.BindAddressSSL != "" && cli.SSLCert == "" {
		mlog.Fatal("ssl-cert is required when specifying ssl-listen")
	}
	if cli.EnableQuic && cli.BindAddressSSL == "" {
		mlog.Fatal("ssl-listen is required when specifying quic")
	}

	// set keepalive options
	config.DisableKeepAlivesBE = cli.DisableKeepAlivesBE
	config.DisableKeepAlivesFE = cli.DisableKeepAlivesFE

	// timeouts
	config.RequestTimeout = cli.ReqTimeout
	config.IdleTimeout = cli.IdleTimeout
	config.ReadTimeout = cli.ReadTimeout

	// redirects
	config.MaxRedirects = cli.MaxRedirects
	config.MaxSizeRedirect = cli.MaxSizeRedirect

	// additional content types to allow
	config.AllowContentVideo = cli.AllowContentVideo
	config.AllowContentAudio = cli.AllowContentAudio

	// other options
	config.EnableXFwdFor = cli.EnableXFwdFor
	config.AllowCredentialURLs = cli.AllowCredentialURLs
	config.MaxSize = cli.MaxSize * 1024 // convert from KB to Bytes
	config.ServerName = ServerName
	config.UserAgent = cli.UserAgent

	// configure metrics collection in camo
	config.CollectMetrics = cli.Metrics

	// now configure a standard logger
	mlog.SetFlags(mlog.Lstd)
	if cli.NoLogTS {
		mlog.SetFlags(mlog.Flags() ^ mlog.Ltimestamp)
	}

	if cli.Verbose {
		mlog.SetFlags(mlog.Flags() | mlog.Ldebug)
		mlog.Debug("debug logging enabled")
	}

	var filters []camo.FilterFunc
	if cli.FilterRuleset != "" {
		var err error
		filters, err = loadFilterList(cli.FilterRuleset)
		if err != nil {
			mlog.Fatal("Could not read filter-ruleset", err)
		}
	}

	AddHeaders := map[string]string{
		"X-Content-Type-Options":  "nosniff",
		"X-XSS-Protection":        "1; mode=block",
		"Content-Security-Policy": "default-src 'none'; img-src data:; style-src 'unsafe-inline'",
	}

	for _, v := range cli.AddHeaders {
		mlog.Debugf("will add header -> '%s'", v)
		s := strings.SplitN(v, ":", 2)
		if len(s) != 2 {
			mlog.Printf("ignoring bad header -> '%s'", v)
			continue
		}

		s0 := strings.TrimSpace(s[0])
		s1 := strings.TrimSpace(s[1])

		if len(s0) == 0 || len(s1) == 0 {
			mlog.Printf("ignoring bad header -> '%s'", v)
			continue
		}
		AddHeaders[s[0]] = s[1]
	}

	if cli.AutoMaxProcs && gomaxecs.IsECS() {
		// #nosec G104
		gomaxecs.Set(gomaxecs.WithLogger(func(str string, params ...any) {
			mlog.Info(str, params)
		}))
	}

	if cli.LogJson {
		mlog.SetEmitter(&mlog.FormatWriterJSON{})
	}

	proxy, err := camo.New(config, filters)
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
	if cli.Metrics {
		mlog.Printf("Enabling metrics at /metrics")

		version := os.Getenv("APP_INFO_VERSION")
		if version == "" {
			version = ServerVersion
		}

		revision := os.Getenv("APP_INFO_REVISION")
		tags := ""
		if buildInfo, ok := debug.ReadBuildInfo(); ok {
			for _, v := range buildInfo.Settings {
				if v.Key == "-tags" {
					tags = v.Value
				}
				if v.Key == "vcs.revision" {
					if revision == "" {
						revision = v.Value
					}
				}
			}
		}

		prometheus.MustRegister(prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: metricNamespace,
				Name:      "build_info",
				Help: fmt.Sprintf(
					"A metric with a constant '1' value labeled by version, revision, branch, goversion from which %s was built, and the goos and goarch for the build.",
					metricNamespace,
				),
				ConstLabels: prometheus.Labels{
					"version":   version,
					"revision":  os.Getenv("APP_INFO_REVISION"),
					"branch":    os.Getenv("APP_INFO_BRANCH"),
					"buildDate": os.Getenv("APP_INFO_BUILD_DATE"),
					"goversion": runtime.Version(),
					"goos":      runtime.GOOS,
					"goarch":    runtime.GOARCH,
					"tags":      tags,
				},
			},
			func() float64 { return 1 },
		))
		prometheus.MustRegister(vcoll.NewCollector(metricNamespace))

		// Wrap the dumb router in instrumentation.
		router = promhttp.InstrumentHandlerDuration(responseDuration, router)
		router = promhttp.InstrumentHandlerCounter(responseCount, router)
		router = promhttp.InstrumentHandlerResponseSize(responseSize, router)

		// also configure expvars. this is usually a side effect of importing
		// exvar, as it auto-adds it to the default servemux. Since we want
		// to avoid it being available that when metrics is not enabled, we add
		// it in manually only if metrics IS enabled.
		if !cli.NoDebugVars {
			mux.Handle("/debug/vars", expvar.Handler())
		}
		mux.Handle("/metrics", promhttp.Handler())
	}

	if cli.Profile {
		mlog.Printf("Enabling cpu profile at /debug/pprof")
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		// mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		// mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		// mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	mux.Handle("/", router)

	var httpSrv *http.Server
	var tlsSrv *http.Server
	var quicSrv *http3.Server

	if cli.BindAddress != "" {
		httpSrv = &http.Server{
			Addr:        cli.BindAddress,
			ReadTimeout: config.ReadTimeout,
			IdleTimeout: config.IdleTimeout,
			Handler:     mux,
		}
	}

	if cli.BindAddressSSL != "" {
		tlsSrv = &http.Server{
			Addr:        cli.BindAddressSSL,
			ReadTimeout: config.ReadTimeout,
			IdleTimeout: config.IdleTimeout,
			Handler:     mux,
		}

		if cli.EnableQuic {
			quicSrv = &http3.Server{
				Addr:        cli.BindAddressSSL,
				IdleTimeout: config.IdleTimeout,
				Handler:     mux,
			}
			// wrap default mux to set some default quic reference headers on tls responses
			tlsSrv.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				quicSrv.SetQUICHeaders(w.Header()) // #nosec G104 - ignore error. should only happen if server.Port isn't discoverable
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

		closeWait := 300 * time.Millisecond

		if httpSrv != nil {
			ctx, cancel := context.WithTimeout(context.Background(), closeWait)
			// Even though ctx may be expired by then, it is good practice to call its
			// cancellation function in any case. Failure to do so may keep the
			// context and its parent alive longer than necessary.
			defer cancel()
			if err := httpSrv.Shutdown(ctx); err != nil {
				mlog.Info("Error gracefully shutting down HTTP server:", err)
			}
		}

		if tlsSrv != nil {
			ctx, cancel := context.WithTimeout(context.Background(), closeWait)
			defer cancel()
			if err := tlsSrv.Shutdown(ctx); err != nil {
				mlog.Info("Error gracefully shutting down HTTP/TLS server:", err)
			}
		}

		if quicSrv != nil {
			ctx, cancel := context.WithTimeout(context.Background(), closeWait)
			defer cancel()
			if err := quicSrv.Shutdown(ctx); err != nil {
				mlog.Info("Error gracefully shutting down HTTP3/QUIC server:", err)
			}
		}

		close(idleConnsClosed)
	}()

	if cli.BindSocket != "" {
		if _, err := os.Stat(cli.BindSocket); err == nil {
			mlog.Fatal("Cannot bind to unix socket, file aready exists.")
		}

		mlog.Printf("Starting HTTP server on: unix:%s", cli.BindSocket)
		go func() {
			ln, err := net.Listen("unix", cli.BindSocket)
			if err != nil {
				mlog.Fatal("Error listening on unix socket", err)
			}

			if err := httpSrv.Serve(ln); err != http.ErrServerClosed {
				mlog.Fatal(err)
			}
		}()
	}

	if httpSrv != nil {
		mlog.Printf("Starting HTTP server on: tcp:%s", cli.BindAddress)
		go func() {
			if err := httpSrv.ListenAndServe(); err != http.ErrServerClosed {
				mlog.Fatal(err)
			}
		}()
	}

	if tlsSrv != nil {
		mlog.Printf("Starting HTTP/TLS server on: tcp:%s", cli.BindAddressSSL)
		go func() {
			if err := tlsSrv.ListenAndServeTLS(cli.SSLCert, cli.SSLKey); err != http.ErrServerClosed {
				mlog.Fatal(err)
			}
		}()
	}

	if quicSrv != nil {
		mlog.Printf("Starting HTTP3/QUIC server on: udp:%s", cli.BindAddressSSL)
		go func() {
			if err := quicSrv.ListenAndServeTLS(cli.SSLCert, cli.SSLKey); err != http.ErrServerClosed {
				mlog.Fatal(err)
			}
		}()
	}

	// just block waiting for closure
	<-idleConnsClosed
}

func main() {
	cli := CLI{}
	_ = kong.Parse(
		&cli,
		kong.Name("go-camo"),
		kong.Description("An image proxy that proxies non-secure images over SSL/TLS"),
		kong.UsageOnError(),
		kong.DefaultEnvars("GOCAMO"),
		kong.Groups{
			"general":    "Flags for general behavior",
			"listeners":  "Flags for listeners",
			"proxy":      "Flags for proxy behavior",
			"response":   "Flags for responses",
			"logmetrics": "Flags for logging and metrics",
		},
		kong.Vars{"version": ServerVersion},
	)
	cli.Run()
}
