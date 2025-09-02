// Copyright (c) 2012-2023 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package encoding

import (
	"fmt"
	"testing"

	"github.com/dropwhile/assert"
)

func TestEncoder(t *testing.T) {
	t.Parallel()

	f := func(encoder EncoderFunc, hmac, edig, eURL, sURL string) {
		t.Helper()
		hmacKey := []byte(hmac)
		// test specific encoder
		encodedURL := encoder(hmacKey, sURL)
		assert.Equal(t, encodedURL, fmt.Sprintf("/%s/%s", edig, eURL), "encoded url does not match")
	}

	// hex
	f(
		HexEncodeURL, "test", "0f6def1cb147b0e84f39cbddc5ea10c80253a6f3",
		"687474703a2f2f676f6c616e672e6f72672f646f632f676f706865722f66726f6e74706167652e706e67",
		"http://golang.org/doc/gopher/frontpage.png",
	)

	// base64
	f(
		B64EncodeURL, "test", "D23vHLFHsOhPOcvdxeoQyAJTpvM",
		"aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n",
		"http://golang.org/doc/gopher/frontpage.png",
	)
}

func TestDecoder(t *testing.T) {
	t.Parallel()

	f := func(decoder DecoderFunc, hmac, edig, eURL, sURL string) {
		t.Helper()
		hmacKey := []byte(hmac)
		// test specific decoder
		encodedURL, err := decoder(hmacKey, edig, eURL)
		assert.Nil(t, err, "decoded url failed to verify")
		assert.Equal(t, encodedURL, sURL, "decoded url does not match")

		// also test generic "guessing" decoder
		encodedURL, ok := DecodeURL(hmacKey, edig, eURL)
		assert.True(t, ok, "decoded url failed to verify")
		assert.Equal(t, encodedURL, sURL, "decoded url does not match")
	}

	// hex
	f(
		HexDecodeURL, "test", "0f6def1cb147b0e84f39cbddc5ea10c80253a6f3",
		"687474703a2f2f676f6c616e672e6f72672f646f632f676f706865722f66726f6e74706167652e706e67",
		"http://golang.org/doc/gopher/frontpage.png",
	)

	// base64
	f(
		B64DecodeURL, "test", "D23vHLFHsOhPOcvdxeoQyAJTpvM",
		"aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n",
		"http://golang.org/doc/gopher/frontpage.png",
	)
}

func TestBadDecodes(t *testing.T) {
	t.Parallel()

	f := func(decoder DecoderFunc, hmac, edig, eURL string) {
		t.Helper()
		hmacKey := []byte(hmac)
		// test specific decoder
		encodedURL, err := decoder(hmacKey, edig, eURL)
		assert.NotNil(t, err, "decoded url verfied when it shouldn't have")
		assert.Equal(t, encodedURL, "", "decoded url result not empty")

		// also test generic "guessing" decoder
		encodedURL, ok := DecodeURL(hmacKey, edig, eURL)
		assert.False(t, ok, "decoded url verfied when it shouldn't have")
		assert.Equal(t, encodedURL, "", "decoded url result not empty")
	}

	// hex
	f(
		HexDecodeURL, "test", "000",
		"687474703a2f2f676f6c616e672e6f72672f646f632f676f706865722f66726f6e74706167652e706e67",
	)
	f(
		HexDecodeURL, "test", "0f6def1cb147b0e84f39cbddc5ea10c80253a6f3",
		"000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
	)

	// base64
	f(
		B64DecodeURL, "test", "000",
		"aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n",
	)
	f(
		B64DecodeURL, "test", "D23vHLFHsOhPOcvdxeoQyAJTpvM",
		"00000000000000000000000000000000000000000000000000000000",
	)

	// mixmatch
	// hex
	f(
		HexDecodeURL, "test", "0f6def1cb147b0e84f39cbddc5ea10c80253a6f3",
		"aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n",
	)

	// base64
	f(
		B64DecodeURL, "test", "D23vHLFHsOhPOcvdxeoQyAJTpvM",
		"687474703a2f2f676f6c616e672e6f72672f646f632f676f706865722f66726f6e74706167652e706e67",
	)
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
