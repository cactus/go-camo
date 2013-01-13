package encoding

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/hex"
	"github.com/cactus/gologit"
)

// DecodeUrl ensures the url is properly verified via HMAC, and then
// unencodes the url, returning the url (if valid) and whether the
// HMAC was verified.
func DecodeUrl(hmackey *[]byte, hexdig string, hexurl string) (string, bool) {
	urlBytes, err := hex.DecodeString(hexurl)
	if err != nil {
		gologit.Debugln("Bad Hex Decode of URL", hexurl)
		return "", false
	}
	macBytes, err := hex.DecodeString(hexdig)
	if err != nil {
		gologit.Debugln("Bad Hex Decode of MAC", hexurl)
		return "", false
	}

	mac := hmac.New(sha1.New, *hmackey)
	mac.Write(urlBytes)
	macSum := mac.Sum(nil)
	if subtle.ConstantTimeCompare(macSum, macBytes) != 1 {
		gologit.Debugf("Bad signature: %x != %x\n", macSum, macBytes)
		return "", false
	}
	return string(urlBytes), true
}

// EncodeUrl takes an HMAC key and a url, and returns url
// path partial consisitent of signature and encoded url.
func EncodeUrl(hmacKey *[]byte, oUrl string) string {
	oBytes := []byte(oUrl)
	mac := hmac.New(sha1.New, *hmacKey)
	mac.Write(oBytes)
	macSum := hex.EncodeToString(mac.Sum(nil))
	encodedUrl := hex.EncodeToString(oBytes)
	hexurl := "/" + macSum + "/" + encodedUrl
	return hexurl
}
