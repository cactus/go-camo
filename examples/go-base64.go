// Copyright (c) 2012-2019 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

//go:build ignore
// +build ignore

package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

var CAMO_HOST = "https://img.example.com"

func wrapEncode(data []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(data), "=")
}

func GenCamoUrl(hmacKey []byte, srcUrl string, extraHeaders map[string]string) string {
	if strings.HasPrefix(srcUrl, "https:") {
		return srcUrl
	}

	hasExtraHeaders := false
	if len(extraHeaders) > 0 {
		hasExtraHeaders = true

	}

	oBytes := []byte(srcUrl)
	mac := hmac.New(sha1.New, hmacKey)
	mac.Write(oBytes)

	encodedExtraHeaders := ""
	if hasExtraHeaders {
		extraHeaderBytes, _ := json.Marshal(extraHeaders)
		mac.Write(extraHeaderBytes)
		encodedExtraHeaders = wrapEncode(extraHeaderBytes)
	}

	macSum := wrapEncode(mac.Sum(nil))
	encodedUrl := wrapEncode(oBytes)
	encurl := CAMO_HOST + "/" + macSum + "/" + encodedUrl
	if hasExtraHeaders {
		encurl = encurl + "/" + encodedExtraHeaders
	}
	return encurl
}

func main() {
	fmt.Println(GenCamoUrl([]byte("test"), "http://golang.org/doc/gopher/frontpage.png", nil))
	// https://img.example.com/D23vHLFHsOhPOcvdxeoQyAJTpvM/aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n

	fmt.Println(
		GenCamoUrl(
			[]byte("test"),
			"http://golang.org/doc/gopher/frontpage.png",
			map[string]string{
				"content-disposition": "attachment; filename=\"image.png\"",
			},
		),
	)
	// https://img.example.com/uLOqbvq5Esc9kQunfCd8HReDR40/aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n/eyJjb250ZW50LWRpc3Bvc2l0aW9uIjpbImF0dGFjaG1lbnQ7IGZpbGVuYW1lPVwiaW1hZ2UucG5nXCIiXX0
}
