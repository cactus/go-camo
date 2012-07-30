package camoproxy

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
)

// DecodeUrl ensures the url is properly verified via HMAC, and then
// unencodes the url, returning the url (if valid) and whether the
// HMAC was verified.
func DecodeUrl(hmackey *[]byte, hexdig string, hexurl string) (surl string, valid bool) {
	urlBytes, err := hex.DecodeString(hexurl)
	if err != nil {
		Logger.Debugln("Bad Hex Decode", hexurl)
		return
	}
	surl = string(urlBytes)
	mac := hmac.New(sha1.New, *hmackey)
	mac.Write([]byte(surl))
	macSum := hex.EncodeToString(mac.Sum([]byte{}))
	if macSum != hexdig {
		Logger.Debugf("Bad signature: %s != %s\n", macSum, hexdig)
		return
	}
	valid = true
	return
}

// EncodeUrl takes an HMAC key and a url, and returns url
// path partial consisitent of signature and encoded url.
func EncodeUrl(hmacKey *[]byte, oUrl string) string {
	mac := hmac.New(sha1.New, *hmacKey)
	mac.Write([]byte(oUrl))
	macSum := hex.EncodeToString(mac.Sum([]byte{}))
	encodedUrl := hex.EncodeToString([]byte(oUrl))
	hexurl := "/" + macSum + "/" + encodedUrl
	return hexurl
}
