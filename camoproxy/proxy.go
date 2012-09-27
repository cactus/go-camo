// Package camoproxy provides an HTTP proxy server with content type
// restrictions as well as regex host allow and deny list support.
package camoproxy

import (
	"code.google.com/p/gorilla/mux"
	"github.com/cactus/go-camo/camoproxy/encoding"
	"github.com/cactus/gologit"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"syscall"
)

// Config holds configuration data used when creating a Proxy with New.
type Config struct {
	// HmacKey is a string to be used as the hmac key
	HmacKey         string
	// AllowList is a list of string represenstations of regex (not compiled
	// regex) that are used as a whitelist filter. If an AllowList is present,
	// then anything not matching is dropped. If no AllowList is present,
	// no Allow filtering is done.
	AllowList       []string
	// DenyList is a list of string represenstations of regex (not compiled
	// regex). The deny filter check occurs after the allow filter check
	// (if any).
	DenyList        []string
	// MaxSize is the maximum valid image size response (in bytes).
	MaxSize         int64
	// FollowRedirects is a boolean that specifies whether upstream redirects
	// are followed (10 depth) or not.
	FollowRedirects bool
	// Request timeout is a timeout for fetching upstream data.
	RequestTimeout  time.Duration
}

// Interface for Proxy to use for stats/metrics.
// This must be goroutine safe, as AddBytes and AddServed will be called from
// many goroutines.
type ProxyMetrics interface {
	AddBytes(bc int64)
	AddServed()
	//GetStats() (b uint64, c uint64)
}

// A Proxy is a Camo like HTTP proxy, that provides content type
// restrictions as well as regex host allow and deny list support.
type Proxy struct {
	client    *http.Client
	hmacKey   []byte
	allowList []*regexp.Regexp
	denyList  []*regexp.Regexp
	maxSize   int64
	metrics   ProxyMetrics
}

// ServerHTTP handles the client request, validates the request is validly
// HMAC signed, filters based on the Allow/Deny list, and then proxies
// valid requests to the desired endpoint. Responses are filtered for
// proper image content types.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	gologit.Debugln("Request:", req.URL)
	if p.metrics != nil {
		go p.metrics.AddServed()
	}

	w.Header().Set("Server", ServerNameVer)

	vars := mux.Vars(req)
	surl, ok := encoding.DecodeUrl(&p.hmacKey, vars["sigHash"], vars["encodedUrl"])
	if !ok {
		http.Error(w, "Bad Signature", http.StatusForbidden)
		return
	}
	gologit.Debugln("URL:", surl)

	u, err := url.Parse(surl)
	if err != nil {
		gologit.Debugln(err)
		http.Error(w, "Bad url", http.StatusBadRequest)
		return
	}

	if u.Host == "" {
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

	// filter out denyList urls based on regexes. Do this second
	// as denyList takes precedence
	for _, rgx := range p.denyList {
		if rgx.MatchString(u.Host) {
			http.Error(w, "Denylist host failure", http.StatusNotFound)
			return
		}
	}

	nreq, err := http.NewRequest("GET", surl, nil)
	if err != nil {
		gologit.Debugln("Could not create NewRequest", err)
		http.Error(w, "Error Fetching Resource", http.StatusBadGateway)
		return
	}

	// filter headers
	p.copyHeader(&nreq.Header, &req.Header, &ValidReqHeaders)
	if req.Header.Get("X-Forwarded-For") == "" {
		host, _, err := net.SplitHostPort(req.RemoteAddr)
		if err == nil && !addr1918match.MatchString(host) {
			nreq.Header.Add("X-Forwarded-For", host)
		}
	}
	nreq.Header.Add("connection", "close")
	nreq.Header.Add("user-agent", ServerNameVer)

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
		gologit.Debugln("Content length exceeded", surl)
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
		h.Set("X-Content-Type-Options", "nosniff")
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
	h.Set("X-Content-Type-Options", "nosniff")
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

// sets a proxy metrics (ProxyMetrics interface) for the proxy
func (p *Proxy) SetMetricsCollector(pm ProxyMetrics) {
	p.metrics = pm
}

// Returns a new Proxy. An error is returned if there was a failure
// to parse the regex from the passed Config.
func New(pc Config) (*Proxy, error) {
	tr := &http.Transport{
		Dial: func(netw, addr string) (net.Conn, error) {
			c, err := net.DialTimeout(netw, addr, pc.RequestTimeout)
			if err != nil {
				return nil, err
			}
			// also set time limit on reading
			c.SetDeadline(time.Now().Add(pc.RequestTimeout))
			return c, nil
		}}

	// spawn an idle conn trimmer
	go func() {
		// prunes every 5 minutes. this is just a guess at an
		// initial value. very busy severs may want to lower this...
		time.Sleep(5 * time.Minute)
		tr.CloseIdleConnections()
	}()

	// build/compile regex
	client := &http.Client{Transport: tr}
	if !pc.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return errors.New("Not following redirect")
		}
	}

	allow := make([]*regexp.Regexp, 0)
	deny := make([]*regexp.Regexp, 0)

	var c *regexp.Regexp
	var err error
	for _, v := range pc.DenyList {
		c, err = regexp.Compile(v)
		if err != nil {
			return nil, err
		}
		deny = append(deny, c)
	}
	for _, v := range pc.AllowList {
		c, err = regexp.Compile(v)
		if err != nil {
			return nil, err
		}
		allow = append(allow, c)
	}

	return &Proxy{
		client:    client,
		hmacKey:   []byte(pc.HmacKey),
		allowList: allow,
		denyList:  deny,
		maxSize:   pc.MaxSize}, nil
}
