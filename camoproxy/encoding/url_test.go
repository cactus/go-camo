package encoding

import (
	"testing"
)


func TestEncodeUrl(t *testing.T) {
	t.Parallel()
	hmacKey := []byte("test")
	oUrl := "http://golang.org/doc/gopher/frontpage.png"

	encodedUrl := EncodeUrl(&hmacKey, oUrl)
	if encodedUrl != "/0f6def1cb147b0e84f39cbddc5ea10c80253a6f3/687474703a2f2f676f6c616e672e6f72672f646f632f676f706865722f66726f6e74706167652e706e67" {
		t.Error("encoded url does not match")
	}
}

func TestDecodeUrl(t *testing.T) {
	t.Parallel()
	hmacKey := []byte("test")
	hexdig := "0f6def1cb147b0e84f39cbddc5ea10c80253a6f3"
	hexurl := "687474703a2f2f676f6c616e672e6f72672f646f632f676f706865722f66726f6e74706167652e706e67"

	encodedUrl, ok := DecodeUrl(&hmacKey, hexdig, hexurl)
	if !ok {
		t.Error("decoded url failed to verify")
	}
	if encodedUrl != "http://golang.org/doc/gopher/frontpage.png" {
		t.Error("decoded url did not match")
	}

	hexdig = "000"
	encodedUrl, ok = DecodeUrl(&hmacKey, hexdig, hexurl)
	if ok {
		t.Error("decoded url failed to verify")
	}

	hexdig = "0f6def1cb147b0e84f39cbddc5ea10c80253a6f3"
	hexurl = "000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	encodedUrl, ok = DecodeUrl(&hmacKey, hexdig, hexurl)
	if ok {
		t.Error("decoded url failed to verify")
	}
}

