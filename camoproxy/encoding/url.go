package encoding

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/hex"
	"encoding/base64"
	"github.com/cactus/gologit"
	"strings"
)

func validateUrl(hmackey *[]byte, macbytes *[]byte, urlbytes *[]byte) bool {
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

// HexDecodeUrl ensures the url is properly verified via HMAC, and then
// unencodes the url, returning the url (if valid) and whether the
// HMAC was verified.
func HexDecodeUrl(hmackey []byte, hexdig string, hexurl string) (string, bool) {
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

	if ok := validateUrl(&hmackey, &macBytes, &urlBytes); !ok {
		return "", false
	}
	return string(urlBytes), true
}

// HexEncodeUrl takes an HMAC key and a url, and returns url
// path partial consisitent of signature and encoded url.
func HexEncodeUrl(hmacKey []byte, oUrl string) string {
	oBytes := []byte(oUrl)
	mac := hmac.New(sha1.New, hmacKey)
	mac.Write(oBytes)
	macSum := hex.EncodeToString(mac.Sum(nil))
	encodedUrl := hex.EncodeToString(oBytes)
	hexurl := "/" + macSum + "/" + encodedUrl
	return hexurl
}

// B64DecodeUrl ensures the url is properly verified via HMAC, and then
// unencodes the url, returning the url (if valid) and whether the
// HMAC was verified.
func B64DecodeUrl(hmackey []byte, encdig string, encurl string) (string, bool) {
	urlBytes, err := b64decode(encurl)
	if err != nil {
		gologit.Debugln("Bad B64 Decode of URL", encurl)
		return "", false
	}
	macBytes, err := b64decode(encdig)
	if err != nil {
		gologit.Debugln("Bad B64 Decode of MAC", encurl)
		return "", false
	}

	if ok := validateUrl(&hmackey, &macBytes, &urlBytes); !ok {
		return "", false
	}
	return string(urlBytes), true
}

// B64EncodeUrl takes an HMAC key and a url, and returns url
// path partial consisitent of signature and encoded url.
func B64EncodeUrl(hmacKey []byte, oUrl string) string {
	oBytes := []byte(oUrl)
	mac := hmac.New(sha1.New, hmacKey)
	mac.Write(oBytes)
	macSum := b64encode(mac.Sum(nil))
	encodedUrl := b64encode(oBytes)
	encurl := "/" + macSum + "/" + encodedUrl
	return encurl
}


// DecodeUrl ensures the url is properly verified via HMAC, and then
// unencodes the url, returning the url (if valid) and whether the
// HMAC was verified. Tries to HexDecode the url, then B64Decode if that fails.
func DecodeUrl(hmackey []byte, encdig string, encurl string) (string, bool) {
	var decoder func([]byte, string, string) (string, bool)
	if len(encdig) == 40 {
		decoder = HexDecodeUrl
	} else {
		decoder = B64DecodeUrl
	}

	urlBytes, ok := decoder(hmackey, encdig, encurl)
	if !ok {
		gologit.Debugln("Bad Decode of URL", encurl)
		return "", false
	}
	return string(urlBytes), true
}
