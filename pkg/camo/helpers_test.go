// Copyright (c) 2012-2023 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package camo

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/cactus/go-camo/v2/pkg/encoding"
	"github.com/cactus/go-camo/v2/pkg/router"
)

func makeReq(config Config, testURL string) (*http.Request, error) {
	k := []byte(config.HMACKey)
	hexDig, hexURL := encoding.B64EncodeURL(k, testURL)
	out := "http://example.com" + "/" + hexDig + "/" + hexURL
	req, err := http.NewRequest("GET", out, nil)
	if err != nil {
		return nil, fmt.Errorf("Error building req url '%s': %s", testURL, err.Error())
	}
	return req, nil
}

func processRequest(req *http.Request, status int, camoConfig Config, filters []FilterFunc) (*http.Response, error) {
	camoServer, err := New(camoConfig, filters)
	if err != nil {
		return nil, fmt.Errorf("Error building Camo: %s", err.Error())
	}

	router := router.NewDumbRouter(
		camoConfig.ServerName,
		map[string]string{"X-Go-Camo": "test"},
		camoServer,
	)

	record := httptest.NewRecorder()
	router.ServeHTTP(record, req)
	resp := record.Result()
	te := resp.Header.Get("Transfer-Encoding")
	// recorder does not seem to set resp.TransferEncoding correctly,
	// so if the response header is present, then set it.
	if len(te) > 0 && len(resp.TransferEncoding) == 0 {
		encodings := strings.Split(te, ",")
		slices.Reverse(encodings)
		resp.TransferEncoding = encodings
	}
	if got, want := resp.StatusCode, status; got != want {
		return resp, fmt.Errorf("response code = %d, wanted %d", got, want)
	}
	return resp, nil
}

func makeTestReq(testURL string, status int, config Config) (*http.Response, error) {
	req, err := makeReq(config, testURL)
	if err != nil {
		return nil, err
	}

	result, err := processRequest(req, status, config, nil)
	if err != nil {
		return result, err
	}
	return result, nil
}

type Tuple[A any, B any] struct {
	a A
	b B
}

func (t *Tuple[A, B]) UnmarshalJSON(p []byte) error {
	var tmp []json.RawMessage
	if err := json.Unmarshal(p, &tmp); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[0], &t.a); err != nil {
		return err
	}

	if reflect.TypeFor[B]() == reflect.TypeFor[error]() {
		var s string
		if err := json.Unmarshal(tmp[1], &s); err != nil {
			return err
		}
		if s != "" {
			err := errors.New(s)
			t.b = err.(B)
		}
	} else {
		if err := json.Unmarshal(tmp[1], &t.b); err != nil {
			return err
		}
	}
	return nil
}

func newHTTPBinMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/redirect-to", func(w http.ResponseWriter, r *http.Request) {
		if urlParam := r.URL.Query().Get("url"); urlParam != "" {
			status := 302
			if codeParam := r.URL.Query().Get("status_code"); codeParam != "" {
				if c, err := strconv.ParseInt(codeParam, 10, 0); err == nil {
					status = int(c)
				}
			}
			fmt.Println(urlParam, status)
			http.Redirect(w, r, urlParam, status)
			return
		}
		http.NotFound(w, r)
	})

	mux.HandleFunc("/redirect/{times}", func(w http.ResponseWriter, r *http.Request) {
		urlParam := r.URL.Query().Get("url")
		status := 302
		if codeParam := r.URL.Query().Get("status_code"); codeParam != "" {
			if c, err := strconv.ParseInt(codeParam, 10, 0); err == nil {
				status = int(c)
			}
		}

		counter := 0
		if times := r.PathValue("times"); times != "" {
			fmt.Println("times", times)
			if v, err := strconv.ParseInt(times, 10, 0); err == nil {
				counter = int(v)
			}
		}
		fmt.Println("counter before", counter)
		counter--
		fmt.Println("counter after", counter)

		redirectTo := fmt.Sprintf("/redirect/%d?%s", counter, r.URL.Query().Encode())
		if counter <= 0 {
			redirectTo = urlParam
		}
		http.Redirect(w, r, redirectTo, status)
	})

	mux.HandleFunc("/response-headers", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.RawQuery)
		// do purposefully bad things to smuggle incorrect formatting
		// through go's safer query param parser
		rawQuery := strings.ReplaceAll(r.URL.RawQuery, ";", "%3B")

		params, err := url.ParseQuery(rawQuery)
		if err != nil {
			fmt.Println("qp err", err)
			rawQuery = strings.ReplaceAll(rawQuery, "%", "_")
			params, err = url.ParseQuery(rawQuery)
			if err != nil {
				fmt.Println("qp err2", err)
				w.WriteHeader(503)
				return
			}
		}

		fmt.Println(params)
		for k, v := range params {
			for i := range v {
				val := v[i]
				if strings.Contains(val, "_") {
					val = strings.ReplaceAll(val, "_", "%")
				}
				w.Header().Add(k, val)
			}
		}
		fmt.Println(w.Header())
		w.WriteHeader(200)
	})

	return mux
}

func onlyAllowHTTPTestServer(tsURL string) func(net.IP, string) bool {
	return func(i net.IP, s string) bool {
		u, err := url.Parse(tsURL)
		if err != nil {
			fmt.Println(err)
		}
		ips, err := hostnameToIPs(u.Hostname())
		if err != nil {
			fmt.Println(err)
		}
		for _, ip := range ips {
			if ip.Equal(i) && u.Port() == s {
				return true
			}
		}
		return false
	}
}

type ltsPair struct {
	pattern string
	handler http.HandlerFunc
}

func newLtsHandler(pattern string, handler http.HandlerFunc) ltsPair {
	return ltsPair{
		pattern, handler,
	}
}

func newLoopbackTestServer(c Config, handlers ...ltsPair) (Config, *httptest.Server) {
	mux := newHTTPBinMux()
	// handlers are optional, but by convention only accepts a single param
	for _, pair := range handlers {
		mux.HandleFunc(pair.pattern, pair.handler)
	}

	config := c
	config.insecureTestMode = true
	ts := httptest.NewServer(mux)
	config.insecureTestModeAllow = onlyAllowHTTPTestServer(ts.URL)
	return config, ts
}
