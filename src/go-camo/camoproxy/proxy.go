package camoproxy

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var validReqHeaders = map[string]bool{
	"Accept":            true,
	"Accept-Charset":    true,
	"Accept-Encoding":   true,
	"Accept-Language":   true,
	"Cache-Control":     true,
	"If-None-Match":     true,
	"If-Modified-Since": true,
}

func validateURL(path string, key []byte) (surl string, valid bool) {
	pathParts := strings.SplitN(path[1:], "/", 3)
	valid = false
	if len(pathParts) != 2 {
		log.Println("Bad path format", pathParts)
		return
	}
	hexdig, hexurl := pathParts[0], pathParts[1]
	urlBytes, err := hex.DecodeString(hexurl)
	if err != nil {
		log.Println("Bad Hex Decode", hexurl)
		return
	}
	surl = string(urlBytes)
	//log.Println(surl)
	mac := hmac.New(sha1.New, key)
	mac.Write([]byte(surl))
	macSum := hex.EncodeToString(mac.Sum([]byte{}))
	if macSum != hexdig {
		log.Printf("Bad signature: %s != %s", macSum, hexdig)
		return
	}
	valid = true
	return
}

type ProxyHandler struct {
	Transport       *http.Transport
	HMacKey         []byte
	RegexpAllowlist []*regexp.Regexp
	RegexpDenylist  []*regexp.Regexp
	MaxSize         int64
}

func (p *ProxyHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	//log.Println("Request:", req.URL)

	// do fiddly things
	if req.Method != "GET" {
		log.Println("Something other than GET received", req.Method)
		http.Error(w, "Method Not Implemented", http.StatusNotImplemented)
		return
	}

	surl, ok := validateURL(req.URL.Path, p.HMacKey)
	if !ok {
		http.Error(w, "Bad Signature", http.StatusForbidden)
		return
	}
	//log.Println("URL:", surl)

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
	if len(p.RegexpAllowlist) > 0 {
		matchFound = false
		for _, rgx := range p.RegexpAllowlist {
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
	for _, rgx := range p.RegexpDenylist {
		if rgx.MatchString(u.Host) {
			http.Error(w, "Denylist host failure", http.StatusNotFound)
			return
		}
	}

	nreq, err := http.NewRequest("GET", surl, nil)
	if err != nil {
		log.Println("Something weird happened")
		http.Error(w, "Error Fetching Resource", http.StatusNotFound)
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

	resp, err := p.Transport.RoundTrip(nreq)
	if err != nil {
		log.Println("Could not connect to endpoint", err)
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
		log.Println("Content length exceeded", surl)
		http.Error(w, "Content length exceeded", http.StatusBadRequest)
		return
	}

	switch resp.StatusCode {
	case 300:
		log.Println("Multiple choices not supported")
		http.Error(w, "Multiple choices not supported", http.StatusNotFound)
		return
	case 301, 302, 303:
		// check for redirects. we do not follow.
		log.Println("Refusing to follow redirects")
		http.Error(w, "Refusing to follow redirects", http.StatusNotFound)
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
	//log.Println(req, resp.StatusCode)
}

