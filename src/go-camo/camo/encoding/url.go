// Copyright (c) 2012-2016 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package encoding

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"strings"

	"github.com/cactus/gologit"
)

func validateURL(hmackey *[]byte, macbytes *[]byte, urlbytes *[]byte) bool {
	mac := hmac.New(sha1.New, *hmackey)
	mac.Write(*urlbytes)
	macSum := mac.Sum(nil)

	// ensure lengths are equal. if not, return false
	if len(macSum) != len(*macbytes) {
		gologit.Debugf("Bad signature: %x != %x\n", macSum, macbytes)
		return false
	}

	if subtle.ConstantTimeCompare(macSum, *macbytes) != 1 {
		gologit.Debugf("Bad signature: %x != %x\n", macSum, macbytes)
		return false
	}
	return true
}

func b64encode(data []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(data), "=")
}

func b64decode(str string) ([]byte, error) {
	padChars := (4 - (len(str) % 4)) % 4
	for i := 0; i < padChars; i++ {
		str = str + "="
	}
	decBytes, ok := base64.URLEncoding.DecodeString(str)
	return decBytes, ok
}

// HexDecodeURL ensures the url is properly verified via HMAC, and then
// unencodes the url, returning the url (if valid) and whether the
// HMAC was verified.
func HexDecodeURL(hmackey []byte, hexdig string, hexURL string) (string, bool) {
	urlBytes, err := hex.DecodeString(hexURL)
	if err != nil {
		gologit.Debugln("Bad Hex Decode of URL", hexURL)
		return "", false
	}
	macBytes, err := hex.DecodeString(hexdig)
	if err != nil {
		gologit.Debugln("Bad Hex Decode of MAC", hexURL)
		return "", false
	}

	if ok := validateURL(&hmackey, &macBytes, &urlBytes); !ok {
		return "", false
	}
	return string(urlBytes), true
}

// HexEncodeURL takes an HMAC key and a url, and returns url
// path partial consisitent of signature and encoded url.
func HexEncodeURL(hmacKey []byte, oURL string) string {
	oBytes := []byte(oURL)
	mac := hmac.New(sha1.New, hmacKey)
	mac.Write(oBytes)
	macSum := hex.EncodeToString(mac.Sum(nil))
	encodedURL := hex.EncodeToString(oBytes)
	hexURL := "/" + macSum + "/" + encodedURL
	return hexURL
}

// B64DecodeURL ensures the url is properly verified via HMAC, and then
// unencodes the url, returning the url (if valid) and whether the
// HMAC was verified.
func B64DecodeURL(hmackey []byte, encdig string, encURL string) (string, bool) {
	urlBytes, err := b64decode(encURL)
	if err != nil {
		gologit.Debugln("Bad B64 Decode of URL", encURL)
		return "", false
	}
	macBytes, err := b64decode(encdig)
	if err != nil {
		gologit.Debugln("Bad B64 Decode of MAC", encURL)
		return "", false
	}

	if ok := validateURL(&hmackey, &macBytes, &urlBytes); !ok {
		return "", false
	}
	return string(urlBytes), true
}

// B64EncodeURL takes an HMAC key and a url, and returns url
// path partial consisitent of signature and encoded url.
func B64EncodeURL(hmacKey []byte, oURL string) string {
	oBytes := []byte(oURL)
	mac := hmac.New(sha1.New, hmacKey)
	mac.Write(oBytes)
	macSum := b64encode(mac.Sum(nil))
	encodedURL := b64encode(oBytes)
	encURL := "/" + macSum + "/" + encodedURL
	return encURL
}

// DecodeURL ensures the url is properly verified via HMAC, and then
// unencodes the url, returning the url (if valid) and whether the
// HMAC was verified. Tries to HexDecode the url, then B64Decode if that fails.
func DecodeURL(hmackey []byte, encdig string, encURL string) (string, bool) {
	var decoder func([]byte, string, string) (string, bool)
	if len(encdig) == 40 {
		decoder = HexDecodeURL
	} else {
		decoder = B64DecodeURL
	}

	urlBytes, ok := decoder(hmackey, encdig, encURL)
	if !ok {
		gologit.Debugln("Bad Decode of URL", encURL)
		return "", false
	}
	return string(urlBytes), true
}
