// Copyright (c) 2012-2019 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package camo provides an HTTP proxy server with content type
// restrictions as well as regex host allow list support.
package camo

import (
	"net/http"

	"github.com/cactus/mlog"
)

func httpReqToMlogMap(req *http.Request) mlog.Map {
	return mlog.Map{
		"method":            req.Method,
		"path":              req.RequestURI,
		"proto":             req.Proto,
		"header":            req.Header,
		"content_length":    req.ContentLength,
		"transfer_encoding": req.TransferEncoding,
		"host":              req.Host,
		"remote_addr":       req.RemoteAddr,
	}
}

func httpRespToMlogMap(resp *http.Response) mlog.Map {
	return mlog.Map{
		"status":            resp.StatusCode,
		"proto":             resp.Proto,
		"header":            resp.Header,
		"content_length":    resp.ContentLength,
		"transfer_encoding": resp.TransferEncoding,
	}
}
