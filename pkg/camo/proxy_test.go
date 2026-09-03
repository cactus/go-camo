// Copyright (c) 2012-2023 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package camo

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"codeberg.org/dropwhile/mlog"
	"github.com/cactus/go-camo/v2/pkg/assert"
)

var camoConfig = Config{
	HMACKey:             []byte("0x24FEEDFACEDEADBEEFCAFE"),
	MaxSize:             5120 * 1024,
	RequestTimeout:      time.Duration(10) * time.Second,
	MaxRedirects:        3,
	ServerName:          "go-camo",
	AllowContentVideo:   false,
	AllowCredentialURLs: false,
}

func skipIfCI(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping test. CI environments generally enable something similar to unbound's private-address functionality, making this test fail.")
	}
}

func TestNotFound(t *testing.T) {
	t.Parallel()
	req, err := http.NewRequest("GET", "http://example.com/favicon.ico", nil)
	assert.Nil(t, err)

	resp, err := processRequest(req, 404, camoConfig, nil)
	assert.Nil(t, err)
	statusCodeAssert(t, 404, resp)
	bodyAssert(t, "404 Not Found\n", resp)
	headerAssert(t, "test", "X-Go-Camo", resp)
	headerAssert(t, "go-camo", "Server", resp)
}

func TestSimpleValidImageURL(t *testing.T) {
	t.Parallel()
	testURL := "http://www.google.com/images/srpr/logo11w.png"
	resp, err := makeTestReq(testURL, 200, camoConfig)
	assert.Nil(t, err)
	headerAssert(t, "test", "X-Go-Camo", resp)
	headerAssert(t, "go-camo", "Server", resp)
}

func TestChunkedImageFile(t *testing.T) {
	t.Parallel()
	testURL := "https://www.igvita.com/posts/12/spdyproxy-diagram.png"
	_, err := makeTestReq(testURL, 200, camoConfig)
	assert.Nil(t, err)
}

func TestFollowRedirects(t *testing.T) {
	t.Parallel()
	testURL := "http://cl.ly/1K0X2Y2F1P0o3z140p0d/boom-headshot.gif"
	_, err := makeTestReq(testURL, 200, camoConfig)
	assert.Nil(t, err)
}

func TestStrangeFormatRedirects(t *testing.T) {
	t.Parallel()
	testURL := "http://cl.ly/DPcp/Screen%20Shot%202012-01-17%20at%203.42.32%20PM.png"
	_, err := makeTestReq(testURL, 200, camoConfig)
	assert.Nil(t, err)
}

func TestRedirectsWithPathOnly(t *testing.T) {
	t.Parallel()
	testURL := "http://httpbin.org/redirect-to?status_code=302&url=%2Fredirect-to%3Furl%3Dhttp%3A%2F%2Fwww.google.com%2Fimages%2Fsrpr%2Flogo11w.png%26status_code%3D302"
	_, err := makeTestReq(testURL, 200, camoConfig)
	assert.Nil(t, err)
}

func TestFollowPermRedirects(t *testing.T) {
	t.Parallel()
	testURL := "http://httpbin.org/redirect-to?status_code=301&url=http://www.google.com/images/srpr/logo11w.png"
	_, err := makeTestReq(testURL, 200, camoConfig)
	assert.Nil(t, err)
}

func TestFollowTempRedirects(t *testing.T) {
	t.Parallel()
	testURL := "http://httpbin.org/redirect-to?status_code=302&url=http://www.google.com/images/srpr/logo11w.png"
	_, err := makeTestReq(testURL, 200, camoConfig)
	assert.Nil(t, err)
}

func TestBadContentType(t *testing.T) {
	t.Parallel()
	testURL := "http://httpbin.org/response-headers?Content-Type=what"
	_, err := makeTestReq(testURL, 400, camoConfig)
	assert.Nil(t, err)
}

func TestContentTypeParams(t *testing.T) {
	t.Parallel()
	testURL := "http://httpbin.org/response-headers?Content-Type=image/svg%2Bxml;charset=UTF-8"
	resp, err := makeTestReq(testURL, 200, camoConfig)

	assert.Nil(t, err)
	headerAssert(t, "image/svg+xml; charset=UTF-8", "content-type", resp)
}

