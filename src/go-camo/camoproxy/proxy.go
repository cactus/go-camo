// Package camoproxy provides an HTTP proxy server with content type
// restrictions as well as regex host allow and deny list support.
package camoproxy

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"github.com/cactus/gologit"
)

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

// A ProxyHandler is a Camo like HTTP proxy, that provides content type
// restrictions as well as regex host allow and deny list support
type ProxyHandler struct {
	Client          *http.Client
	HMacKey         []byte
	Allowlist       []*regexp.Regexp
	Denylist        []*regexp.Regexp
	MaxSize         int64
	log             *gologit.DebugLogger
}

// ServerHTTP handles the client request, validates the request is validly
// HMAC signed, filters based on the Allow/Deny list, and then proxies
// valid requests to the desired endpoint. Responses are filtered for 
// proper image content types.
func (p *ProxyHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	p.log.Debugln("Request:", req.URL)

	// do fiddly things
	if req.Method != "GET" {
		log.Println("Something other than GET received", req.Method)
		http.Error(w, "Method Not Implemented", http.StatusNotImplemented)
		return
	}

	surl, ok := p.validateURL(req.URL.Path, p.HMacKey)
	if !ok {
		http.Error(w, "Bad Signature", http.StatusForbidden)
		return
	}
	p.log.Debugln("URL:", surl)

	u, err := url.Parse(surl)
	if err != nil {
		log.Println(err)
		http.Error(w, "Bad url", http.StatusBadRequest)
		return
	}

	if u.Host == "" {
		http.Error(w, "Bad url", http.StatusNotFound)
		return
	}

	// if Allowlist is set, require match
	matchFound := true
	if len(p.Allowlist) > 0 {
		matchFound = false
		for _, rgx := range p.Allowlist {
			if rgx.MatchString(u.Host) {
				matchFound = true
			}
		}
	}
	if !matchFound {
		http.Error(w, "Allowlist host failure", http.StatusNotFound)
		return
	}

	// filter out Denylist urls based on regexes. Do this second
	// as Denylist takes precedence
	for _, rgx := range p.Denylist {
		if rgx.MatchString(u.Host) {
			http.Error(w, "Denylist host failure", http.StatusNotFound)
			return
		}
	}

	nreq, err := http.NewRequest("GET", surl, nil)
	if err != nil {
		p.log.Debugln("Could not create NewRequest", err)
		http.Error(w, "Error Fetching Resource", http.StatusBadGateway)
		return
	}

	// filter headers
	for hdr, val := range req.Header {
		if validReqHeaders[hdr] {
			nreq.Header[hdr] = val
		}
	}
	nreq.Header.Add("connection", "close")
	nreq.Header.Add("user-agent", "pew pew pew")

	resp, err := p.Client.Do(nreq)
	if err != nil {
		p.log.Debugln("Could not connect to endpoint", err)
		if strings.Contains(err.Error(), "timeout") {
			http.Error(w, "Error Fetching Resource", http.StatusBadGateway)
		} else {
			http.Error(w, "Error Fetching Resource", http.StatusNotFound)
		}
		return
	}
	defer resp.Body.Close()

	// check for too large a response
	if resp.ContentLength > p.MaxSize {
		p.log.Debugln("Content length exceeded", surl)
		http.Error(w, "Content length exceeded", http.StatusNotFound)
		return
	}

	switch resp.StatusCode {
	case 300:
		log.Println("Multiple choices not supported")
		http.Error(w, "Multiple choices not supported", http.StatusNotFound)
		return
	case 301, 302, 303:
		// if we get a redirect here, we either disabled following, 
		// or followed until max depth and still got one (redirect loop)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	case 404:
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	case 200:
		// check content type
		ct, ok := resp.Header[http.CanonicalHeaderKey("content-type")]
		if !ok || ct[0][:6] != "image/" {
			log.Println("Non-Image content-type returned", u)
			http.Error(w, "Non-Image content-type returned",
				http.StatusBadRequest)
			return
		}
	}

	for hdr, val := range resp.Header {
		h := w.Header()
		h[hdr] = val
	}

	h := w.Header()
	h.Add("X-Content-Type-Options", "nosniff")
	if resp.StatusCode == 304 && h.Get("Content-Type") != "" {
		h.Del("Content-Type")
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
	p.log.Debugln(req, resp.StatusCode)
}


// validateURL ensures the url is properly verified via HMAC, and then
// unencodes the url, returning the url (if valid) and whether the
// HMAC was verified.
func (p *ProxyHandler) validateURL(path string, key []byte) (surl string, valid bool) {
	pathParts := strings.SplitN(path[1:], "/", 3)
	valid = false
	if len(pathParts) != 2 {
		p.log.Println("Bad path format", pathParts)
		return
	}
	hexdig, hexurl := pathParts[0], pathParts[1]
	urlBytes, err := hex.DecodeString(hexurl)
	if err != nil {
		p.log.Println("Bad Hex Decode", hexurl)
		return
	}
	surl = string(urlBytes)
	p.log.Debugln(surl)
	mac := hmac.New(sha1.New, key)
	mac.Write([]byte(surl))
	macSum := hex.EncodeToString(mac.Sum([]byte{}))
	if macSum != hexdig {
		p.log.Printf("Bad signature: %s != %s", macSum, hexdig)
		return
	}
	valid = true
	return
}


func New(hmacKey []byte, allowList []string, denyList []string, maxSize int64, logger *gologit.DebugLogger, follow bool) *ProxyHandler {
	tr := &http.Transport{
		Dial: func(netw, addr string) (net.Conn, error) {
			// 2 second timeout on requests
			timeout := time.Second * 3
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
	if !follow {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return errors.New("Not following redirect")
		}
	}

	allow := make([]*regexp.Regexp, 0)
	deny := make([]*regexp.Regexp, 0)

	var c *regexp.Regexp
	var err error
	for _, v := range denyList {
		c, err = regexp.Compile(v)
		if err != nil {
			log.Fatal(err)
		}
		deny = append(deny, c)
	}
	for _, v := range allowList {
		c, err = regexp.Compile(v)
		if err != nil {
			log.Fatal(err)
		}
		allow = append(allow, c)
	}

	return &ProxyHandler{
		Client:    client,
		HMacKey:   hmacKey,
		Allowlist: allow,
		Denylist:  deny,
		MaxSize:   maxSize,
		log:       logger}
}
