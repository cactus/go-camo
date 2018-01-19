// Copyright (c) 2012-2016 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package camo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-camo/camo/encoding"
	"go-camo/router"

	"github.com/stretchr/testify/assert"
)

var camoConfig = Config{
	HMACKey:        []byte("0x24FEEDFACEDEADBEEFCAFE"),
	MaxSize:        5120 * 1024,
	RequestTimeout: time.Duration(10) * time.Second,
	MaxRedirects:   3,
	ServerName:     "go-camo",
}

func makeReq(testURL string) (*http.Request, error) {
	k := []byte(camoConfig.HMACKey)
	hexURL := encoding.B64EncodeURL(k, testURL)
	out := "http://example.com" + hexURL
	req, err := http.NewRequest("GET", out, nil)
	if err != nil {
		return nil, fmt.Errorf("Error building req url '%s': %s", testURL, err.Error())
	}
	return req, nil
}

func processRequest(req *http.Request, status int, camoConfig Config) (*httptest.ResponseRecorder, error) {
	camoServer, err := New(camoConfig)
	if err != nil {
		return nil, fmt.Errorf("Error building Camo: %s", err.Error())
	}

	router := &router.DumbRouter{
		AddHeaders:  map[string]string{"X-Go-Camo": "test"},
		ServerName:  camoConfig.ServerName,
		CamoHandler: camoServer,
	}

	record := httptest.NewRecorder()
	router.ServeHTTP(record, req)
	if got, want := record.Code, status; got != want {
		return record, fmt.Errorf("response code = %d, wanted %d", got, want)
	}
	return record, nil
}

