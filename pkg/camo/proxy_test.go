// Copyright (c) 2012-2019 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package camo

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/cactus/mlog"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
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
	assert.Check(t, err)

	resp, err := processRequest(req, 404, camoConfig, nil)
	if assert.Check(t, err) {
		statusCodeAssert(t, 404, resp)
		bodyAssert(t, "404 Not Found\n", resp)
		headerAssert(t, "test", "X-Go-Camo", resp)
		headerAssert(t, "go-camo", "Server", resp)
	}
}

func TestSimpleValidImageURL(t *testing.T) {
	t.Parallel()
	testURL := "http://www.google.com/images/srpr/logo11w.png"
	resp, err := makeTestReq(testURL, 200, camoConfig)
	if assert.Check(t, err) {
		headerAssert(t, "test", "X-Go-Camo", resp)
		headerAssert(t, "go-camo", "Server", resp)
	}
}

func TestGoogleChartURL(t *testing.T) {
	t.Parallel()
	testURL := "http://chart.apis.google.com/chart?chs=920x200&chxl=0:%7C2010-08-13%7C2010-09-12%7C2010-10-12%7C2010-11-11%7C1:%7C0%7C0%7C0%7C0%7C0%7C0&chm=B,EBF5FB,0,0,0&chco=008Cd6&chls=3,1,0&chg=8.3,20,1,4&chd=s:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA&chxt=x,y&cht=lc"
	_, err := makeTestReq(testURL, 200, camoConfig)
	assert.Check(t, err)
}

func TestChunkedImageFile(t *testing.T) {
	t.Parallel()
	testURL := "https://www.igvita.com/posts/12/spdyproxy-diagram.png"
	_, err := makeTestReq(testURL, 200, camoConfig)
	assert.Check(t, err)
}

func TestFollowRedirects(t *testing.T) {
	t.Parallel()
	testURL := "http://cl.ly/1K0X2Y2F1P0o3z140p0d/boom-headshot.gif"
	_, err := makeTestReq(testURL, 200, camoConfig)
	assert.Check(t, err)
}

func TestStrangeFormatRedirects(t *testing.T) {
	t.Parallel()
	testURL := "http://cl.ly/DPcp/Screen%20Shot%202012-01-17%20at%203.42.32%20PM.png"
	_, err := makeTestReq(testURL, 200, camoConfig)
	assert.Check(t, err)
}

func TestRedirectsWithPathOnly(t *testing.T) {
	t.Parallel()
	testURL := "http://mockbin.org/redirect/302?to=%2Fredirect%2F302%3Fto%3Dhttp%3A%2F%2Fwww.google.com%2Fimages%2Fsrpr%2Flogo11w.png"
	_, err := makeTestReq(testURL, 200, camoConfig)
	assert.Check(t, err)
}

func TestFollowPermRedirects(t *testing.T) {
	t.Parallel()
	testURL := "http://mockbin.org/redirect/301?to=http://www.google.com/images/srpr/logo11w.png"
	_, err := makeTestReq(testURL, 200, camoConfig)
	assert.Check(t, err)
}

func TestFollowTempRedirects(t *testing.T) {
	t.Parallel()
	testURL := "http://mockbin.org/redirect/302?to=http://www.google.com/images/srpr/logo11w.png"
	_, err := makeTestReq(testURL, 200, camoConfig)
	assert.Check(t, err)
}

func TestBadContentType(t *testing.T) {
	t.Parallel()
	testURL := "http://httpbin.org/response-headers?Content-Type=what"
	_, err := makeTestReq(testURL, 400, camoConfig)
	assert.Check(t, err)
}

func TestContentTypeParams(t *testing.T) {
	t.Parallel()
	testURL := "http://httpbin.org/response-headers?Content-Type=image/svg%2Bxml;charset=UTF-8"
	resp, err := makeTestReq(testURL, 200, camoConfig)

	assert.Check(t, err)
	if assert.Check(t, err) {
		headerAssert(t, "image/svg+xml; charset=UTF-8", "content-type", resp)
	}
}

