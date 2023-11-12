// Copyright (c) 2012-2023 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package camo

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cactus/go-camo/v2/pkg/camo/encoding"
	"github.com/cactus/go-camo/v2/pkg/router"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func makeReq(config Config, testURL string) (*http.Request, error) {
	k := []byte(config.HMACKey)
	hexURL := encoding.B64EncodeURL(k, testURL)
	out := "http://example.com" + hexURL
	req, err := http.NewRequest("GET", out, nil)
	if err != nil {
		return nil, fmt.Errorf("Error building req url '%s': %s", testURL, err.Error())
	}
	return req, nil
}

func processRequest(req *http.Request, status int, camoConfig Config, filters []FilterFunc) (*http.Response, error) {
	var (
		camoServer *Proxy
		err        error
	)

	if len(filters) == 0 {
		camoServer, err = New(camoConfig)
		if err != nil {
			return nil, fmt.Errorf("Error building Camo: %s", err.Error())
		}
	} else {
		camoServer, err = NewWithFilters(camoConfig, filters)
		if err != nil {
			return nil, fmt.Errorf("Error building Camo: %s", err.Error())
		}
	}

	router := &router.DumbRouter{
		AddHeaders:  map[string]string{"X-Go-Camo": "test"},
		ServerName:  camoConfig.ServerName,
		CamoHandler: camoServer,
	}

	record := httptest.NewRecorder()
	router.ServeHTTP(record, req)
	resp := record.Result()
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

func bodyAssert(t *testing.T, expected string, resp *http.Response) {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	assert.Check(t, err)
	bodyString := string(body)
	assert.Check(t, is.Equal(expected, bodyString),
		"Expected 404 response body but got '%s' instead",
		bodyString,
	)
}

func headerAssert(t *testing.T, expected, name string, resp *http.Response) {
	t.Helper()
	assert.Check(t,
		is.Equal(expected, resp.Header.Get(name)),
		"Expected response header mismatch",
	)
}

func statusCodeAssert(t *testing.T, expected int, resp *http.Response) {
	t.Helper()
	assert.Check(t,
		is.Equal(expected, resp.StatusCode),
		"Expected %d but got '%d' instead",
		expected, resp.StatusCode,
	)
}
