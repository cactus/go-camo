// Copyright (c) 2012-2018 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package camo provides an HTTP proxy server with content type
// restrictions as well as regex host allow list support.
package camo

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/cactus/go-camo/pkg/camo/encoding"

	"github.com/cactus/mlog"
)

// Config holds configuration data used when creating a Proxy with New.
type Config struct {
	// HMACKey is a byte slice to be used as the hmac key
	HMACKey []byte
	// AllowList is a list of string represenstations of regex (not compiled
	// regex) that are used as a whitelist filter. If an AllowList is present,
	// then anything not matching is dropped. If no AllowList is present,
	// no Allow filtering is done.
	AllowList []string
	// Server name used in Headers and Via checks
	ServerName string
	// MaxSize is the maximum valid image size response (in bytes).
	MaxSize int64
	// MaxRedirects is the maximum number of redirects to follow.
	MaxRedirects int
	// Request timeout is a timeout for fetching upstream data.
	RequestTimeout time.Duration
	// Keepalive enable/disable
	DisableKeepAlivesFE bool
	DisableKeepAlivesBE bool
	// x-forwarded-for enable/disable
	EnableXFwdFor bool
	// additional content types to allow
	AllowContentVideo bool
	// allow URLs to contain user/pass credentials
	AllowCredetialURLs bool
	// no ip filtering (test mode)
	noIPFiltering bool
}

// ProxyMetrics interface for Proxy to use for stats/metrics.
// This must be goroutine safe, as AddBytes and AddServed will be called from
// many goroutines.
type ProxyMetrics interface {
	AddBytes(bc int64)
	AddServed()
}

// A Proxy is a Camo like HTTP proxy, that provides content type
// restrictions as well as regex host allow list support.
type Proxy struct {
	// compiled allow list regex
	allowList         []*regexp.Regexp
	acceptTypesRe     []*regexp.Regexp
	metrics           ProxyMetrics
	client            *http.Client
	config            *Config
	acceptTypesString string
}