func makeTestReq(testURL string, status int) (*httptest.ResponseRecorder, error) {
	req, err := makeReq(testURL)
	if err != nil {
		return nil, err
	}
	record, err := processRequest(req, status, camoConfig)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func TestNotFound(t *testing.T) {
	t.Parallel()
	req, err := http.NewRequest("GET", "http://example.com/favicon.ico", nil)
	assert.Nil(t, err)

	record, err := processRequest(req, 404, camoConfig)
	if assert.Nil(t, err) {
		assert.Equal(t, 404, record.Code, "Expected 404 but got '%d' instead", record.Code)
		assert.Equal(t, "404 Not Found\n", record.Body.String(), "Expected 404 response body but got '%s' instead", record.Body.String())
		// validate headers
		assert.Equal(t, "test", record.HeaderMap.Get("X-Go-Camo"), "Expected custom response header not found")
		assert.Equal(t, "go-camo", record.HeaderMap.Get("Server"), "Expected 'Server' response header not found")
	}
}

func TestSimpleValidImageURL(t *testing.T) {
	t.Parallel()
	testURL := "http://www.google.com/images/srpr/logo11w.png"
	record, err := makeTestReq(testURL, 200)
	if assert.Nil(t, err) {
		// validate headers
		assert.Equal(t, "test", record.HeaderMap.Get("X-Go-Camo"), "Expected custom response header not found")
		assert.Equal(t, "go-camo", record.HeaderMap.Get("Server"), "Expected 'Server' response header not found")
	}
}

func TestGoogleChartURL(t *testing.T) {
	t.Parallel()
	testURL := "http://chart.apis.google.com/chart?chs=920x200&chxl=0:%7C2010-08-13%7C2010-09-12%7C2010-10-12%7C2010-11-11%7C1:%7C0%7C0%7C0%7C0%7C0%7C0&chm=B,EBF5FB,0,0,0&chco=008Cd6&chls=3,1,0&chg=8.3,20,1,4&chd=s:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA&chxt=x,y&cht=lc"
	_, err := makeTestReq(testURL, 200)
	assert.Nil(t, err)
}

func TestChunkedImageFile(t *testing.T) {
	t.Parallel()
	testURL := "http://www.igvita.com/posts/12/spdyproxy-diagram.png"
	_, err := makeTestReq(testURL, 200)
	assert.Nil(t, err)
}

func TestFollowRedirects(t *testing.T) {
	t.Parallel()
	testURL := "http://cl.ly/1K0X2Y2F1P0o3z140p0d/boom-headshot.gif"
	_, err := makeTestReq(testURL, 200)
	assert.Nil(t, err)
}

func TestStrangeFormatRedirects(t *testing.T) {
	t.Parallel()
	testURL := "http://cl.ly/DPcp/Screen%20Shot%202012-01-17%20at%203.42.32%20PM.png"
	_, err := makeTestReq(testURL, 200)
	assert.Nil(t, err)
}

func TestRedirectsWithPathOnly(t *testing.T) {
	t.Parallel()
	testURL := "http://httpbin.org/redirect-to?url=%2Fredirect-to%3Furl%3Dhttp%3A%2F%2Fwww.google.com%2Fimages%2Fsrpr%2Flogo11w.png"
	_, err := makeTestReq(testURL, 200)
	assert.Nil(t, err)
}

func TestFollowTempRedirects(t *testing.T) {
	t.Parallel()
	testURL := "http://httpbin.org/redirect-to?url=http://www.google.com/images/srpr/logo11w.png"
	_, err := makeTestReq(testURL, 200)
	assert.Nil(t, err)
}

func TestBadContentType(t *testing.T) {
	t.Parallel()
	testURL := "http://httpbin.org/response-headers?Content-Type=what"
	_, err := makeTestReq(testURL, 400)
	assert.Nil(t, err)
}

func Test404InfiniRedirect(t *testing.T) {
	t.Parallel()
	testURL := "http://httpbin.org/redirect/4"
	_, err := makeTestReq(testURL, 404)
	assert.Nil(t, err)
}

func Test404URLWithoutHTTPHost(t *testing.T) {
	t.Parallel()
	testURL := "/picture/Mincemeat/Pimp.jpg"
	_, err := makeTestReq(testURL, 404)
	assert.Nil(t, err)
}

func Test404ImageLargerThan5MB(t *testing.T) {
	t.Parallel()
	testURL := "http://apod.nasa.gov/apod/image/0505/larryslookout_spirit_big.jpg"
	_, err := makeTestReq(testURL, 404)
	assert.Nil(t, err)
}

func Test404HostNotFound(t *testing.T) {
	t.Parallel()
	testURL := "http://flabergasted.cx"
	_, err := makeTestReq(testURL, 404)
	assert.Nil(t, err)
}

func Test404OnExcludes(t *testing.T) {
	t.Parallel()
	testURL := "http://iphone.internal.example.org/foo.cgi"
	_, err := makeTestReq(testURL, 404)
	assert.Nil(t, err)
}

func Test404OnNonImageContent(t *testing.T) {
	t.Parallel()
	testURL := "https://github.com/atmos/cinderella/raw/master/bootstrap.sh"
	_, err := makeTestReq(testURL, 404)
	assert.Nil(t, err)
}

func Test404On10xIpRange(t *testing.T) {
	t.Parallel()
	testURL := "http://10.0.0.1/foo.cgi"
	_, err := makeTestReq(testURL, 404)
	assert.Nil(t, err)
}

func Test404On169Dot254Net(t *testing.T) {
	t.Parallel()
	testURL := "http://169.254.0.1/foo.cgi"
	_, err := makeTestReq(testURL, 404)
	assert.Nil(t, err)
}

func Test404On172Dot16Net(t *testing.T) {
	t.Parallel()
	for i := 16; i < 32; i++ {
		testURL := "http://172.%d.0.1/foo.cgi"
		_, err := makeTestReq(fmt.Sprintf(testURL, i), 404)
		assert.Nil(t, err)
	}
}

func Test404On192Dot168Net(t *testing.T) {
	t.Parallel()
	testURL := "http://192.168.0.1/foo.cgi"
	_, err := makeTestReq(testURL, 404)
	assert.Nil(t, err)
}

func Test404OnLocalhost(t *testing.T) {
	t.Parallel()
	testURL := "http://localhost/foo.cgi"
	record, err := makeTestReq(testURL, 404)
	if assert.Nil(t, err) {
		assert.Equal(t, "Bad url host\n", record.Body.String(), "Expected 404 response body but got '%s' instead", record.Body.String())
	}
}

// Test will fail if dns relay implements dns rebind prevention
func Test404OnLoopback(t *testing.T) {
	t.Parallel()
	testURL := "http://i.i.com.com/foo.cgi"
	record, err := makeTestReq(testURL, 404)
	if assert.Nil(t, err) {
		assert.Equal(t, "Denylist host failure\n", record.Body.String(), "Expected 404 response body but got '%s' instead", record.Body.String())
	}
}

func TestSupplyAcceptIfNoneGiven(t *testing.T) {
	t.Parallel()
	testURL := "http://images.anandtech.com/doci/6673/OpenMoboAMD30_575px.png"
	req, err := makeReq(testURL)
	req.Header.Del("Accept")
	assert.Nil(t, err)
	_, err = processRequest(req, 200, camoConfig)
	assert.Nil(t, err)
}

func TestTimeout(t *testing.T) {
	t.Parallel()
	c := Config{
		HMACKey:        []byte("0x24FEEDFACEDEADBEEFCAFE"),
		MaxSize:        5120 * 1024,
		RequestTimeout: time.Duration(500) * time.Millisecond,
		MaxRedirects:   3,
		ServerName:     "go-camo",
	}
	cc := make(chan bool, 1)
	received := make(chan bool)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received <- true
		<-cc
		r.Close = true
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	req, err := makeReq(ts.URL)
	assert.Nil(t, err)

	errc := make(chan error, 1)
	go func() {
		code := 504
		_, err := processRequest(req, code, c)
		errc <- err
	}()

	<-received
	select {
	case e := <-errc:
		assert.Nil(t, e)
		cc <- true
	case <-time.After(1 * time.Second):
		cc <- true
		t.Errorf("timeout didn't fire in time")
	}
	close(cc)
}
