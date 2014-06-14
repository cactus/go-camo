// +build ignore
package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"strings"
)

var CAMO_HOST = "https://img.example.com"

func b64encode(data []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(data), "=")
}

func GenCamoUrl(hmacKey []byte, srcUrl string) string {
	if strings.HasPrefix(srcUrl, "https:") {
		return srcUrl
	}
	oBytes := []byte(srcUrl)
	mac := hmac.New(sha1.New, hmacKey)
	mac.Write(oBytes)
	macSum := b64encode(mac.Sum(nil))
	encodedUrl := b64encode(oBytes)
	encurl := CAMO_HOST + "/" + macSum + "/" + encodedUrl
	return encurl
}

func main() {
	fmt.Println(GenCamoUrl([]byte("test"), "http://golang.org/doc/gopher/frontpage.png"))
}
