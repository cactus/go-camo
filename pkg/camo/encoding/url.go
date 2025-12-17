// Copyright (c) 2012-2023 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package encoding

import (
	"crypto/hmac"
	"crypto/sha1" // #nosec G505 -- used for hmac only
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"codeberg.org/dropwhile/mlog"
)

// DecoderFunc is a function type that defines a url decoder.
type DecoderFunc func([]byte, string, string) (string, error)

// EncoderFunc is a function type that defines a url encoder.
type EncoderFunc func([]byte, string) string

func validateURL(hmackey *[]byte, macbytes *[]byte, urlbytes *[]byte) error {
	mac := hmac.New(sha1.New, *hmackey)
	mac.Write(*urlbytes) // #nosec G104 -- doesn't apply to hmac
	macSum := mac.Sum(nil)

	// ensure lengths are equal. if not, return false
	if len(macSum) != len(*macbytes) {
		return fmt.Errorf("mismatched length")
	}

	if subtle.ConstantTimeCompare(macSum, *macbytes) != 1 {
		return fmt.Errorf("invalid mac")
	}
	return nil
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
func HexDecodeURL(hmackey []byte, hexdig string, hexURL string) (string, error) {
	urlBytes, err := hex.DecodeString(hexURL)
	if err != nil {
		return "", fmt.Errorf("bad url decode")
	}
	macBytes, err := hex.DecodeString(hexdig)
	if err != nil {
		return "", fmt.Errorf("bad mac decode")
	}

	if err = validateURL(&hmackey, &macBytes, &urlBytes); err != nil {
		return "", fmt.Errorf("invalid signature: %s", err)
	}
	return string(urlBytes), nil
}

// HexEncodeURL takes an HMAC key and a url, and returns url
// path partial consisitent of signature and encoded url.
func HexEncodeURL(hmacKey []byte, oURL string) string {
	oBytes := []byte(oURL)
	mac := hmac.New(sha1.New, hmacKey)
	mac.Write(oBytes) // #nosec G104 -- doesn't apply to hmac
	macSum := hex.EncodeToString(mac.Sum(nil))
	encodedURL := hex.EncodeToString(oBytes)
	hexURL := "/" + macSum + "/" + encodedURL
	return hexURL
}

// B64DecodeURL ensures the url is properly verified via HMAC, and then
// unencodes the url, returning the url (if valid) and whether the
// HMAC was verified.
func B64DecodeURL(hmackey []byte, encdig string, encURL string) (string, error) {
	urlBytes, err := b64decode(encURL)
	if err != nil {
		return "", fmt.Errorf("bad url decode")
	}
	macBytes, err := b64decode(encdig)
	if err != nil {
		return "", fmt.Errorf("bad mac decode")
	}

	if err := validateURL(&hmackey, &macBytes, &urlBytes); err != nil {
		return "", fmt.Errorf("invalid signature: %s", err)
	}
	return string(urlBytes), nil
}

// B64EncodeURL takes an HMAC key and a url, and returns url
// path partial consisitent of signature and encoded url.
func B64EncodeURL(hmacKey []byte, oURL string) string {
	oBytes := []byte(oURL)
	mac := hmac.New(sha1.New, hmacKey)
	mac.Write(oBytes) // #nosec G104 -- doesn't apply to hmac
	macSum := b64encode(mac.Sum(nil))
	encodedURL := b64encode(oBytes)
	encURL := "/" + macSum + "/" + encodedURL
	return encURL
}

// DecodeURL ensures the url is properly verified via HMAC, and then
// unencodes the url, returning the url (if valid) and whether the
// HMAC was verified. Tries either HexDecode or B64Decode, depending on the
// length of the encoded hmac.
func DecodeURL(hmackey []byte, encdig string, encURL string) (string, bool) {
	var decoder DecoderFunc
	if len(encdig) == 40 {
		decoder = HexDecodeURL
	} else {
		decoder = B64DecodeURL
	}

	urlBytes, err := decoder(hmackey, encdig, encURL)
	if err != nil {
		if mlog.HasDebug() {
			mlog.Debugf("Bad Decode of URL: %s", err)
		}
		return "", false
	}
	return urlBytes, true
}
