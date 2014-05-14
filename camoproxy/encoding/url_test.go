package encoding

import (
	"testing"
	"fmt"
)


type enctesto struct {
	encoder func(hmacKey []byte, oUrl string) string
	hmac, edig, eurl, surl string
}

var enctests = []enctesto{
	// hex
	{HexEncodeUrl, "test", "0f6def1cb147b0e84f39cbddc5ea10c80253a6f3",
	 "687474703a2f2f676f6c616e672e6f72672f646f632f676f706865722f66726f6e74706167652e706e67",
	 "http://golang.org/doc/gopher/frontpage.png"},

	// base64
	{B64EncodeUrl, "test", "D23vHLFHsOhPOcvdxeoQyAJTpvM",
	 "aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n",
	 "http://golang.org/doc/gopher/frontpage.png"},
}

func TestEncoder(t *testing.T) {
	t.Parallel()
	for _, p := range enctests {
		hmacKey := []byte(p.hmac)
		// test specific encoder
		encodedUrl := p.encoder(hmacKey, p.surl)
		if encodedUrl != fmt.Sprintf("/%s/%s", p.edig, p.eurl) {
			t.Error("encoded url does not match")
		}
	}
}

type dectesto struct {
	decoder func(hmackey []byte, encdig string, encurl string) (string, bool)
	hmac, edig, eurl, surl string
}

var dectests = []dectesto{
	// hex
	{HexDecodeUrl, "test", "0f6def1cb147b0e84f39cbddc5ea10c80253a6f3",
	 "687474703a2f2f676f6c616e672e6f72672f646f632f676f706865722f66726f6e74706167652e706e67",
	 "http://golang.org/doc/gopher/frontpage.png"},

	// base64
	{B64DecodeUrl, "test", "D23vHLFHsOhPOcvdxeoQyAJTpvM",
	 "aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n",
	 "http://golang.org/doc/gopher/frontpage.png"},
}

func TestDecoder(t *testing.T) {
	t.Parallel()
	for _, p := range dectests {
		hmacKey := []byte(p.hmac)
		// test specific decoder
		encodedUrl, ok := p.decoder(hmacKey, p.edig, p.eurl)
		if !ok {
			t.Error("decoded url failed to verify")
		}
		if encodedUrl != p.surl {
			t.Error("decoded url does not match")
		}
		// also test generic "guessing" decoder
		encodedUrl, ok = DecodeUrl(hmacKey, p.edig, p.eurl)
		if !ok {
			t.Error("decoded url failed to verify")
		}
		if encodedUrl != p.surl {
			t.Error("decoded url does not match")
		}
	}
}

var baddectests = []dectesto{
	// hex
	{HexDecodeUrl, "test", "000",
	 "687474703a2f2f676f6c616e672e6f72672f646f632f676f706865722f66726f6e74706167652e706e67", ""},
	{HexDecodeUrl, "test", "0f6def1cb147b0e84f39cbddc5ea10c80253a6f3",
	 "000000000000000000000000000000000000000000000000000000000000000000000000000000000000", ""},

	// base64
	{B64DecodeUrl, "test", "000",
	 "aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n", ""},
	{B64DecodeUrl, "test", "D23vHLFHsOhPOcvdxeoQyAJTpvM",
	 "00000000000000000000000000000000000000000000000000000000", ""},

	// mixmatch
	// hex
	{HexDecodeUrl, "test", "0f6def1cb147b0e84f39cbddc5ea10c80253a6f3",
	 "aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n",
	 "http://golang.org/doc/gopher/frontpage.png"},

	// base64
	{B64DecodeUrl, "test", "D23vHLFHsOhPOcvdxeoQyAJTpvM",
	 "687474703a2f2f676f6c616e672e6f72672f646f632f676f706865722f66726f6e74706167652e706e67",
	 "http://golang.org/doc/gopher/frontpage.png"},
}

func TestBadDecodes(t *testing.T) {
	t.Parallel()
	for _, p := range baddectests {
		hmacKey := []byte(p.hmac)
		// test specific decoder
		encodedUrl, ok := p.decoder(hmacKey, p.edig, p.eurl)
		if ok {
			t.Error("decoded url failed to verify")
		}
		if encodedUrl != "" {
			t.Error("decoded url result not empty")
		}
		// also test generic "guessing" decoder
		encodedUrl, ok = DecodeUrl(hmacKey, p.edig, p.eurl)
		if ok {
			t.Error("decoded url failed to verify")
		}
		if encodedUrl != "" {
			t.Error("decoded url result not empty")
		}
	}
}
