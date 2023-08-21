// Copyright (c) 2012-2019 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package encoding

import (
	"crypto/hmac"
	"crypto/sha1" // #nosec G505 -- used for hmac only
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/cactus/mlog"
)

// DecoderFunc is a function type that defines a url decoder.
type DecoderFunc func([]byte, string, string, string) (string, SimpleHeader, error)

// EncoderFunc is a function type that defines a url encoder.
type EncoderFunc func([]byte, string, SimpleHeader) (string, error)

type SimpleHeader map[string]string
type Codec int
type encodeFunc func([]byte) string
type decodeFunc func(string) ([]byte, error)

const (
	B64Codec Codec = iota
	HexCodec
)

func (codec Codec) String() string {
	switch codec {
	case B64Codec:
		return "B64Codec"
	case HexCodec:
		return "HexCodec"
	}
	return "unknown"
}

func validateUrlMac(hmackey []byte, macBytes []byte, urlBytes []byte, extraHdrBytes []byte) error {
	mac := hmac.New(sha1.New, hmackey)
	mac.Write(urlBytes) // #nosec G104 -- doesn't apply to hmac
	if len(extraHdrBytes) != 0 {
		mac.Write(extraHdrBytes) // #nosec G104 -- doesn't apply to hmac
	}
	macSum := mac.Sum(nil)

	// ensure lengths are equal. if not, return false
	if len(macSum) != len(macBytes) {
		return fmt.Errorf("mismatched length")
	}

	if subtle.ConstantTimeCompare(macSum, macBytes) != 1 {
		return fmt.Errorf("invalid mac")
	}
	return nil
}

func generateUrlMac(hmacKey []byte, oBytes []byte, oExtraHdrBytes []byte) []byte {
	mac := hmac.New(sha1.New, hmacKey)
	mac.Write(oBytes) // #nosec G104 -- doesn't apply to hmac
	if len(oExtraHdrBytes) > 0 {
		mac.Write(oExtraHdrBytes)
	}
	return mac.Sum(nil)
}

func extraHdrUnmarshal(extraHdrBytes []byte) (SimpleHeader, error) {
	var headers SimpleHeader
	err := json.Unmarshal(extraHdrBytes, &headers)
	if err != nil {
		return nil, err
	}
	return headers, nil
}

func extraHdrMarshal(headers SimpleHeader) ([]byte, error) {
	headerBytes, err := json.Marshal(headers)
	if err != nil {
		return nil, err
	}
	return headerBytes, nil
}

func decodeURL(decode decodeFunc, hmackey []byte, encDig, encURL, encExtraHdr string) (string, SimpleHeader, error) {
	macBytes, err := decode(encDig)
	if err != nil {
		return "", nil, fmt.Errorf("bad mac decode")
	}

	urlBytes, err := decode(encURL)
	if err != nil {
		return "", nil, fmt.Errorf("bad url decode")
	}

	var extraHdrBytes []byte
	if len(encExtraHdr) > 0 {
		var err error
		extraHdrBytes, err = decode(encExtraHdr)
		if err != nil {
			return "", nil, fmt.Errorf("bad additional data decode")
		}
	}

	if err := validateUrlMac(hmackey, macBytes, urlBytes, extraHdrBytes); err != nil {
		return "", nil, fmt.Errorf("invalid signature: %s", err)
	}

	// !!only unmarshal _after_ hmac validation!!
	var extraHdr SimpleHeader
	if len(extraHdrBytes) > 0 {
		var err error
		// !!only unmarshal _after_ hmac validation!!
		extraHdr, err = extraHdrUnmarshal(extraHdrBytes)
		if err != nil {
			return "", nil, fmt.Errorf("bad additional data unmarshal")
		}
	}
	return string(urlBytes), extraHdr, nil
}

// DecodeURL ensures the url is properly verified via HMAC, and then
// unencodes the url, returning the url (if valid) and whether the
// HMAC was verified. Tries either HexDecode or B64Decode, depending on the
// length of the encoded hmac.
func DecodeURL(hmackey []byte, encDig, encURL, encExtraHdr string) (string, SimpleHeader, bool) {
	var decode decodeFunc
	if len(encDig) == 40 {
		decode = hex.DecodeString
	} else {
		decode = base64.RawURLEncoding.DecodeString
	}

	urlBytes, extraHdr, err := decodeURL(decode, hmackey, encDig, encURL, encExtraHdr)
	if err != nil {
		if mlog.HasDebug() {
			mlog.Debugf("Bad Decode of URL: %s", err)
		}
		return "", nil, false
	}
	return urlBytes, extraHdr, true
}

func EncodeURL(codec Codec, hmacKey []byte, oURL string, oExtraHdr SimpleHeader) (string, error) {
	oURLBytes := []byte(oURL)

	var encFunc encodeFunc
	switch codec {
	case B64Codec:
		encFunc = base64.RawURLEncoding.EncodeToString
	case HexCodec:
		encFunc = hex.EncodeToString
	default:
		return "", fmt.Errorf("bad encoded specified")
	}

	var oExtraHdrBytes []byte
	if len(oExtraHdr) > 0 {
		var err error
		oExtraHdrBytes, err = extraHdrMarshal(oExtraHdr)
		if err != nil {
			return "", err
		}
	}

	urlMacSum := generateUrlMac(hmacKey, oURLBytes, oExtraHdrBytes)
	encURL := "/" + encFunc(urlMacSum) + "/" + encFunc(oURLBytes)
	if len(oExtraHdrBytes) > 0 {
		encURL = encURL + "/" + encFunc(oExtraHdrBytes)
	}
	return encURL, nil
}
