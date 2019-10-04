// Copyright (c) 2012-2019 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
)

var CAMO_HOST = "https://img.example.com"

func GenCamoUrl(hmacKey []byte, srcUrl string) string {
	if strings.HasPrefix(srcUrl, "https:") {
		return srcUrl
	}
	oBytes := []byte(srcUrl)
	mac := hmac.New(sha1.New, hmacKey)
	mac.Write(oBytes)
	macSum := hex.EncodeToString(mac.Sum(nil))
	encodedUrl := hex.EncodeToString(oBytes)
	hexurl := CAMO_HOST + "/" + macSum + "/" + encodedUrl
	return hexurl
}

func main() {
	fmt.Println(GenCamoUrl([]byte("test"), "http://golang.org/doc/gopher/frontpage.png"))
}
