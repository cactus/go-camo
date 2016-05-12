// Copyright (c) 2012-2016 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package camo

import (
	"regexp"
)

// Headers that are acceptable to pass from the client to the remote
// server. Only those present and true, are forwarded. Empty implies
// no filtering.
var ValidReqHeaders = map[string]bool{
	"Accept":         true,
	"Accept-Charset": true,
	// images (aside from xml/svg), don't generally benefit (generally) from
	// compression
	"Accept-Encoding":   false,
	"Accept-Language":   true,
	"Cache-Control":     true,
	"If-None-Match":     true,
	"If-Modified-Since": true,
	"X-Forwarded-For":   true,
}

// Headers that are acceptable to pass from the remote server to the
// client. Only those present and true, are forwarded. Empty implies
// no filtering.
var ValidRespHeaders = map[string]bool{
	// Do not offer to accept range requests
	"Accept-Ranges":     false,
	"Cache-Control":     true,
	"Content-Encoding":  true,
	"Content-Type":      true,
	"Transfer-Encoding": true,
	"Expires":           true,
	"Last-Modified":     true,
	// override in response with either nothing, or ServerNameVer
	"Server": false,
}

// addr1918PrefixMatch is a regex for matching the prefix of hosts in
// x-forward-for header filtering for rfc1918 addresses
var addr1918PrefixRegex = regexp.MustCompile(`^(127\.|10\.|169\.254|192\.168|172\.(?:(?:1[6-9])|(?:2[0-9])|(?:3[0-1])))`)

// match for localhost
var localhostRegex = regexp.MustCompile(`^localhost\.?(localdomain)?\.?$`)
