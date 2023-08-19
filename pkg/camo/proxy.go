// Copyright (c) 2012-2019 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package camo provides an HTTP proxy server with content type
// restrictions as well as regex host allow list support.
package camo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"strings"
	"syscall"
	"time"

	"github.com/cactus/go-camo/v2/pkg/camo/encoding"
	"github.com/cactus/go-camo/v2/pkg/htrie"

	"github.com/cactus/mlog"
)

//lint:file-ignore ST1005 Ignore string case error to maintain existing responses
// as some may people parse these

// Config holds configuration data used when creating a Proxy with New.
type Config struct {
	// HMACKey is a byte slice to be used as the hmac key
	HMACKey []byte
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
	AllowContentAudio bool
	// allow URLs to contain user/pass credentials
	AllowCredentialURLs bool
	// Whether to call/increment metrics
	CollectMetrics bool
	// no ip filtering (test mode)
	noIPFiltering bool
}

// The FilterFunc type is a function that validates a *url.URL
// A true value approves the url. A false value rejects the url.
type FilterFunc func(*url.URL) (bool, error)

// A Proxy is a Camo like HTTP proxy, that provides content type
// restrictions as well as regex host allow list support.
type Proxy struct {
	client              *http.Client
	config              *Config
	upstreamProxyConfig *upstreamProxyConfig
	acceptTypesFilter   *htrie.GlobPathChecker
	acceptTypesString   string
	filters             []FilterFunc
	filtersLen          int
}