func TestBadContentTypeSmuggle(t *testing.T) {
	t.Parallel()
	testURL := "http://httpbin.org/response-headers?Content-Type=image/png,%20text/html;%20charset%3DUTF-8"
	_, err := makeTestReq(testURL, 400, camoConfig)
	assert.Nil(t, err)

	testURL = "http://httpbin.org/response-headers?Content-Type=image/png,text/html;%20charset%3DUTF-8"
	_, err = makeTestReq(testURL, 400, camoConfig)
	assert.Nil(t, err)

	testURL = "http://httpbin.org/response-headers?Content-Type=image/png%20text/html"
	_, err = makeTestReq(testURL, 400, camoConfig)
	assert.Nil(t, err)

	testURL = "http://httpbin.org/response-headers?Content-Type=image/png%;text/html"
	_, err = makeTestReq(testURL, 400, camoConfig)
	assert.Nil(t, err)

	testURL = "http://httpbin.org/response-headers?Content-Type=image/png;%20charset%3DUTF-8;text/html"
	_, err = makeTestReq(testURL, 400, camoConfig)
	assert.Nil(t, err)
}

func TestXForwardedFor(t *testing.T) {
	t.Parallel()

	camoConfigWithoutFwd4 := Config{
		HMACKey:          []byte("0x24FEEDFACEDEADBEEFCAFE"),
		MaxSize:          180 * 1024,
		RequestTimeout:   time.Duration(10) * time.Second,
		MaxRedirects:     3,
		ServerName:       "go-camo",
		EnableXFwdFor:    true,
		insecureTestMode: true,
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Close = true
		w.Header().Set("Content-Type", "image/png")
		_, err := w.Write([]byte(r.Header.Get("X-Forwarded-For")))
		assert.Nil(t, err)
	}))
	defer ts.Close()

	req, err := makeReq(camoConfig, ts.URL)
	assert.Nil(t, err)

	req.Header.Set("X-Forwarded-For", "2.2.2.2, 1.1.1.1")

	resp, err := processRequest(req, 200, camoConfigWithoutFwd4, nil)
	assert.Nil(t, err)
	bodyAssert(t, "2.2.2.2, 1.1.1.1", resp)

	camoConfigWithoutFwd4.EnableXFwdFor = false
	resp, err = processRequest(req, 200, camoConfigWithoutFwd4, nil)
	assert.Nil(t, err)
	bodyAssert(t, "", resp)
}

func TestVideoContentTypeAllowed(t *testing.T) {
	t.Parallel()

	camoConfigWithVideo := Config{
		HMACKey:           []byte("0x24FEEDFACEDEADBEEFCAFE"),
		MaxSize:           180 * 1024,
		RequestTimeout:    time.Duration(10) * time.Second,
		MaxRedirects:      3,
		ServerName:        "go-camo",
		AllowContentVideo: true,
	}

	testURL := "https://www.w3schools.com/tags/mov_bbb.mp4"

	// try a range request (should succeed, MaxSize is larger than requested range)
	req, err := makeReq(camoConfigWithVideo, testURL)
	assert.Nil(t, err)
	req.Header.Add("Range", "bytes=0-10")
	resp, err := processRequest(req, 206, camoConfigWithVideo, nil)
	assert.Equal(t, resp.Header.Get("Content-Range"), "bytes 0-10/788493")
	assert.Nil(t, err)

	// try a range request (should fail, MaxSize is smaller than requested range size)
	camoConfigWithVideo.MaxSize = 1 * 1024
	req, err = makeReq(camoConfigWithVideo, testURL)
	assert.Nil(t, err)
	req.Header.Add("Range", "bytes=0-1025")
	_, err = processRequest(req, 404, camoConfigWithVideo, nil)
	assert.Nil(t, err)

	// try a range request (should fail, MaxSize is smaller than requested range start)
	camoConfigWithVideo.MaxSize = 1 * 1024
	req, err = makeReq(camoConfigWithVideo, testURL)
	assert.Nil(t, err)
	req.Header.Add("Range", "bytes=1025-1026")
	_, err = processRequest(req, 404, camoConfigWithVideo, nil)
	assert.Nil(t, err)

	// try a range request (should fail, MaxSize is smaller than requested range start)
	camoConfigWithVideo.MaxSize = 1 * 1024
	req, err = makeReq(camoConfigWithVideo, testURL)
	assert.Nil(t, err)
	req.Header.Add("Range", "bytes=0-100,1026")
	_, err = processRequest(req, 404, camoConfigWithVideo, nil)
	assert.Nil(t, err)

	// try full request (should fail, too large)
	_, err = makeTestReq(testURL, 404, camoConfigWithVideo)
	assert.Nil(t, err)

	// bump limit, try again (should succeed)
	camoConfigWithVideo.MaxSize = 5000 * 1024
	_, err = makeTestReq(testURL, 200, camoConfigWithVideo)
	assert.Nil(t, err)
}

