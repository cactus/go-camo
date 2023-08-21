// Copyright (c) 2012-2019 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package encoding

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

type enctesto struct {
	codec     Codec
	hmac      string
	edig      string
	eURL      string
	eExtraHdr string
	sURL      string
	sExtraHdr SimpleHeader
}

var goodTests = []enctesto{
	{
		HexCodec,
		"test", "0f6def1cb147b0e84f39cbddc5ea10c80253a6f3",
		"687474703a2f2f676f6c616e672e6f72672f646f632f676f706865722f66726f6e74706167652e706e67",
		"",
		"http://golang.org/doc/gopher/frontpage.png",
		nil,
	},
	{
		HexCodec,
		"test", "c35015f0abed9bf891d5a68d36597762b874dfd3",
		"687474703a2f2f676f6c616e672e6f72672f646f632f676f706865722f66726f6e74706167652e706e67",
		"7b2268656164657231223a2276616c756531227d",
		"http://golang.org/doc/gopher/frontpage.png",
		SimpleHeader{"header1": "value1"},
	},
	{
		HexCodec,
		"test", "ce99b055e7de84953ff13e01bc6f3bc3e730b5bd",
		"687474703a2f2f676f6c616e672e6f72672f646f632f676f706865722f66726f6e74706167652e706e67",
		"7b2268656164657231223a2276616c756531222c2268656164657232223a2276616c756532227d",
		"http://golang.org/doc/gopher/frontpage.png",
		SimpleHeader{"header1": "value1", "header2": "value2"},
	},

	{
		B64Codec,
		"test", "D23vHLFHsOhPOcvdxeoQyAJTpvM",
		"aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n",
		"",
		"http://golang.org/doc/gopher/frontpage.png",
		nil,
	},
	{
		B64Codec,
		"test", "w1AV8Kvtm_iR1aaNNll3Yrh039M",
		"aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n",
		"eyJoZWFkZXIxIjoidmFsdWUxIn0",
		"http://golang.org/doc/gopher/frontpage.png",
		SimpleHeader{"header1": "value1"},
	},
	{
		B64Codec,
		"test", "zpmwVefehJU_8T4BvG87w-cwtb0",
		"aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n",
		"eyJoZWFkZXIxIjoidmFsdWUxIiwiaGVhZGVyMiI6InZhbHVlMiJ9",
		"http://golang.org/doc/gopher/frontpage.png",
		SimpleHeader{"header1": "value1", "header2": "value2"},
	},
}

func TestEncoder(t *testing.T) {
	t.Parallel()
	for i, p := range goodTests {
		hmacKey := []byte(p.hmac)
		// test specific encoder
		encodedURL, err := EncodeURL(p.codec, hmacKey, p.sURL, p.sExtraHdr)
		assert.NilError(t, err)
		url := fmt.Sprintf("/%s/%s", p.edig, p.eURL)
		if p.eExtraHdr != "" {
			url = url + "/" + p.eExtraHdr
		}
		assert.Check(t, is.Equal(encodedURL, url), "%s[%d]: encoded url does not match", p.codec.String(), i)
	}
}

func TestDecoder(t *testing.T) {
	t.Parallel()
	for i, p := range goodTests {
		codecName := p.codec.String()
		hmacKey := []byte(p.hmac)
		// test specific decoder
		encodedURL, extraHdr, err := DecodeURL(hmacKey, p.edig, p.eURL, p.eExtraHdr)
		assert.Check(t, err, "%s[%d]: decoded url failed to verify", codecName, i)
		assert.Check(t, is.Equal(encodedURL, p.sURL), "%s[%d]: decoded url does not match", codecName, i)
		assert.Check(t, is.DeepEqual(extraHdr, p.sExtraHdr), "%s[%d]: decoded extraHdr does not match", codecName, i)

		// also test generic "guessing" decoder
		encodedURL, extraHdr, ok := DecodeURL(hmacKey, p.edig, p.eURL, p.eExtraHdr)
		assert.Check(t, ok, "%s[%d]: decoded url failed to verify", codecName, i)
		assert.Check(t, is.Equal(encodedURL, p.sURL), "%s[%d]: decoded url does not match", codecName, i)
		assert.Check(t, is.DeepEqual(extraHdr, p.sExtraHdr), "%s[%d]: decoded extraHdr does not match", codecName, i)
	}
}

func BenchmarkHexEncoder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		EncodeURL(HexCodec, []byte("test"), "http://golang.org/doc/gopher/frontpage.png", nil)
	}
}

func BenchmarkB64Encoder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		EncodeURL(B64Codec, []byte("test"), "http://golang.org/doc/gopher/frontpage.png", nil)
	}
}