func TestBadContentTypeSmuggle(t *testing.T) {
	t.Parallel()
	testURL := "http://httpbin.org/response-headers?Content-Type=image/png,%20text/html;%20charset%3DUTF-8"
	_, err := makeTestReq(testURL, 400, camoConfig)
	assert.Check(t, err)

	testURL = "http://httpbin.org/response-headers?Content-Type=image/png,text/html;%20charset%3DUTF-8"
	_, err = makeTestReq(testURL, 400, camoConfig)
	assert.Check(t, err)

	testURL = "http://httpbin.org/response-headers?Content-Type=image/png%20text/html"
	_, err = makeTestReq(testURL, 400, camoConfig)
	assert.Check(t, err)

	testURL = "http://httpbin.org/response-headers?Content-Type=image/png%;text/html"
	_, err = makeTestReq(testURL, 400, camoConfig)
	assert.Check(t, err)

	testURL = "http://httpbin.org/response-headers?Content-Type=image/png;%20charset%3DUTF-8;text/html"
	_, err = makeTestReq(testURL, 400, camoConfig)
	assert.Check(t, err)
}

func TestXForwardedFor(t *testing.T) {
	t.Parallel()

	camoConfigWithoutFwd4 := Config{
		HMACKey:        []byte("0x24FEEDFACEDEADBEEFCAFE"),
		MaxSize:        180 * 1024,
		RequestTimeout: time.Duration(10) * time.Second,
		MaxRedirects:   3,
		ServerName:     "go-camo",
		EnableXFwdFor:  true,
		noIPFiltering:  true,
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Close = true
		w.Header().Set("Content-Type", "image/png")
		_, err := w.Write([]byte(r.Header.Get("X-Forwarded-For")))
		assert.Check(t, err)
	}))
	defer ts.Close()

	req, err := makeReq(camoConfig, ts.URL)
	assert.Check(t, err)

	req.Header.Set("X-Forwarded-For", "2.2.2.2, 1.1.1.1")

	resp, err := processRequest(req, 200, camoConfigWithoutFwd4, nil)
	assert.Check(t, err)
	bodyAssert(t, "2.2.2.2, 1.1.1.1", resp)

	camoConfigWithoutFwd4.EnableXFwdFor = false
	resp, err = processRequest(req, 200, camoConfigWithoutFwd4, nil)
	assert.Check(t, err)
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

	testURL := "http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ForBiggerEscapes.mp4"

	// try a range request (should succeed, MaxSize is larger than requested range)
	req, err := makeReq(camoConfigWithVideo, testURL)
	assert.Check(t, err)
	req.Header.Add("Range", "bytes=0-10")
	resp, err := processRequest(req, 206, camoConfigWithVideo, nil)
	assert.Check(t, is.Equal(resp.Header.Get("Content-Range"), "bytes 0-10/2299653"))
	assert.Check(t, err)

	// try a range request (should fail, MaxSize is smaller than requested range)
	camoConfigWithVideo.MaxSize = 1 * 1024
	req, err = makeReq(camoConfigWithVideo, testURL)
	assert.Check(t, err)
	req.Header.Add("Range", "bytes=0-1025")
	_, err = processRequest(req, 404, camoConfigWithVideo, nil)
	assert.Check(t, err)

	// try full request (should fail, too large)
	_, err = makeTestReq(testURL, 404, camoConfigWithVideo)
	assert.Check(t, err)

	// bump limit, try again (should succeed)
	camoConfigWithVideo.MaxSize = 5000 * 1024
	_, err = makeTestReq(testURL, 200, camoConfigWithVideo)
	//fmt.Println(err)
	assert.Check(t, err)
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
	assert.Check(t, err)

	// try a range request
	req, err := makeReq(camoConfigWithAudio, testURL)
	assert.Check(t, err)
	req.Header.Add("Range", "bytes=0-10")
	resp, err := processRequest(req, 206, camoConfigWithAudio, nil)
	assert.Check(t, is.Equal(resp.Header.Get("Content-Range"), "bytes 0-10/49872"))
	assert.Check(t, err)
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
	assert.Check(t, err)
}