func TestAudioContentTypeAllowed(t *testing.T) {
	t.Parallel()

	camoConfigWithAudio := Config{
		HMACKey:           []byte("0x24FEEDFACEDEADBEEFCAFE"),
		MaxSize:           180 * 1024,
		RequestTimeout:    time.Duration(10) * time.Second,
		MaxRedirects:      3,
		ServerName:        "go-camo",
		AllowContentAudio: true,
	}

	testURL := "https://actions.google.com/sounds/v1/alarms/alarm_clock.ogg"
	_, err := makeTestReq(testURL, 200, camoConfigWithAudio)
	assert.Nil(t, err)

	// try a range request
	req, err := makeReq(camoConfigWithAudio, testURL)
	assert.Nil(t, err)
	req.Header.Add("Range", "bytes=0-10")
	resp, err := processRequest(req, 206, camoConfigWithAudio, nil)
	assert.Equal(t, resp.Header.Get("Content-Range"), "bytes 0-10/49872")
	assert.Nil(t, err)
}

func TestCredetialURLsAllowed(t *testing.T) {
	t.Parallel()

	camoConfigWithCredentials := Config{
		HMACKey:             []byte("0x24FEEDFACEDEADBEEFCAFE"),
		MaxSize:             180 * 1024,
		RequestTimeout:      time.Duration(10) * time.Second,
		MaxRedirects:        3,
		ServerName:          "go-camo",
		AllowCredentialURLs: true,
	}

	testURL := "http://user:pass@www.google.com/images/srpr/logo11w.png"
	_, err := makeTestReq(testURL, 200, camoConfigWithCredentials)
	assert.Nil(t, err)
}

func TestMaxSizeRedirect(t *testing.T) {
	t.Parallel()

	camoConfigWithMaxSizeRedirect := Config{
		HMACKey:           []byte("0x24FEEDFACEDEADBEEFCAFE"),
		MaxSize:           1 * 1024,
		RequestTimeout:    time.Duration(10) * time.Second,
		MaxRedirects:      3,
		MaxSizeRedirect:   "http://example.com/some-image.png",
		ServerName:        "go-camo",
		AllowContentVideo: true,
	}

	testURL := "https://www.w3schools.com/tags/mov_bbb.mp4"

	// try a range request (should fail, MaxSize is smaller than requested range)
	req, err := makeReq(camoConfigWithMaxSizeRedirect, testURL)
	assert.Nil(t, err)
	req.Header.Add("Range", "bytes=0-1025")
	resp, err := processRequest(req, 302, camoConfigWithMaxSizeRedirect, nil)
	assert.Equal(t, resp.Header.Get("Location"), camoConfigWithMaxSizeRedirect.MaxSizeRedirect)
	assert.Nil(t, err)
}

func TestMaxSizeBackendContentLength(t *testing.T) {
	t.Parallel()

	chunksCount := 100
	chunkSizeBytes := 512
	maxAllowedBytes := 1024 * 1

	camoConfigWithMaxSize := Config{
		HMACKey:           []byte("0x24FEEDFACEDEADBEEFCAFE"),
		MaxSize:           int64(maxAllowedBytes),
		RequestTimeout:    time.Duration(10) * time.Second,
		ServerName:        "go-camo",
		AllowContentVideo: true,
		insecureTestMode:  true,
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("x-content-type-options", "nosniff")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", chunksCount*chunkSizeBytes))
		for i := 1; i <= chunksCount; i++ {
			w.Write(bytes.Repeat([]byte("x"), chunkSizeBytes))
		}
	}))
	defer ts.Close()

	// try a range request (should fail, MaxSize is smaller than requested range)
	req, err := makeReq(camoConfigWithMaxSize, ts.URL)
	assert.Nil(t, err)
	resp, err := processRequest(req, 404, camoConfigWithMaxSize, nil)
	assert.Nil(t, err)
	b, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	err = resp.Body.Close()
	assert.Nil(t, err)
	assert.True(t, len(b) < maxAllowedBytes)
	assert.Equal(t, string(b), "Content length exceeded\n")
}

