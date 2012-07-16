// Package camoproxy provides an HTTP proxy server with content type
// restrictions as well as regex host allow and deny list support.
package camoproxy

import (
	"code.google.com/p/gorilla/mux"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/cactus/gologit"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// Logger for handling logging.
var Logger = gologit.New(false)

// Headers that are acceptible to pass from the client to the remote
// server. Only those present and true, are forwarded.
var validReqHeaders = map[string]bool{
	"Accept":            true,
	"Accept-Charset":    true,
	"Accept-Encoding":   true,
	"Accept-Language":   true,
	"Cache-Control":     true,
	"If-None-Match":     true,
	"If-Modified-Since": true,
}

// Headers that are acceptible to pass from the remote server to the
// client. Only those present and true, are forwarded. Empty implies
// no filtering.
var validRespHeaders = map[string]bool{}

// ProxyConfig holds configuration data used when creating a
// ProxyHandler with New.
type ProxyConfig struct {
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
	RequestTimeout  uint
}

// A ProxyHandler is a Camo like HTTP proxy, that provides content type
// restrictions as well as regex host allow and deny list support
type ProxyHandler struct {
	client    *http.Client
	hmacKey   []byte
	allowList []*regexp.Regexp
	denyList  []*regexp.Regexp
	maxSize   int64
	stats     *proxyStats
}

// StatsHandler returns an http.Handler that returns running totals and stats
// about the server.
func (p *ProxyHandler) StatsHandler() http.Handler {
	p.stats.Enable = true
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(200)
			c, b := p.stats.GetStats()
			fmt.Fprintf(w, "ClientsServed, BytesServed\n%d, %d\n", c, b)
		})
}

// ServerHTTP handles the client request, validates the request is validly
// HMAC signed, filters based on the Allow/Deny list, and then proxies
// valid requests to the desired endpoint. Responses are filtered for 
// proper image content types.
func (p *ProxyHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	Logger.Debugln("Request:", req.URL)
	if p.stats.Enable {
		p.stats.AddServed()
	}

	vars := mux.Vars(req)
	surl, ok := DecodeUrl(&p.hmacKey, vars["sigHash"], vars["encodedUrl"])
	if !ok {
		http.Error(w, "Bad Signature", http.StatusForbidden)
		return
	}
	Logger.Debugln("URL:", surl)

	u, err := url.Parse(surl)
	if err != nil {
		Logger.Debugln(err)
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
		Logger.Debugln("Could not create NewRequest", err)
		http.Error(w, "Error Fetching Resource", http.StatusBadGateway)
		return
	}

	// filter headers
	p.copyHeader(&nreq.Header, &req.Header, &validReqHeaders)
	nreq.Header.Add("connection", "close")
	nreq.Header.Add("user-agent", "pew pew pew")

	resp, err := p.client.Do(nreq)
	if err != nil {
		Logger.Debugln("Could not connect to endpoint", err)
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
		Logger.Debugln("Content length exceeded", surl)
		http.Error(w, "Content length exceeded", http.StatusNotFound)
		return
	}

	switch resp.StatusCode {
	case 200:
		// check content type
		ct, ok := resp.Header[http.CanonicalHeaderKey("content-type")]
		if !ok || ct[0][:6] != "image/" {
			Logger.Debugln("Non-Image content-type returned", u)
			http.Error(w, "Non-Image content-type returned",
				http.StatusBadRequest)
			return
		}
	case 300:
		Logger.Debugln("Multiple choices not supported")
		http.Error(w, "Multiple choices not supported", http.StatusNotFound)
		return
	case 301, 302, 303:
		// if we get a redirect here, we either disabled following, 
		// or followed until max depth and still got one (redirect loop)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	case 304:
		h := w.Header()
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
	p.copyHeader(&h, &resp.Header, &validRespHeaders)
	h.Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(resp.StatusCode)
	bW, err := io.Copy(w, resp.Body)
	if err != nil {
		Logger.Println("Error writing response:", err)
		return
	}
	if p.stats.Enable {
		p.stats.AddBytes(bW)
	}
	Logger.Debugln(req, resp.StatusCode)
}

// copy headers from src into dst
// empty filter map will result in no filtering being done
func (p *ProxyHandler) copyHeader(dst, src *http.Header, filter *map[string]bool) {
	f := *filter
	filtering := false
	if len(f) > 0 {
		filtering = true
	}

	for k, vv := range *src {
		if x, ok := f[k]; filtering && !ok && !x {
			continue
		}
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// Returns a new ProxyHandler. An error is returned if there was a failure
// to parse the regex from the passed ProxyConfig.
func New(pc ProxyConfig) (*ProxyHandler, error) {
	tr := &http.Transport{
		Dial: func(netw, addr string) (net.Conn, error) {
			// 2 second timeout on requests
			timeout := time.Second * time.Duration(pc.RequestTimeout)
			c, err := net.DialTimeout(netw, addr, timeout)
			if err != nil {
				return nil, err
			}
			// also set time limit on reading
			c.SetDeadline(time.Now().Add(timeout))
			return c, nil
		}}

	// spawn an idle conn trimmer
	go func() {
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

	return &ProxyHandler{
		client:    client,
		hmacKey:   []byte(pc.HmacKey),
		allowList: allow,
		denyList:  deny,
		maxSize:   pc.MaxSize,
		stats:     &proxyStats{}}, nil
}

// DecodeUrl ensures the url is properly verified via HMAC, and then
// unencodes the url, returning the url (if valid) and whether the
// HMAC was verified.
func DecodeUrl(hmackey *[]byte, hexdig string, hexurl string) (surl string, valid bool) {
	urlBytes, err := hex.DecodeString(hexurl)
	if err != nil {
		Logger.Debugln("Bad Hex Decode", hexurl)
		return
	}
	surl = string(urlBytes)
	mac := hmac.New(sha1.New, *hmackey)
	mac.Write([]byte(surl))
	macSum := hex.EncodeToString(mac.Sum([]byte{}))
	if macSum != hexdig {
		Logger.Debugf("Bad signature: %s != %s\n", macSum, hexdig)
		return
	}
	valid = true
	return
}

func EncodeUrl(hmacKey *[]byte, oUrl string) string {
	mac := hmac.New(sha1.New, *hmacKey)
	mac.Write([]byte(oUrl))
	macSum := hex.EncodeToString(mac.Sum([]byte{}))
	encodedUrl := hex.EncodeToString([]byte(oUrl))
	hexurl := "/" + macSum + "/" + encodedUrl
	return hexurl
}

