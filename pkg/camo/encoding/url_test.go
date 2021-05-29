// Copyright (c) 2012-2019 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package encoding

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

type enctesto struct {
	encoder EncoderFunc
	hmac    string
	edig    string
	eURL    string
	sURL    string
}

var enctests = []enctesto{
	// hex
	{HexEncodeURL, "test", "0f6def1cb147b0e84f39cbddc5ea10c80253a6f3",
		"687474703a2f2f676f6c616e672e6f72672f646f632f676f706865722f66726f6e74706167652e706e67",
		"http://golang.org/doc/gopher/frontpage.png"},

	// base64
	{B64EncodeURL, "test", "D23vHLFHsOhPOcvdxeoQyAJTpvM",
		"aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n",
		"http://golang.org/doc/gopher/frontpage.png"},
}

func TestEncoder(t *testing.T) {
	t.Parallel()
	for _, p := range enctests {
		hmacKey := []byte(p.hmac)
		// test specific encoder
		encodedURL := p.encoder(hmacKey, p.sURL)
		assert.Check(t, is.Equal(encodedURL, fmt.Sprintf("/%s/%s", p.edig, p.eURL)), "encoded url does not match")
	}
}

type dectesto struct {
	decoder DecoderFunc
	hmac    string
	edig    string
	eURL    string
	sURL    string
}

var dectests = []dectesto{
	// hex
	{HexDecodeURL, "test", "0f6def1cb147b0e84f39cbddc5ea10c80253a6f3",
		"687474703a2f2f676f6c616e672e6f72672f646f632f676f706865722f66726f6e74706167652e706e67",
		"http://golang.org/doc/gopher/frontpage.png"},

	// base64
	{B64DecodeURL, "test", "D23vHLFHsOhPOcvdxeoQyAJTpvM",
		"aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n",
		"http://golang.org/doc/gopher/frontpage.png"},
}

func TestDecoder(t *testing.T) {
	t.Parallel()
	for _, p := range dectests {
		hmacKey := []byte(p.hmac)
		// test specific decoder
		encodedURL, err := p.decoder(hmacKey, p.edig, p.eURL)
		assert.Check(t, err, "decoded url failed to verify")
		assert.Check(t, is.Equal(encodedURL, p.sURL), "decoded url does not match")

		// also test generic "guessing" decoder
		encodedURL, ok := DecodeURL(hmacKey, p.edig, p.eURL)
		assert.Check(t, ok, "decoded url failed to verify")
		assert.Check(t, is.Equal(encodedURL, p.sURL), "decoded url does not match")
	}
}

func BenchmarkHexEncoder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		HexEncodeURL([]byte("test"), "http://golang.org/doc/gopher/frontpage.png")
	}
}

func BenchmarkB64Encoder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		B64EncodeURL([]byte("test"), "http://golang.org/doc/gopher/frontpage.png")
	}
}

func BenchmarkHexDecoder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		HexDecodeURL([]byte("test"), "0f6def1cb147b0e84f39cbddc5ea10c80253a6f3", "687474703a2f2f676f6c616e672e6f72672f646f632f676f706865722f66726f6e74706167652e706e67")
	}
}

func BenchmarkB64Decoder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		B64DecodeURL([]byte("test"), "D23vHLFHsOhPOcvdxeoQyAJTpvM", "aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n")
	}
}

func BenchmarkGuessingDecoderHex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DecodeURL([]byte("test"), "0f6def1cb147b0e84f39cbddc5ea10c80253a6f3", "687474703a2f2f676f6c616e672e6f72672f646f632f676f706865722f66726f6e74706167652e706e67")
	}
}

func BenchmarkGuessingDecoderB64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DecodeURL([]byte("test"), "D23vHLFHsOhPOcvdxeoQyAJTpvM", "aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n")
	}
}

var baddectests = []dectesto{
	// hex
	{HexDecodeURL, "test", "000",
		"687474703a2f2f676f6c616e672e6f72672f646f632f676f706865722f66726f6e74706167652e706e67", ""},
	{HexDecodeURL, "test", "0f6def1cb147b0e84f39cbddc5ea10c80253a6f3",
		"000000000000000000000000000000000000000000000000000000000000000000000000000000000000", ""},

	// base64
	{B64DecodeURL, "test", "000",
		"aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n", ""},
	{B64DecodeURL, "test", "D23vHLFHsOhPOcvdxeoQyAJTpvM",
		"00000000000000000000000000000000000000000000000000000000", ""},

	// mixmatch
	// hex
	{HexDecodeURL, "test", "0f6def1cb147b0e84f39cbddc5ea10c80253a6f3",
		"aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n",
		"http://golang.org/doc/gopher/frontpage.png"},

	// base64
	{B64DecodeURL, "test", "D23vHLFHsOhPOcvdxeoQyAJTpvM",
		"687474703a2f2f676f6c616e672e6f72672f646f632f676f706865722f66726f6e74706167652e706e67",
		"http://golang.org/doc/gopher/frontpage.png"},
}

func TestBadDecodes(t *testing.T) {
	t.Parallel()
	for _, p := range baddectests {
		hmacKey := []byte(p.hmac)
		// test specific decoder
		encodedURL, err := p.decoder(hmacKey, p.edig, p.eURL)
		assert.Check(t, err != nil, "decoded url verfied when it shouldn't have")
		assert.Check(t, is.Equal(encodedURL, ""), "decoded url result not empty")

		// also test generic "guessing" decoder
		encodedURL, ok := DecodeURL(hmacKey, p.edig, p.eURL)
		assert.Check(t, !ok, "decoded url verfied when it shouldn't have")
		assert.Check(t, is.Equal(encodedURL, ""), "decoded url result not empty")
	}
}