func TestMaxSizeBackendChunked(t *testing.T) {
	t.Parallel()

	chunksCount := 100
	chunkSizeBytes := 1024
	maxAllowedBytes := 1024 * 40

	camoConfigWithMaxSize := Config{
		HMACKey:           []byte("0x24FEEDFACEDEADBEEFCAFE"),
		MaxSize:           int64(maxAllowedBytes),
		RequestTimeout:    time.Duration(10) * time.Second,
		ServerName:        "go-camo",
		AllowContentVideo: true,
		insecureTestMode:  true,
		CollectMetrics:    true,
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rc := http.NewResponseController(w)

		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Content-Type", "video/mp4")
		w.WriteHeader(200)
		for i := 1; i <= chunksCount; i++ {
			w.Write(bytes.Repeat([]byte("x"), chunkSizeBytes))
			rc.Flush()
		}
	}))
	defer ts.Close()

	// try a chunked request (should succeed, but be truncated)
	req, err := makeReq(camoConfigWithMaxSize, ts.URL)
	assert.Nil(t, err)
	resp, err := processRequest(req, 200, camoConfigWithMaxSize, nil)
	assert.Nil(t, err)
	b, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	err = resp.Body.Close()
	assert.Nil(t, err)
	assert.Equal(t, len(b), maxAllowedBytes)

	// check trailer
	assert.Equal(t, resp.Trailer.Get("Camo-Chunked-Truncation"), "true")
}

func TestSupplyAcceptIfNoneGiven(t *testing.T) {
	t.Parallel()
	testURL := "https://picsum.photos/200/300"
	req, err := makeReq(camoConfig, testURL)
	assert.Nil(t, err)
	req.Header.Del("Accept")
	_, err = processRequest(req, 200, camoConfig, nil)
	assert.Nil(t, err)
}

func Test404OnVideo(t *testing.T) {
	t.Parallel()
	testURL := "https://www.w3schools.com/tags/mov_bbb.mp4"
	_, err := makeTestReq(testURL, 400, camoConfig)
	assert.Nil(t, err)
}

func Test404OnCredentialURL(t *testing.T) {
	t.Parallel()
	testURL := "http://user:pass@www.google.com/images/srpr/logo11w.png"
	_, err := makeTestReq(testURL, 404, camoConfig)
	assert.Nil(t, err)
}

func Test404InfiniRedirect(t *testing.T) {
	t.Parallel()
	testURL := "http://httpbin.org/redirect/4"
	_, err := makeTestReq(testURL, 404, camoConfig)
	assert.Nil(t, err)
}

func Test404URLWithoutHTTPHost(t *testing.T) {
	t.Parallel()
	testURL := "/picture/Mincemeat/Pimp.jpg"
	_, err := makeTestReq(testURL, 404, camoConfig)
	assert.Nil(t, err)
}

func Test404ImageLargerThan5MB(t *testing.T) {
	t.Parallel()
	testURL := "https://apod.nasa.gov/apod/image/0505/larryslookout_spirit_big.jpg"
	_, err := makeTestReq(testURL, 404, camoConfig)
	assert.Nil(t, err)
}

func Test404HostNotFound(t *testing.T) {
	t.Parallel()
	testURL := "http://flabergasted.cx"
	_, err := makeTestReq(testURL, 404, camoConfig)
	assert.Nil(t, err)
}

func Test404OnExcludes(t *testing.T) {
	t.Parallel()
	testURL := "http://iphone.internal.example.org/foo.cgi"
	_, err := makeTestReq(testURL, 404, camoConfig)
	assert.Nil(t, err)
}

func Test404OnNonImageContent(t *testing.T) {
	t.Parallel()
	testURL := "https://github.com/atmos/cinderella/raw/master/bootstrap.sh"
	_, err := makeTestReq(testURL, 404, camoConfig)
	assert.Nil(t, err)
}