// ServerHTTP handles the client request, validates the request is validly
// HMAC signed, filters based on the Allow list, and then proxies
// valid requests to the desired endpoint. Responses are filtered for
// proper image content types.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
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

	if mlog.HasDebug() {
		mlog.Debugm("client request", httpReqToMlogMap(req))
	}

	sURL, ok := encoding.DecodeURL(p.config.HMACKey, sigHash, encodedURL)
	if !ok {
		http.Error(w, "Bad Signature", http.StatusForbidden)
		return
	}

	if mlog.HasDebug() {
		mlog.Debugm("signed client url", mlog.Map{"url": sURL})
	}

	u, err := url.Parse(sURL)
	if err != nil {
		if mlog.HasDebug() {
			mlog.Debugm("url parse error", mlog.Map{"err": err})
		}
		http.Error(w, "Bad url", http.StatusBadRequest)
		return
	}

	err = p.checkURL(u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	nreq, err := http.NewRequestWithContext(req.Context(), req.Method, sURL, nil)
	if err != nil {
		if mlog.HasDebug() {
			mlog.Debugm("could not create NewRequest", mlog.Map{"err": err})
		}
		http.Error(w, "Error Fetching Resource", http.StatusBadGateway)
		return
	}

	// filter headers
	p.copyHeaders(&nreq.Header, &req.Header, &ValidReqHeaders)

	// x-forwarded-for (if appropriate)
	if p.config.EnableXFwdFor {
		xfwd4 := req.Header.Get("X-Forwarded-For")
		if xfwd4 == "" {
			hostIP, _, err := net.SplitHostPort(req.RemoteAddr)
			if err == nil {
				// add forwarded for header, as long as it isn't a private
				// ip address (use isRejectedIP to get private filtering for free)
				if ip := net.ParseIP(hostIP); ip != nil {
					if !isRejectedIP(ip) {
						nreq.Header.Add("X-Forwarded-For", hostIP)
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

	if mlog.HasDebug() {
		mlog.Debugm("built outgoing request", mlog.Map{"req": httpReqToMlogMap(nreq)})
	}

	resp, err := p.client.Do(nreq)

	if resp != nil {
		defer func() {
			if err := resp.Body.Close(); err != nil {
				if mlog.HasDebug() {
					mlog.Debug("error on body close. ignoring.")
				}
			}
		}()
	}

	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			// handle client aborting request early in the request lifetime
			if mlog.HasDebug() {
				mlog.Debugm("client aborted request (early)", httpReqToMlogMap(req))
			}
			return
		case errors.Is(err, ErrRedirect):
			// Got a bad redirect
			if mlog.HasDebug() {
				mlog.Debugm("bad redirect from server", mlog.Map{"err": err})
			}
			http.Error(w, "Error Fetching Resource", http.StatusNotFound)
			return
		case errors.Is(err, ErrRejectIP):
			// Got a deny list failure from Dial.Control
			if mlog.HasDebug() {
				mlog.Debugm("ip filter rejection from dial.control", mlog.Map{"err": err})
			}
			http.Error(w, "Error Fetching Resource", http.StatusNotFound)
			return
		case errors.Is(err, ErrInvalidHostPort):
			// Got a deny list failure from Dial.Control
			if mlog.HasDebug() {
				mlog.Debugm("invalid host/port rejection from dial.control", mlog.Map{"err": err})
			}
			http.Error(w, "Error Fetching Resource", http.StatusNotFound)
			return
		case errors.Is(err, ErrInvalidNetType):
			// Got a deny list failure from Dial.Control
			if mlog.HasDebug() {
				mlog.Debugm("net type rejection from dial.control", mlog.Map{"err": err})
			}
			http.Error(w, "Error Fetching Resource", http.StatusNotFound)
			return
		}

		// handle other errors
		if mlog.HasDebug() {
			mlog.Debugm("could not connect to endpoint", mlog.Map{"err": err})
		}

		// this is a bit janky, but some of these errors don't support
		// the newer error semantics yet...
		switch errString := err.Error(); {
		case containsOneOf(errString, "timeout", "Client.Timeout"):
			http.Error(w, "Error Fetching Resource", http.StatusGatewayTimeout)
		case strings.Contains(errString, "use of closed"):
			http.Error(w, "Error Fetching Resource", http.StatusBadGateway)
		default:
			// some other error. call it a not found (camo compliant)
			http.Error(w, "Error Fetching Resource", http.StatusNotFound)
		}
		return
	}

	if mlog.HasDebug() {
		mlog.Debugm("response from upstream", httpRespToMlogMap(resp))
	}

	// check for too large a response
	if p.config.MaxSize > 0 && resp.ContentLength > p.config.MaxSize {
		if p.config.CollectMetrics {
			contentLengthExceeded.Inc()
		}
		if mlog.HasDebug() {
			mlog.Debugm("content length exceeded", mlog.Map{"url": sURL})
		}
		http.Error(w, "Content length exceeded", http.StatusNotFound)
		return
	}

	var responseContentType string
	switch resp.StatusCode {
	case 200, 206:
		contentType := resp.Header.Get("Content-Type")

		// early abort if content type is empty. avoids empty mime parsing overhead.
		if contentType == "" {
			if mlog.HasDebug() {
				mlog.Debug("Empty content-type returned")
			}
			http.Error(w, "Empty content-type returned", http.StatusBadRequest)
			return
		}

		// content-type: image/png; charset=...
		// be careful of malformed content type records. apparently some browsers
		// only loosely validate this, and tend to pick whichever one "seems correct",
		// or have a "default fallback" such as text/html, which would be insecure in
		// this context.
		// content-type: image/png, text/html; charset=...
		mediatype, param, err := mime.ParseMediaType(contentType)
		if err != nil || !p.acceptTypesFilter.CheckPath(mediatype) {
			if mlog.HasDebug() {
				mlog.Debugm("Unsupported content-type returned", mlog.Map{"type": u})
			}
			http.Error(w, "Unsupported content-type returned", http.StatusBadRequest)
			return
		}

		// add params back in, as certain content types have various optional and/or
		// required parameters.
		// refs: https://www.iana.org/assignments/media-types/media-types.xhtml
		responseContentType = mime.FormatMediaType(mediatype, param)

		// also check if the parsed content type is empty, just to be safe.
		// note: round trip of mediatype and params _should_ be fine, but guard
		// against implementation changes or bugs.
		if responseContentType == "" {
			if mlog.HasDebug() {
				mlog.Debug("Unsupported content-type returned")
			}
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
	// set content type based on parsed content type, not originally supplied
	h.Set("content-type", responseContentType)
	w.WriteHeader(resp.StatusCode)

	// get a []byte from bufpool, and put it back on defer
	buf := *bufPool.Get().(*[]byte)
	defer bufPool.Put(&buf)

	// wrap body in limit reader, so even while chunk/streaming, we read
	// less than desired max size
	var bodyRC io.ReadCloser = resp.Body
	if p.config.MaxSize > 0 {
		bodyRC = NewLimitReadCloser(resp.Body, p.config.MaxSize)
	}

	// since this uses io.Copy/CopyBuffer from the respBody, it is streaming
	// from the request to the response. This means it will nearly
	// always end up with a chunked response.
	written, err := io.CopyBuffer(w, bodyRC, buf)
	if err != nil {
		if p.config.CollectMetrics {
			responseFailed.Inc()
		}
		if err == context.Canceled || errors.Is(err, context.Canceled) {
			// client aborted/closed request, which is why copy failed to finish
			if mlog.HasDebug() {
				mlog.Debugm("client aborted request (late)", mlog.Map{"req": req})
			}
			return
		}

		// got an early EOF from the server side
		if errors.Is(err, io.ErrUnexpectedEOF) {
			if mlog.HasDebug() {
				mlog.Debugm("server sent unexpected EOF", mlog.Map{"req": req})
			}
			return
		}

		// only log broken pipe errors at debug level
		if isBrokenPipe(err) {
			if mlog.HasDebug() {
				mlog.Debugm("error writing response", mlog.Map{"err": err, "req": req})
			}
			return
		}

		// unknown error (not: a broken pipe; server early EOF; client close)
		mlog.Printm("error writing response", mlog.Map{"err": err, "req": req})
		return
	}

	if p.config.MaxSize > 0 && written >= p.config.MaxSize {
		if p.config.CollectMetrics {
			responseTruncated.Inc()
		}
		if mlog.HasDebug() {
			mlog.Debugm("response to client truncated: size > MaxSize", mlog.Map{"req": req})
		}
		return
	}

	if mlog.HasDebug() {
		mlog.Debugm("response to client", mlog.Map{"resp": w})
	}
}

func (p *Proxy) checkURL(reqURL *url.URL) error {
	// ensure we have an http or https url
	// (eg. no file:// or other)
	scheme := reqURL.Scheme
	if !(scheme == "http" || scheme == "https") {
		return errors.New("Bad url scheme")
	}

	uHostname := reqURL.Hostname()
	// reject empy hostnames
	if uHostname == "" {
		return errors.New("Bad url host")
	}

	// ensure valid idna lookup mapping for hostname
	// this is also used by htrie filtering (localsFilter and p.filters)
	cleanHostname, err := htrie.CleanHostname(uHostname)
	if err != nil {
		mlog.Infof("Filter lookup rejected: malformed hostname: %s", uHostname)
		return errors.New("Bad url host")
	}

	// reject localhost urls
	// lower case for matching is done by IdnaLookupMap above, so no need to
	// ToLower here also
	if localsFilter.CheckCleanHostname(cleanHostname) {
		return errors.New("Bad url host")
	}

	// if not allowed, reject credentialed/userinfo urls
	if !p.config.AllowCredentialURLs && reqURL.User != nil {
		return errors.New("Userinfo URL rejected")
	}

	// evaluate filters. first false (or filter error) value "fails"
	for i := 0; i < p.filtersLen; i++ {
		if chk, err := p.filters[i](reqURL); err != nil || !chk {
			return errors.New("Rejected due to filter-ruleset")
		}
	}

	// if using a proxy config, block directly url requests to the proxy
	if p.upstreamProxyConfig.hasProxy &&
		p.upstreamProxyConfig.matchesAny(uHostname, reqURL.Port()) {
		return errors.New("Rejected due to filter-ruleset")
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

// NewWithFilters returns a new Proxy that utilises the passed in proxy filters.
// filters are evaluated in order, and the first false response from a filter
// function halts further evaluation and fails the request.
func NewWithFilters(pc Config, filters []FilterFunc) (*Proxy, error) {
	proxy, err := New(pc)
	if err != nil {
		return nil, err
	}

	filterFuncs := make([]FilterFunc, 0)
	// check for nil entries, and copy the slice in case the original
	// is mutated.
	for _, filter := range filters {
		if filter != nil {
			filterFuncs = append(filterFuncs, filter)
		}
	}
	proxy.filters = filterFuncs
	proxy.filtersLen = len(filterFuncs)
	return proxy, nil
}

// New returns a new Proxy. Returns an error if Proxy could not be constructed.
func New(pc Config) (*Proxy, error) {
	doFiltering := !pc.noIPFiltering

	upstreamProxyConf := &upstreamProxyConfig{}
	upstreamProxyConf.init()

	dailer := &net.Dialer{
		Timeout:   3 * time.Second,
		KeepAlive: 30 * time.Second,
		// Move ip filtering to dial.control, this avoids cases where
		// an adversary may return an unblocked ip on name resolution
		// the first time, and a blocked ip the second time.
		// if the server's resolver has no caching, or hits an unlucky ttl expiry,
		// an ip check in checkURL may pass (unblocked ip), then the bad ip may
		// be connected to (re-resolve in dailer).
		// Moving the ip filtering here avoids that.
		Control: func(network string, address string, conn syscall.RawConn) error {
			// reject not tcp/tcp6 connection attempts
			if !(network == "tcp4" || network == "tcp6") {
				return fmt.Errorf("%s is not a safe network type: %w", network, ErrInvalidNetType)
			}

			// ip/allow-list/deny-list filtering
			if doFiltering {
				host, port, err := net.SplitHostPort(address)
				if err != nil {
					return fmt.Errorf("%s:%s is not a valid host/port pair: %w", address, err, ErrInvalidHostPort)
				}

				// filter out rejected networks
				if ip := net.ParseIP(host); ip != nil {
					if isRejectedIP(ip) && !upstreamProxyConf.matchesIP(ip, port) {
						return ErrRejectIP
					}
				} else {
					if ips, err := net.LookupIP(host); err == nil {
						for _, ip := range ips {
							if isRejectedIP(ip) && !upstreamProxyConf.matchesIP(ip, port) {
								return ErrRejectIP
							}
						}
					}
				}
			}
			return nil
		},
	}

	tr := &http.Transport{
		DialContext: dailer.DialContext,

		// Use proxy from environment
		// It uses HTTP proxies as directed by the $HTTP_PROXY and $NO_PROXY
		// (or $http_proxy and $no_poxy) environment variables.
		//
		// Also Note: This is invoked as a once.Do deep in go http client, and
		// reified, to avoid constant overhead. Nice.
		Proxy: http.ProxyFromEnvironment,

		// max idle conns. Go DetaultTransport uses 100, which seems like a
		// fairly reasonable number. Very busy servers may wish to raise
		// or lower this value.
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 8,

		// more defaults from DefaultTransport, with a few tweaks
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   3 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,

		DisableKeepAlives: pc.DisableKeepAlivesBE,
		// no need for compression with images
		// some xml/svg can be compressed, but apparently some clients can
		// exhibit weird behavior when those are compressed
		DisableCompression: true,
	}

	client := &http.Client{
		Transport: tr,
		// timeout
		Timeout: pc.RequestTimeout,
	}

	acceptTypes := []string{"image/*"}
	// add additional accept types, if appropriate
	if pc.AllowContentVideo {
		acceptTypes = append(acceptTypes, "video/*")
	}
	if pc.AllowContentAudio {
		acceptTypes = append(acceptTypes, "audio/*")
	}

	// re-use the htrie glob path checker for accept types validation
	acceptTypesFilter := htrie.NewGlobPathChecker()
	for _, v := range acceptTypes {
		err := acceptTypesFilter.AddRule("|i|" + v)
		if err != nil {
			return nil, err
		}
	}

	p := &Proxy{
		client:              client,
		config:              &pc,
		acceptTypesString:   strings.Join(acceptTypes, ", "),
		acceptTypesFilter:   acceptTypesFilter,
		upstreamProxyConfig: upstreamProxyConf,
	}

	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= pc.MaxRedirects {
			if mlog.HasDebug() {
				mlog.Debug("Got bad redirect: Too many redirects", mlog.Map{"url": req})
			}
			return fmt.Errorf("Too many redirects: %w", ErrRedirect)
		}
		err := p.checkURL(req.URL)
		if err != nil {
			if mlog.HasDebug() {
				mlog.Debugm("Got bad redirect", mlog.Map{"url": req})
			}
			return fmt.Errorf("Bad redirect: %w", ErrRedirect)
		}

		return nil
	}

	return p, nil
}