func TestSupplyAcceptIfNoneGiven(t *testing.T) {
	t.Parallel()
	testURL := "http://images.anandtech.com/doci/6673/OpenMoboAMD30_575px.png"
	req, err := makeReq(camoConfig, testURL)
	req.Header.Del("Accept")
	assert.Check(t, err)
	_, err = processRequest(req, 200, camoConfig, nil)
	assert.Check(t, err)
}

func Test404OnVideo(t *testing.T) {
	t.Parallel()
	testURL := "http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ForBiggerEscapes.mp4"
	_, err := makeTestReq(testURL, 400, camoConfig)
	assert.Check(t, err)
}

func Test404OnCredentialURL(t *testing.T) {
	t.Parallel()
	testURL := "http://user:pass@www.google.com/images/srpr/logo11w.png"
	_, err := makeTestReq(testURL, 404, camoConfig)
	assert.Check(t, err)
}

func Test404InfiniRedirect(t *testing.T) {
	t.Parallel()
	testURL := "http://mockbin.org/redirect/302/4"
	_, err := makeTestReq(testURL, 404, camoConfig)
	assert.Check(t, err)
}

func Test404URLWithoutHTTPHost(t *testing.T) {
	t.Parallel()
	testURL := "/picture/Mincemeat/Pimp.jpg"
	_, err := makeTestReq(testURL, 404, camoConfig)
	assert.Check(t, err)
}

func Test404ImageLargerThan5MB(t *testing.T) {
	t.Parallel()
	testURL := "https://apod.nasa.gov/apod/image/0505/larryslookout_spirit_big.jpg"
	_, err := makeTestReq(testURL, 404, camoConfig)
	assert.Check(t, err)
}

func Test404HostNotFound(t *testing.T) {
	t.Parallel()
	testURL := "http://flabergasted.cx"
	_, err := makeTestReq(testURL, 404, camoConfig)
	assert.Check(t, err)
}

func Test404OnExcludes(t *testing.T) {
	t.Parallel()
	testURL := "http://iphone.internal.example.org/foo.cgi"
	_, err := makeTestReq(testURL, 404, camoConfig)
	assert.Check(t, err)
}

func Test404OnNonImageContent(t *testing.T) {
	t.Parallel()
	testURL := "https://github.com/atmos/cinderella/raw/master/bootstrap.sh"
	_, err := makeTestReq(testURL, 404, camoConfig)
	assert.Check(t, err)
}

func Test404On10xIpRange(t *testing.T) {
	t.Parallel()
	testURL := "http://10.0.0.1/foo.cgi"
	_, err := makeTestReq(testURL, 404, camoConfig)
	assert.Check(t, err)
}

func Test404On169Dot254Net(t *testing.T) {
	t.Parallel()
	testURL := "http://169.254.0.1/foo.cgi"
	_, err := makeTestReq(testURL, 404, camoConfig)
	assert.Check(t, err)
}

func Test404On172Dot16Net(t *testing.T) {
	t.Parallel()
	for i := 16; i < 32; i++ {
		testURL := "http://172.%d.0.1/foo.cgi"
		_, err := makeTestReq(fmt.Sprintf(testURL, i), 404, camoConfig)
		assert.Check(t, err)
	}
}

func Test404On192Dot168Net(t *testing.T) {
	t.Parallel()
	testURL := "http://192.168.0.1/foo.cgi"
	_, err := makeTestReq(testURL, 404, camoConfig)
	assert.Check(t, err)
}

func Test404OnLocalhost(t *testing.T) {
	t.Parallel()
	testURL := "http://localhost/foo.cgi"
	resp, err := makeTestReq(testURL, 404, camoConfig)
	if assert.Check(t, err) {
		bodyAssert(t, "Bad url host\n", resp)
	}
}

func Test404OnLocalhostWithPort(t *testing.T) {
	t.Parallel()
	testURL := "http://localhost:80/foo.cgi"
	resp, err := makeTestReq(testURL, 404, camoConfig)
	if assert.Check(t, err) {
		bodyAssert(t, "Bad url host\n", resp)
	}
}

func Test404OnRedirectWithLocalhostTarget(t *testing.T) {
	t.Parallel()
	testURL := "http://mockbin.org/redirect/302?to=http://localhost/some.png"
	resp, err := makeTestReq(testURL, 404, camoConfig)
	if assert.Check(t, err) {
		bodyAssert(t, "Error Fetching Resource\n", resp)
	}
}