func Test404On10xIpRange(t *testing.T) {
	t.Parallel()
	testURL := "http://10.0.0.1/foo.cgi"
	_, err := makeTestReq(testURL, 404, camoConfig)
	assert.Nil(t, err)
}

func Test404On169Dot254Net(t *testing.T) {
	t.Parallel()
	testURL := "http://169.254.0.1/foo.cgi"
	_, err := makeTestReq(testURL, 404, camoConfig)
	assert.Nil(t, err)
}

func Test404On172Dot16Net(t *testing.T) {
	t.Parallel()
	for i := 16; i < 32; i++ {
		testURL := "http://172.%d.0.1/foo.cgi"
		_, err := makeTestReq(fmt.Sprintf(testURL, i), 404, camoConfig)
		assert.Nil(t, err)
	}
}

func Test404On192Dot168Net(t *testing.T) {
	t.Parallel()
	testURL := "http://192.168.0.1/foo.cgi"
	_, err := makeTestReq(testURL, 404, camoConfig)
	assert.Nil(t, err)
}

func Test404OnLocalhost(t *testing.T) {
	t.Parallel()
	testURL := "http://localhost/foo.cgi"
	resp, err := makeTestReq(testURL, 404, camoConfig)
	assert.Nil(t, err)
	bodyAssert(t, "Bad url host\n", resp)
}

func Test404OnLocalhostWithPort(t *testing.T) {
	t.Parallel()
	testURL := "http://localhost:80/foo.cgi"
	resp, err := makeTestReq(testURL, 404, camoConfig)
	assert.Nil(t, err)
	bodyAssert(t, "Bad url host\n", resp)
}

func Test404OnRedirectWithLocalhostTarget(t *testing.T) {
	t.Parallel()
	testURL := "http://httpbin.org/redirect-to?status_code=302&url=http://localhost/some.png"
	resp, err := makeTestReq(testURL, 404, camoConfig)
	assert.Nil(t, err)
	bodyAssert(t, "Error Fetching Resource\n", resp)
}

func Test404OnRedirectWithLoopbackIP(t *testing.T) {
	t.Parallel()
	testURL := "http://httpbin.org/redirect-to?status_code=302&url=http://127.0.0.100/some.png"
	resp, err := makeTestReq(testURL, 404, camoConfig)
	assert.Nil(t, err)
	bodyAssert(t, "Error Fetching Resource\n", resp)
}

func Test404OnRedirectWithLoopbackIPwCreds(t *testing.T) {
	t.Parallel()
	testURL := "http://httpbin.org/redirect-to?status_code=302&url=http://user:pass@127.0.0.100/some.png"
	resp, err := makeTestReq(testURL, 404, camoConfig)
	assert.Nil(t, err)
	bodyAssert(t, "Error Fetching Resource\n", resp)
}

// Test will always pass if dns relay implements dns rebind prevention
//
// Even if local dns is doing rebinding prevention, we will still get back the
// same final response. The difference is where the error originates. If there
// is no dns rebinding prevention in the dns resolver, then the proxy code
// rejects in net.dail. If there IS dns rebinding prevention, the dns resolver
// does not return an ip address, and we get a "No address associated with
// hostname" result.
// As such, there is little sense running this when dns rebinding
// prevention is present in the dns resolver....
func Test404OnLoopback(t *testing.T) {
	skipIfCI(t)
	t.Parallel()

	testURL := "http://httpbin.org/redirect-to?status_code=302&url=http://lvh.me"
	req, err := makeReq(camoConfig, testURL)
	assert.Nil(t, err)

	resp, err := processRequest(req, 404, camoConfig, nil)
	assert.Nil(t, err)
	bodyAssert(t, "Error Fetching Resource\n", resp)
}

func TestMain(m *testing.M) {
	flag.Parse()

	debug := os.Getenv("DEBUG")
	// now configure a standard logger
	mlog.SetFlags(mlog.Lstd)

	if debug != "" {
		mlog.SetFlags(mlog.Flags() | mlog.Ldebug)
		mlog.Debug("debug logging enabled")
	}

	os.Exit(m.Run())
}
