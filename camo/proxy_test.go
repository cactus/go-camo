package camo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cactus/go-camo/camo/encoding"
	"github.com/gorilla/mux"
)

var camoConfig = Config{
	HMACKey:        []byte("0x24FEEDFACEDEADBEEFCAFE"),
	MaxSize:        5120 * 1024,
	RequestTimeout: time.Duration(10) * time.Second,
	MaxRedirects:   3,
	ServerName:     "go-camo",
	AddHeaders:     map[string]string{"X-Go-Camo": "test"},
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

func processRequest(req *http.Request, status int) (*httptest.ResponseRecorder, error) {
	camoServer, err := New(camoConfig)
	if err != nil {
		return nil, fmt.Errorf("Error building Camo: %s", err.Error())
	}

	router := mux.NewRouter()
	router.NotFoundHandler = camoServer.NotFoundHandler()
	router.Handle("/{sigHash}/{encodedURL}", camoServer).Methods("GET")

	record := httptest.NewRecorder()
	router.ServeHTTP(record, req)
	if got, want := record.Code, status; got != want {
		return nil, fmt.Errorf("response code = %d, wanted %d", got, want)
	}
	return record, nil
}

func makeTestReq(testURL string, status int) (*httptest.ResponseRecorder, error) {
	req, err := makeReq(testURL)
	if err != nil {
		return nil, err
	}
	record, err := processRequest(req, status)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func TestNotFound(t *testing.T) {
	t.Parallel()
	req, err := http.NewRequest("GET", "http://example.com/favicon.ico", nil)
	if err != nil {
		t.Error(err.Error())
	}
	record, err := processRequest(req, 404)
	if err != nil {
		t.Error(err.Error())
	}
	if record.Code != 404 {
		t.Errorf("Expected 404 but got '%d' instead", record.Code)
	}
	if record.Body.String() != "404 Not Found\n" {
		t.Errorf("Expected 404 response body but got '%s' instead", record.Body.String())
	}
	// validate headers
	if record.HeaderMap.Get("X-Go-Camo") != "test" {
		t.Error("Expected custom response header not found")
	}
	if record.HeaderMap.Get("Server") != "go-camo" {
		t.Error("Expected 'Server' response header not found")
	}
}

func TestSimpleValidImageURL(t *testing.T) {
	t.Parallel()
	testURL := "http://media.ebaumsworld.com/picture/Mincemeat/Pimp.jpg"
	record, err := makeTestReq(testURL, 200)
	if err != nil {
		t.Error(err.Error())
	}
	// validate headers
	if record.HeaderMap.Get("X-Go-Camo") != "test" {
		t.Error("Expected custom response header not found")
	}
	if record.HeaderMap.Get("Server") != "go-camo" {
		t.Error("Expected 'Server' response header not found")
	}
}

func TestGoogleChartURL(t *testing.T) {
	t.Parallel()
	testURL := "http://chart.apis.google.com/chart?chs=920x200&chxl=0:%7C2010-08-13%7C2010-09-12%7C2010-10-12%7C2010-11-11%7C1:%7C0%7C0%7C0%7C0%7C0%7C0&chm=B,EBF5FB,0,0,0&chco=008Cd6&chls=3,1,0&chg=8.3,20,1,4&chd=s:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA&chxt=x,y&cht=lc"
	_, err := makeTestReq(testURL, 200)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestChunkedImageFile(t *testing.T) {
	t.Parallel()
	testURL := "http://www.igvita.com/posts/12/spdyproxy-diagram.png"
	_, err := makeTestReq(testURL, 200)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestFollowRedirects(t *testing.T) {
	t.Parallel()
	testURL := "http://cl.ly/1K0X2Y2F1P0o3z140p0d/boom-headshot.gif"
	_, err := makeTestReq(testURL, 200)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestStrangeFormatRedirects(t *testing.T) {
	t.Parallel()
	testURL := "http://cl.ly/DPcp/Screen%20Shot%202012-01-17%20at%203.42.32%20PM.png"
	_, err := makeTestReq(testURL, 200)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestRedirectsWithPathOnly(t *testing.T) {
	t.Parallel()
	testURL := "http://blogs.msdn.com/photos/noahric/images/9948044/425x286.aspx"
	_, err := makeTestReq(testURL, 200)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestFollowTempRedirects(t *testing.T) {
	t.Parallel()
	testURL := "http://httpbin.org/redirect-to?url=http://www.google.com/images/srpr/logo11w.png"
	_, err := makeTestReq(testURL, 200)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestBadContentType(t *testing.T) {
	t.Parallel()
	testURL := "http://httpbin.org/response-headers?Content-Type=what"
	_, err := makeTestReq(testURL, 400)
	if err != nil {
		t.Error(err.Error())
	}
}

func Test404InfiniRedirect(t *testing.T) {
	t.Parallel()
	testURL := "http://httpbin.org/redirect/4"
	_, err := makeTestReq(testURL, 404)
	if err != nil {
		t.Error(err.Error())
	}
}

func Test404URLWithoutHTTPHost(t *testing.T) {
	t.Parallel()
	testURL := "/picture/Mincemeat/Pimp.jpg"
	_, err := makeTestReq(testURL, 404)
	if err != nil {
		t.Error(err.Error())
	}
}

func Test404ImageLargerThan5MB(t *testing.T) {
	t.Parallel()
	testURL := "http://apod.nasa.gov/apod/image/0505/larryslookout_spirit_big.jpg"
	_, err := makeTestReq(testURL, 404)
	if err != nil {
		t.Error(err.Error())
	}
}

func Test404HostNotFound(t *testing.T) {
	t.Parallel()
	testURL := "http://flabergasted.cx"
	_, err := makeTestReq(testURL, 404)
	if err != nil {
		t.Error(err.Error())
	}
}

func Test404OnExcludes(t *testing.T) {
	t.Parallel()
	testURL := "http://iphone.internal.example.org/foo.cgi"
	_, err := makeTestReq(testURL, 404)
	if err != nil {
		t.Error(err.Error())
	}
}

func Test404OnNonImageContent(t *testing.T) {
	t.Parallel()
	testURL := "https://github.com/atmos/cinderella/raw/master/bootstrap.sh"
	_, err := makeTestReq(testURL, 404)
	if err != nil {
		t.Error(err.Error())
	}
}

func Test404On10xIpRange(t *testing.T) {
	t.Parallel()
	testURL := "http://10.0.0.1/foo.cgi"
	_, err := makeTestReq(testURL, 404)
	if err != nil {
		t.Error(err.Error())
	}
}

func Test404On169Dot254Net(t *testing.T) {
	t.Parallel()
	testURL := "http://169.254.0.1/foo.cgi"
	_, err := makeTestReq(testURL, 404)
	if err != nil {
		t.Error(err.Error())
	}
}

func Test404On172Dot16Net(t *testing.T) {
	t.Parallel()
	for i := 16; i < 32; i++ {
		testURL := "http://172.%d.0.1/foo.cgi"
		_, err := makeTestReq(fmt.Sprintf(testURL, i), 404)
		if err != nil {
			t.Error(err.Error())
		}
	}
}

func Test404On192Dot168Net(t *testing.T) {
	t.Parallel()
	testURL := "http://192.168.0.1/foo.cgi"
	_, err := makeTestReq(testURL, 404)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestSupplyAcceptIfNoneGiven(t *testing.T) {
	t.Parallel()
	testURL := "http://images.anandtech.com/doci/6673/OpenMoboAMD30_575px.png"
	req, err := makeReq(testURL)
	req.Header.Del("Accept")
	if err != nil {
		t.Error(err.Error())
	}
	_, err = processRequest(req, 200)
	if err != nil {
		t.Error(err.Error())
	}
}