// ServerHTTP handles the client request, validates the request is validly
// HMAC signed, filters based on the Allow list, and then proxies
// valid requests to the desired endpoint. Responses are filtered for
// proper image content types.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if p.metrics != nil {
		p.metrics.AddServed()
	}

	if p.config.DisableKeepAlivesFE {
		w.Header().Set("Connection", "close")
	}

	if req.Header.Get("Via") == p.config.ServerName {
		http.Error(w, "Request loop failure", http.StatusNotFound)
		return
	}

	// split path and get components
	components := strings.Split(req.URL.Path, "/")
	if len(components) < 3 {
		http.Error(w, "Malformed request path", http.StatusNotFound)
		return
	}
	sigHash, encodedURL := components[1], components[2]

	mlog.Debugm("client request", mlog.Map{"req": req})

	sURL, ok := encoding.DecodeURL(p.config.HMACKey, sigHash, encodedURL)
	if !ok {
		http.Error(w, "Bad Signature", http.StatusForbidden)
		return
	}

	mlog.Debugm("signed client url", mlog.Map{"url": sURL})

	u, err := url.Parse(sURL)
	if err != nil {
		mlog.Debugm("url parse error", mlog.Map{"err": err})
		http.Error(w, "Bad url", http.StatusBadRequest)
		return
	}

	err = p.checkURL(u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	nreq, err := http.NewRequest(req.Method, sURL, nil)
	if err != nil {
		mlog.Debugm("could not create NewRequest", mlog.Map{"err": err})
		http.Error(w, "Error Fetching Resource", http.StatusBadGateway)
		return
	}

	// filter headers
	p.copyHeaders(&nreq.Header, &req.Header, &ValidReqHeaders)

	// x-forwarded-for (if appropriate)
	if p.config.EnableXFwdFor {
		xfwd4 := req.Header.Get("X-Forwarded-For")
		if xfwd4 == "" {
			hostIp, _, err := net.SplitHostPort(req.RemoteAddr)
			if err == nil {
				// add forwarded for header, as long as it isn't a private
				// ip address (use isRejectedIP to get private filtering for free)
				if ip := net.ParseIP(hostIp); ip != nil {
					if !isRejectedIP(ip) {
						nreq.Header.Add("X-Forwarded-For", hostIp)
					}
				}
			}
		} else {
			nreq.Header.Add("X-Forwarded-For", xfwd4)
		}
	}

	// add/squash an accept header if the client didn't send one
	nreq.Header.Set("Accept", p.acceptTypesString)

	nreq.Header.Add("User-Agent", p.config.ServerName)
	nreq.Header.Add("Via", p.config.ServerName)

	mlog.Debugm("built outgoing request", mlog.Map{"req": nreq})

	resp, err := p.client.Do(nreq)

	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		mlog.Debugm("could not connect to endpoint", mlog.Map{"err": err})
		// this is a bit janky, but better than peeling off the
		// 3 layers of wrapped errors and trying to get to net.OpErr and
		// still having to rely on string comparison to find out if it is
		// a net.errClosing or not.
		errString := err.Error()
		// go 1.5 changes this to http.httpError
		// go 1.4 has this as net.OpError
		// and the error strings are different depending on which version too.
		if strings.Contains(errString, "timeout") || strings.Contains(errString, "Client.Timeout") {
			http.Error(w, "Error Fetching Resource", http.StatusGatewayTimeout)
		} else if strings.Contains(errString, "use of closed") {
			http.Error(w, "Error Fetching Resource", http.StatusBadGateway)
		} else if strings.Contains(errString, "BadRedirect:") {
			// Got a bad redirect
			mlog.Debugm("response from upstream", mlog.Map{"err": err})
			http.Error(w, "Error Fetching Resource", http.StatusNotFound)
		} else {
			// some other error. call it a not found (camo compliant)
			http.Error(w, "Error Fetching Resource", http.StatusNotFound)
		}
		return
	}

	mlog.Debugm("response from upstream", mlog.Map{"resp": resp})

	// check for too large a response
	if resp.ContentLength > p.config.MaxSize {
		mlog.Debugm("content length exceeded", mlog.Map{"url": sURL})
		http.Error(w, "Content length exceeded", http.StatusNotFound)
		return
	}

	switch resp.StatusCode {
	case 200, 206:
		contentType := resp.Header.Get("Content-Type")

		if contentType == "" {
			mlog.Debug("Empty content-type returned")
			http.Error(w, "Empty content-type returned", http.StatusBadRequest)
			return
		}

		match := false
		for _, re := range p.acceptTypesRe {
			if re.MatchString(contentType) {
				match = true
				break
			}
		}
		if !match {
			mlog.Debugm("Unsupported content-type returned", mlog.Map{"type": u})
			http.Error(w, "Unsupported content-type returned", http.StatusBadRequest)
			return
		}
	case 300:
		http.Error(w, "Multiple choices not supported", http.StatusNotFound)
		return
	case 301, 302, 303, 307:
		// if we get a redirect here, we either disabled following,
		// or followed until max depth and still got one (redirect loop)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	case 304:
		h := w.Header()
		p.copyHeaders(&h, &resp.Header, &ValidRespHeaders)
		w.WriteHeader(304)
		return
	case 404:
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	case 500, 502, 503, 504:
		// upstream errors should probably just 502. client can try later.
		http.Error(w, "Error Fetching Resource", http.StatusBadGateway)
		return
	default:
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	h := w.Header()
	p.copyHeaders(&h, &resp.Header, &ValidRespHeaders)
	w.WriteHeader(resp.StatusCode)

	// get a []byte from bufpool, and put it back on defer
	buf := bufPool.Get().([]byte)
	defer bufPool.Put(buf)

	// since this uses io.Copy/CopyBuffer from the respBody, it is streaming
	// from the request to the response. This means it will nearly
	// always end up with a chunked response.
	bW, err := io.CopyBuffer(w, resp.Body, buf)
	if err != nil {
		// only log broken pipe errors at debug level
		if isBrokenPipe(err) {
			mlog.Debugm("error writing response", mlog.Map{"err": err})
		} else {
			// unknown error and not a broken pipe
			mlog.Printm("error writing response", mlog.Map{"err": err})
		}

		// we may have written some bytes before the error
		if p.metrics != nil && bW != 0 {
			p.metrics.AddBytes(bW)
		}
		return
	}

	if p.metrics != nil {
		p.metrics.AddBytes(bW)
	}
	mlog.Debugm("response to client", mlog.Map{"resp": w})
}