func Test404OnRedirectWithLoopbackIP(t *testing.T) {
	t.Parallel()
	testURL := "http://mockbin.org/redirect/302?to=http://127.0.0.100/some.png"
	resp, err := makeTestReq(testURL, 404, camoConfig)
	if assert.Check(t, err) {
		bodyAssert(t, "Error Fetching Resource\n", resp)
	}
}

func Test404OnRedirectWithLoopbackIPwCreds(t *testing.T) {
	t.Parallel()
	testURL := "http://mockbin.org/redirect/302?to=http://user:pass@127.0.0.100/some.png"
	resp, err := makeTestReq(testURL, 404, camoConfig)
	if assert.Check(t, err) {
		bodyAssert(t, "Error Fetching Resource\n", resp)
	}
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

	testURL := "http://mockbin.org/redirect/302?to=http://test.vcap.me"
	req, err := makeReq(camoConfig, testURL)
	assert.Check(t, err)

	resp, err := processRequest(req, 404, camoConfig, nil)
	if assert.Check(t, err) {
		bodyAssert(t, "Error Fetching Resource\n", resp)
	}
}

func TestDownloadDisposition(t *testing.T) {
	t.Parallel()
	testURL := "https://upload.wikimedia.org/wikipedia/commons/thumb/e/ea/Proxi%2C_Bordeaux%2C_July_2014.JPG/1200px-Proxi%2C_Bordeaux%2C_July_2014.JPG"
	req, err := makeReq(camoConfig, testURL)
	assert.Check(t, err)
	resp, err := processRequest(req, 200, camoConfig, nil)
	headerAssert(t, "", "Content-Disposition", resp)
	assert.Check(t, err)

	req.URL.Path = req.URL.Path + "/download"
	resp, err = processRequest(req, 200, camoConfig, nil)
	headerAssert(t, "attachment; filename=\"Proxi,_Bordeaux,_July_2014.JPG\"", "Content-Disposition", resp)
	assert.Check(t, err)
}

func TestDownloadDispositionUnicode(t *testing.T) {
	t.Parallel()
	testURL := "https://upload.wikimedia.org/wikipedia/commons/thumb/1/16/Aillant-sur-Tholon-FR-89-Proxi_March%C3%A9-09.jpg/1200px-Aillant-sur-Tholon-FR-89-Proxi_March%C3%A9-09.jpg"
	req, err := makeReq(camoConfig, testURL)
	assert.Check(t, err)
	resp, err := processRequest(req, 200, camoConfig, nil)
	headerAssert(t, "", "Content-Disposition", resp)
	assert.Check(t, err)

	req.URL.Path = req.URL.Path + "/download"
	resp, err = processRequest(req, 200, camoConfig, nil)
	headerAssert(t, "attachment; filename*=utf-8''Aillant-sur-Tholon-FR-89-Proxi_March%C3%A9-09.jpg", "Content-Disposition", resp)
	assert.Check(t, err)
}

func TestPathPieces(t *testing.T) {
	t.Parallel()
	testURL := "http://www.google.com/images/srpr/logo11w.png"
	req, err := makeReq(camoConfig, testURL)
	assert.Check(t, err)
	req.URL.Path = req.URL.Path + "/"
	resp, err := processRequest(req, 404, camoConfig, nil)
	bodyAssert(t, "Malformed request path\n", resp)
	assert.Check(t, err)

	req.URL.Path = req.URL.Path + "down"
	resp, err = processRequest(req, 404, camoConfig, nil)
	bodyAssert(t, "Malformed request path\n", resp)
	assert.Check(t, err)

	req.URL.Path = req.URL.Path + "/load"
	resp, err = processRequest(req, 404, camoConfig, nil)
	bodyAssert(t, "404 Not Found\n", resp)
	assert.Check(t, err)

	req.URL.Path = "/download"
	resp, err = processRequest(req, 404, camoConfig, nil)
	bodyAssert(t, "404 Not Found\n", resp)
	assert.Check(t, err)
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