func BenchmarkHexDecoder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		decodeURL(hex.DecodeString, []byte("test"), "0f6def1cb147b0e84f39cbddc5ea10c80253a6f3", "687474703a2f2f676f6c616e672e6f72672f646f632f676f706865722f66726f6e74706167652e706e67", "")
	}
}

func BenchmarkB64Decoder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		decodeURL(base64.RawURLEncoding.DecodeString, []byte("test"), "D23vHLFHsOhPOcvdxeoQyAJTpvM", "aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n", "")
	}
}

func BenchmarkHexGuessingDecoder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DecodeURL([]byte("test"), "0f6def1cb147b0e84f39cbddc5ea10c80253a6f3", "687474703a2f2f676f6c616e672e6f72672f646f632f676f706865722f66726f6e74706167652e706e67", "")
	}
}

func BenchmarkB64GuessingDecoder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DecodeURL([]byte("test"), "D23vHLFHsOhPOcvdxeoQyAJTpvM", "aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n", "")
	}
}

var badTests = []enctesto{
	// bad digest //
	{
		HexCodec,
		// hmac
		"test",
		// encoded digest
		"000",
		// encoded url - http://golang.org/doc/gopher/frontpage.png
		"687474703a2f2f676f6c616e672e6f72672f646f632f676f706865722f66726f6e74706167652e706e67",
		// encoded extraHdr
		"",
		// expected url
		"",
		// expected extraHdr
		nil,
	},
	{
		B64Codec,
		"test", "000",
		"aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n",
		"", "", nil,
	},

	// bad url encoding //
	{
		HexCodec,
		"test", "0f6def1cb147b0e84f39cbddc5ea10c80253a6f3",
		"000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
		"", "", nil,
	},
	{
		B64Codec,
		"test", "D23vHLFHsOhPOcvdxeoQyAJTpvM",
		"00000000000000000000000000000000000000000000000000000000",
		"", "", nil,
	},

	// bad extraHdr encoding //
	// bad extraHdr data (invalid json) //
	{
		HexCodec,
		"test", "0f6def1cb147b0e84f39cbddc5ea10c80253a6f3",
		"687474703a2f2f676f6c616e672e6f72672f646f632f676f706865722f66726f6e74706167652e706e67",
		"000000000000000000000000000000000000000000",
		"http://golang.org/doc/gopher/frontpage.png",
		SimpleHeader{"header1": "value1"},
	},
	{
		B64Codec,
		"test", "D23vHLFHsOhPOcvdxeoQyAJTpvM",
		"aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n",
		"0000000000000000000000000000",
		"http://golang.org/doc/gopher/frontpage.png",
		SimpleHeader{"header1": "value1"},
	},

	// valid endoding of the wrong type
	// eg. b64 encoded url supplied to hex decoder
	{
		HexCodec,
		"test", "0f6def1cb147b0e84f39cbddc5ea10c80253a6f3",
		"aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n",
		"",
		"http://golang.org/doc/gopher/frontpage.png",
		nil,
	},
	{
		B64Codec,
		"test", "D23vHLFHsOhPOcvdxeoQyAJTpvM",
		"687474703a2f2f676f6c616e672e6f72672f646f632f676f706865722f66726f6e74706167652e706e67",
		"",
		"http://golang.org/doc/gopher/frontpage.png",
		nil,
	},
}

func TestBadDecodes(t *testing.T) {
	t.Parallel()
	for _, p := range badTests {
		hmacKey := []byte(p.hmac)

		var decode decodeFunc
		switch p.codec {
		case B64Codec:
			decode = base64.RawURLEncoding.DecodeString
		case HexCodec:
			decode = hex.DecodeString
		}

		// test specific decoder
		encodedURL, extraHdr, err := decodeURL(decode, hmacKey, p.edig, p.eURL, p.eExtraHdr)
		assert.Check(t, err != nil, "decoded url verfied when it shouldn't have")
		assert.Check(t, is.Equal(encodedURL, ""), "decoded url result not empty")
		assert.Check(t, is.Equal(len(extraHdr), 0), "decoded extraHdr header map is not empty")

		// also test generic "guessing" decoder
		encodedURL, extraHdr, ok := DecodeURL(hmacKey, p.edig, p.eURL, p.eExtraHdr)
		assert.Check(t, !ok, "decoded url verfied when it shouldn't have")
		assert.Check(t, is.Equal(encodedURL, ""), "decoded url result not empty")
		assert.Check(t, is.Equal(len(extraHdr), 0), "decoded extraHdr header map is not empty")
	}
}