func (p *Proxy) checkURL(reqURL *url.URL) error {
	// reject localhost urls
	uHostname := strings.ToLower(reqURL.Hostname())
	if uHostname == "" || localhostRegex.MatchString(uHostname) {
		return errors.New("Bad url host")
	}

	// if not allowed, reject credentialed/userinfo urls
	if !p.config.AllowCredetialURLs && reqURL.User != nil {
		return errors.New("Userinfo URL rejected")
	}

	// ip/whitelist/blacklist filtering
	if !p.config.noIPFiltering {
		// if allowList is set, require match
		for _, rgx := range p.allowList {
			if rgx.MatchString(uHostname) {
				return errors.New("Allowlist host failure")
			}
		}

		// filter out rejected networks
		if ip := net.ParseIP(uHostname); ip != nil {
			if isRejectedIP(ip) {
				return errors.New("Denylist host failure")
			}
		} else {
			if ips, err := net.LookupIP(uHostname); err == nil {
				for _, ip := range ips {
					if isRejectedIP(ip) {
						return errors.New("Denylist host failure")
					}
				}
			}
		}
	}

	return nil
}

// copy headers from src into dst
// empty filter map will result in no filtering being done
func (p *Proxy) copyHeaders(dst, src *http.Header, filter *map[string]bool) {
	f := *filter
	filtering := false
	if len(f) > 0 {
		filtering = true
	}

	for k, vv := range *src {
		if x, ok := f[k]; filtering && (!ok || !x) {
			continue
		}
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// SetMetricsCollector sets a proxy metrics (ProxyMetrics interface) for
// the proxy
func (p *Proxy) SetMetricsCollector(pm ProxyMetrics) {
	p.metrics = pm
}

// New returns a new Proxy. An error is returned if there was a failure
// to parse the regex from the passed Config.
func New(pc Config) (*Proxy, error) {
	tr := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 3 * time.Second,
		MaxIdleConnsPerHost: 8,
		DisableKeepAlives:   pc.DisableKeepAlivesBE,
		// no need for compression with images
		// some xml/svg can be compressed, but apparently some clients can
		// exhibit weird behavior when those are compressed
		DisableCompression: true,
	}

	// spawn an idle conn trimmer
	go func() {
		// prunes every 5 minutes. this is just a guess at an
		// initial value. very busy severs may want to lower this...
		for {
			time.Sleep(5 * time.Minute)
			tr.CloseIdleConnections()
		}
	}()

	client := &http.Client{
		Transport: tr,
		// timeout
		Timeout: pc.RequestTimeout,
	}

	var allow []*regexp.Regexp
	var c *regexp.Regexp
	var err error
	// compile allow list
	for _, v := range pc.AllowList {
		c, err = regexp.Compile(strings.TrimSpace(v))
		if err != nil {
			return nil, err
		}
		allow = append(allow, c)
	}

	acceptTypes := []string{"image/*"}
	// add additional accept types
	if pc.AllowContentVideo {
		acceptTypes = append(acceptTypes, "video/*")
	}

	var acceptTypesRe []*regexp.Regexp
	for _, v := range acceptTypes {
		c, err = globToRegexp(v)
		if err != nil {
			return nil, err
		}
		acceptTypesRe = append(acceptTypesRe, c)
	}

	p := &Proxy{
		client:            client,
		config:            &pc,
		allowList:         allow,
		acceptTypesString: strings.Join(acceptTypes, ", "),
		acceptTypesRe:     acceptTypesRe,
	}

	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= pc.MaxRedirects {
			mlog.Debug("Got bad redirect: Too many redirects")
			return errors.New("BadRedirect: Too many redirects")
		}
		err := p.checkURL(req.URL)
		if err != nil {
			mlog.Debugm("Got bad redirect", mlog.Map{"url": req})
			return fmt.Errorf("BadRedirect: %s", err)
		}

		return nil
	}

	return p, nil

}
