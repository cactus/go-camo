// Copyright (c) 2015-2026 Eli Janssen
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.
//
// Inspiration from https://github.com/nalgeon/be

// Package asserthttp provides some simple http test functions
package asserthttp

import (
	"io"
	"net/http"

	"codeberg.org/dropwhile/go-droplibs/assert"
	"codeberg.org/dropwhile/go-droplibs/internal/asserty"
)

// Header fails assertion if the expected http header value is not equal.
func Header(t assert.TestingT, resp *http.Response, headerName string, want string, msg ...string) {
	if ht, ok := t.(asserty.HelperT); ok {
		ht.Helper()
	}

	got := resp.Header.Get(headerName)
	if !asserty.IsEqual(got, want) {
		t.Fatalf(
			"http header ('%s') mismatch. got %q; want %q;%s",
			headerName, got, want, asserty.FormatMsg(msg...),
		)
	}
}

// StatusCode fails assertion if the expected http status code is not equal.
func StatusCode(t assert.TestingT, resp *http.Response, want int, msg ...string) {
	if ht, ok := t.(asserty.HelperT); ok {
		ht.Helper()
	}

	if !asserty.IsEqual(resp.StatusCode, want) {
		t.Fatalf(
			"http status code mismatch. got '%d'; want '%d';%s",
			resp.StatusCode, want, asserty.FormatMsg(msg...),
		)
	}
}

// Body fails assertion if the expected http response body is not equal.
func Body(t assert.TestingT, resp *http.Response, want string, msg ...string) {
	if ht, ok := t.(asserty.HelperT); ok {
		ht.Helper()
	}

	body, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	bodyString := string(body)

	if !asserty.IsEqual(bodyString, want) {
		t.Fatalf(
			"http response body mismatch. got %q; want %q;%s",
			bodyString, want, asserty.FormatMsg(msg...),
		)
	}
}
