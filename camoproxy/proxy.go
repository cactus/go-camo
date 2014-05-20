// Package camoproxy provides an HTTP proxy server with content type
// restrictions as well as regex host allow list support.
package camoproxy

import (
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/cactus/go-camo/camoproxy/encoding"
	"github.com/cactus/gologit"
	"github.com/gorilla/mux"
	httpclient "github.com/mreiferson/go-httpclient"
)

// Config holds configuration data used when creating a Proxy with New.
type Config struct {
	// HmacKey is a string to be used as the hmac key
	HmacKey string
	// AllowList is a list of string represenstations of regex (not compiled
	// regex) that are used as a whitelist filter. If an AllowList is present,
	// then anything not matching is dropped. If no AllowList is present,
	// no Allow filtering is done.
	AllowList []string
	// MaxSize is the maximum valid image size response (in bytes).
	MaxSize int64
	// MaxRedirects is the maximum number of redirects to follow.
	MaxRedirects int
	// Request timeout is a timeout for fetching upstream data.
	RequestTimeout time.Duration
	// Server name used in Headers and Via checks
	ServerName string
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
	client     *http.Client
	hmacKey    []byte
	allowList  []*regexp.Regexp
	maxSize    int64
	metrics    ProxyMetrics
	serverName string
}

// ServerHTTP handles the client request, validates the request is validly
// HMAC signed, filters based on the Allow list, and then proxies
// valid requests to the desired endpoint. Responses are filtered for
// proper image content types.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	gologit.Debugln("Request:", req.URL)
	if p.metrics != nil {
		go p.metrics.AddServed()
	}

	w.Header().Set("Server", p.serverName)
	w.Header().Set("X-Content-Type-Options", "nosniff")

	if req.Header.Get("Via") == p.serverName {
		http.Error(w, "Request loop failure", http.StatusNotFound)
		return
	}

	vars := mux.Vars(req)
	sURL, ok := encoding.DecodeURL(p.hmacKey, vars["sigHash"], vars["encodedURL"])
	if !ok {
		http.Error(w, "Bad Signature", http.StatusForbidden)
		return
	}
	gologit.Debugln("URL:", sURL)

	u, err := url.Parse(sURL)
	if err != nil {
		gologit.Debugln(err)
		http.Error(w, "Bad url", http.StatusBadRequest)
		return
	}

	u.Host = strings.ToLower(u.Host)
	if u.Host == "" || localhostRegex.MatchString(u.Host) {
		http.Error(w, "Bad url", http.StatusNotFound)
		return
	}

	// if allowList is set, require match
	matchFound := true
	if len(p.allowList) > 0 {
		matchFound = false
		for _, rgx := range p.allowList {
			if rgx.MatchString(u.Host) {
				matchFound = true
			}
		}
	}
	if !matchFound {
		http.Error(w, "Allowlist host failure", http.StatusNotFound)
		return
	}

	// filter out rfc1918 hosts
	ip := net.ParseIP(u.Host)
	if ip != nil {
		if addr1918PrefixRegex.MatchString(ip.String()) {
			http.Error(w, "Denylist host failure", http.StatusNotFound)
			return
		}
	}

	nreq, err := http.NewRequest("GET", sURL, nil)
	if err != nil {
		gologit.Debugln("Could not create NewRequest", err)
		http.Error(w, "Error Fetching Resource", http.StatusBadGateway)
		return
	}

	// filter headers
	p.copyHeader(&nreq.Header, &req.Header, &ValidReqHeaders)
	if req.Header.Get("X-Forwarded-For") == "" {
		host, _, err := net.SplitHostPort(req.RemoteAddr)
		if err == nil && !addr1918PrefixRegex.MatchString(host) {
			nreq.Header.Add("X-Forwarded-For", host)
		}
	}

	// add an accept header if the client didn't send one
	if nreq.Header.Get("Accept") == "" {
		nreq.Header.Add("Accept", "image/*")
	}

	nreq.Header.Add("connection", "close")
	nreq.Header.Add("user-agent", p.serverName)
	nreq.Header.Add("via", p.serverName)

	resp, err := p.client.Do(nreq)
	if err != nil {
		gologit.Debugln("Could not connect to endpoint", err)
		if strings.Contains(err.Error(), "timeout") {
			http.Error(w, "Error Fetching Resource", http.StatusBadGateway)
		} else {
			http.Error(w, "Error Fetching Resource", http.StatusNotFound)
		}
		return
	}
	defer resp.Body.Close()

	// check for too large a response
	if resp.ContentLength > p.maxSize {
		gologit.Debugln("Content length exceeded", sURL)
		http.Error(w, "Content length exceeded", http.StatusNotFound)
		return
	}

	switch resp.StatusCode {
	case 200:
		// check content type
		ct, ok := resp.Header[http.CanonicalHeaderKey("content-type")]
		if !ok || ct[0][:6] != "image/" {
			gologit.Debugln("Non-Image content-type returned", u)
			http.Error(w, "Non-Image content-type returned",
				http.StatusBadRequest)
			return
		}
	case 300:
		gologit.Debugln("Multiple choices not supported")
		http.Error(w, "Multiple choices not supported", http.StatusNotFound)
		return
	case 301, 302, 303, 307:
		// if we get a redirect here, we either disabled following,
		// or followed until max depth and still got one (redirect loop)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	case 304:
		h := w.Header()
		p.copyHeader(&h, &resp.Header, &ValidRespHeaders)
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
	p.copyHeader(&h, &resp.Header, &ValidRespHeaders)
	h.Set("Date", formattedDate.String())
	w.WriteHeader(resp.StatusCode)

	// since this uses io.Copy from the respBody, it is streaming
	// from the request to the response. This means it will nearly
	// always end up with a chunked response.
	// Change to the following to send whole body at once, and
	// read whole body at once too:
	//    body, err := ioutil.ReadAll(resp.Body)
	//    if err != nil {
	//        gologit.Println("Error writing response:", err)
	//    }
	//    w.Write(body)
	// Might use quite a bit of memory though. Untested.
	bW, err := io.Copy(w, resp.Body)
	if err != nil {
		// only log if not broken pipe. broken pipe means the client
		// terminated conn for some reason.
		opErr, ok := err.(*net.OpError)
		if !ok || opErr.Err != syscall.EPIPE {
			gologit.Println("Error writing response:", err)
		}
		return
	}

	if p.metrics != nil {
		go p.metrics.AddBytes(bW)
	}
	gologit.Debugln(req, resp.StatusCode)
}

// copy headers from src into dst
// empty filter map will result in no filtering being done
func (p *Proxy) copyHeader(dst, src *http.Header, filter *map[string]bool) {
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
	tr := &httpclient.Transport{
		MaxIdleConnsPerHost: 8,
		ConnectTimeout:      2 * time.Second,
		RequestTimeout:      pc.RequestTimeout}

	// spawn an idle conn trimmer
	go func() {
		// prunes every 5 minutes. this is just a guess at an
		// initial value. very busy severs may want to lower this...
		for {
			time.Sleep(5 * time.Minute)
			tr.CloseIdleConnections()
		}
	}()

	client := &http.Client{Transport: tr}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= pc.MaxRedirects {
			return errors.New("Too many redirects")
		}
		return nil
	}

	var allow []*regexp.Regexp
	var c *regexp.Regexp
	var err error
	// compile allow list
	for _, v := range pc.AllowList {
		c, err = regexp.Compile(v)
		if err != nil {
			return nil, err
		}
		allow = append(allow, c)
	}

	return &Proxy{
		client:     client,
		hmacKey:    []byte(pc.HmacKey),
		allowList:  allow,
		maxSize:    pc.MaxSize,
		serverName: pc.ServerName}, nil
}
