// Copyright (c) 2012-2018 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// go-camo daemon (go-camod)
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/cactus/go-camo/pkg/camo"
	"github.com/cactus/go-camo/pkg/router"
	"github.com/cactus/go-camo/pkg/stats"

	"github.com/cactus/mlog"
	flags "github.com/jessevdk/go-flags"
)

var (
	// ServerVersion holds the server version string
	ServerVersion = "no-version"
)

func main() {
	// command line flags
	var opts struct {
		Version             []bool        `short:"V" long:"version" description:"Print version and exit; specify twice to show license information"`
		AddHeaders          []string      `short:"H" long:"header" description:"Extra header to return for each response. This option can be used multiple times to add multiple headers"`
		HMACKey             string        `short:"k" long:"key" description:"HMAC key"`
		SSLKey              string        `long:"ssl-key" description:"ssl private key (key.pem) path"`
		SSLCert             string        `long:"ssl-cert" description:"ssl cert (cert.pem) path"`
		AllowList           string        `long:"allow-list" description:"Text file of hostname allow regexes (one per line)"`
		BindAddress         string        `long:"listen" default:"0.0.0.0:8080" description:"Address:Port to bind to for HTTP"`
		BindAddressSSL      string        `long:"ssl-listen" description:"Address:Port to bind to for HTTPS/SSL/TLS"`
		MaxSize             int64         `long:"max-size" default:"5120" description:"Max allowed response size (KB)"`
		ReqTimeout          time.Duration `long:"timeout" default:"4s" description:"Upstream request timeout"`
		MaxRedirects        int           `long:"max-redirects" default:"3" description:"Maximum number of redirects to follow"`
		Stats               bool          `long:"stats" description:"Enable Stats"`
		NoLogTS             bool          `long:"no-log-ts" description:"Do not add a timestamp to logging"`
		DisableKeepAlivesFE bool          `long:"no-fk" description:"Disable frontend http keep-alive support"`
		DisableKeepAlivesBE bool          `long:"no-bk" description:"Disable backend http keep-alive support"`
		AllowContentVideo   bool          `long:"allow-content-video" description:"Additionally allow 'video/*' content"`
		Verbose             bool          `short:"v" long:"verbose" description:"Show verbose (debug) log level output"`
		ServerName          string        `long:"server-name" default:"go-camo" description:"Value to use for the HTTP server field"`
		ExposeServerVersion bool          `long:"expose-server-version" description:"Include the server version in the HTTP server response header"`
		EnableXFwdFor       bool          `long:"enable-xfwd4" description:"Enable x-forwarded-for passthrough/generation"`
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

	if opts.ExposeServerVersion {
		ServerResponse = fmt.Sprintf("%s %s", opts.ServerName, ServerVersion)
	}

	// setup -V version output
	if len(opts.Version) > 0 {
		fmt.Printf("%s %s (%s,%s-%s)\n", ServerName, ServerVersion, runtime.Version(), runtime.Compiler, runtime.GOARCH)
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

	if opts.BindAddress == "" && opts.BindAddressSSL == "" {
		mlog.Fatal("One of listen or ssl-listen required")
	}

	if opts.BindAddressSSL != "" && opts.SSLKey == "" {
		mlog.Fatal("ssl-key is required when specifying ssl-listen")
	}
	if opts.BindAddressSSL != "" && opts.SSLCert == "" {
		mlog.Fatal("ssl-cert is required when specifying ssl-listen")
	}

	// set keepalive options
	config.DisableKeepAlivesBE = opts.DisableKeepAlivesBE
	config.DisableKeepAlivesFE = opts.DisableKeepAlivesFE

	// other options
	config.EnableXFwdFor = opts.EnableXFwdFor

	// additional content types to allow
	config.AllowContentVideo = opts.AllowContentVideo

	if opts.AllowList != "" {
		b, err := ioutil.ReadFile(opts.AllowList)
		if err != nil {
			mlog.Fatal("Could not read allow-list", err)
		}
		config.AllowList = strings.Split(string(b), "\n")
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

	// convert from KB to Bytes
	config.MaxSize = opts.MaxSize * 1024
	config.RequestTimeout = opts.ReqTimeout
	config.MaxRedirects = opts.MaxRedirects
	config.ServerName = ServerName

	proxy, err := camo.New(config)
	if err != nil {
		mlog.Fatal("Error creating camo", err)
	}

	dumbrouter := &router.DumbRouter{
		ServerName:  ServerResponse,
		AddHeaders:  AddHeaders,
		CamoHandler: proxy,
	}

	if opts.Stats {
		ps := &stats.ProxyStats{}
		proxy.SetMetricsCollector(ps)
		mlog.Printf("Enabling stats at /status")
		dumbrouter.StatsHandler = stats.Handler(ps)
	}

	http.Handle("/", dumbrouter)

	if opts.BindAddress != "" {
		mlog.Printf("Starting server on: %s", opts.BindAddress)
		go func() {
			srv := &http.Server{
				Addr:        opts.BindAddress,
				ReadTimeout: 30 * time.Second}
			mlog.Fatal(srv.ListenAndServe())
		}()
	}
	if opts.BindAddressSSL != "" {
		mlog.Printf("Starting TLS server on: %s", opts.BindAddressSSL)
		go func() {
			srv := &http.Server{
				Addr:        opts.BindAddressSSL,
				ReadTimeout: 30 * time.Second}
			mlog.Fatal(srv.ListenAndServeTLS(opts.SSLCert, opts.SSLKey))
		}()
	}

	// just block. listen and serve will exit the program if they fail/return
	// so we just need to block to prevent main from exiting.
	select {}
}
