package camoproxy

import (
	"fmt"
	"github.com/cactus/go-camo/camoproxy/encoding"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type testStruct struct {
	Desc   string
	Url    string
	Status int
}

var camoConfig = Config{
	HmacKey:        "0x24FEEDFACEDEADBEEFCAFE",
	MaxSize:        5120 * 1024,
	RequestTimeout: time.Duration(10) * time.Second,
	MaxRedirects:   3,
	ServerName:     "go-camo"}

func makeReq(testUrl string) (*http.Request, error) {
	k := []byte(camoConfig.HmacKey)
	hexurl := encoding.EncodeUrl(&k, testUrl)
	out := "http://example.com" + hexurl
	req, err := http.NewRequest("GET", out, nil)
	if err != nil {
		return nil, fmt.Errorf("Error building req url '%s': %s", testUrl, err.Error())
	}
	return req, nil
}

func processRequest(req *http.Request, status int) (*httptest.ResponseRecorder, error) {
	camoServer, err := New(camoConfig)
	if err != nil {
		return nil, fmt.Errorf("Error building CamoProxy: %s", err.Error())
	}

	router := mux.NewRouter()
	router.Handle("/{sigHash}/{encodedUrl}", camoServer).Methods("GET")

	record := httptest.NewRecorder()
	router.ServeHTTP(record, req)
	if got, want := record.Code, status; got != want {
		return nil, fmt.Errorf("response code = %d, wanted %d", got, want)
	}
	return record, nil
}

func makeTestReq(testUrl string, status int) (*httptest.ResponseRecorder, error) {
	req, err := makeReq(testUrl)
	if err != nil {
		return nil, err
	}
	record, err := processRequest(req, status)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func TestSimpleValidImageUrl(t *testing.T) {
	t.Parallel()
	testUrl := "http://media.ebaumsworld.com/picture/Mincemeat/Pimp.jpg"
	_, err := makeTestReq(testUrl, 200)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestGoogleChartUrl(t *testing.T) {
	t.Parallel()
	testUrl := "http://chart.apis.google.com/chart?chs=920x200&chxl=0:%7C2010-08-13%7C2010-09-12%7C2010-10-12%7C2010-11-11%7C1:%7C0%7C0%7C0%7C0%7C0%7C0&chm=B,EBF5FB,0,0,0&chco=008Cd6&chls=3,1,0&chg=8.3,20,1,4&chd=s:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA&chxt=x,y&cht=lc"
	_, err := makeTestReq(testUrl, 200)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestChunkedImageFile(t *testing.T) {
	t.Parallel()
	testUrl := "http://www.igvita.com/posts/12/spdyproxy-diagram.png"
	_, err := makeTestReq(testUrl, 200)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestFollowRedirects(t *testing.T) {
	t.Parallel()
	testUrl := "http://cl.ly/1K0X2Y2F1P0o3z140p0d/boom-headshot.gif"
	_, err := makeTestReq(testUrl, 200)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestStrangeFormatRedirects(t *testing.T) {
	t.Parallel()
	testUrl := "http://cl.ly/DPcp/Screen%20Shot%202012-01-17%20at%203.42.32%20PM.png"
	_, err := makeTestReq(testUrl, 200)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestRedirectsWithPathOnly(t *testing.T) {
	t.Parallel()
	testUrl := "http://blogs.msdn.com/photos/noahric/images/9948044/425x286.aspx"
	_, err := makeTestReq(testUrl, 200)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestFollowTempRedirects(t *testing.T) {
	t.Parallel()
	testUrl := "http://httpbin.org/redirect-to?url=http://www.google.com/images/srpr/logo11w.png"
	_, err := makeTestReq(testUrl, 200)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func Test404InfiniRedirect(t *testing.T) {
	t.Parallel()
	testUrl := "http://modeselektor.herokuapp.com/"
	_, err := makeTestReq(testUrl, 404)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func Test404UrlWithoutHTTPHost(t *testing.T) {
	t.Parallel()
	testUrl := "/picture/Mincemeat/Pimp.jpg"
	_, err := makeTestReq(testUrl, 404)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func Test404ImageLargerThan5MB(t *testing.T) {
	t.Parallel()
	testUrl := "http://apod.nasa.gov/apod/image/0505/larryslookout_spirit_big.jpg"
	_, err := makeTestReq(testUrl, 404)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func Test404HostNotFound(t *testing.T) {
	t.Parallel()
	testUrl := "http://flabergasted.cx"
	_, err := makeTestReq(testUrl, 404)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func Test404OnExcludes(t *testing.T) {
	t.Parallel()
	testUrl := "http://iphone.internal.example.org/foo.cgi"
	_, err := makeTestReq(testUrl, 404)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func Test404OnNonImageContent(t *testing.T) {
	t.Parallel()
	testUrl := "https://github.com/atmos/cinderella/raw/master/bootstrap.sh"
	_, err := makeTestReq(testUrl, 404)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func Test404On10xIpRange(t *testing.T) {
	t.Parallel()
	testUrl := "http://10.0.0.1/foo.cgi"
	_, err := makeTestReq(testUrl, 404)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func Test404On169Dot254Net(t *testing.T) {
	t.Parallel()
	testUrl := "http://169.254.0.1/foo.cgi"
	_, err := makeTestReq(testUrl, 404)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func Test404On172Dot16Net(t *testing.T) {
	t.Parallel()
	for i := 16; i < 32; i++ {
		testUrl := "http://172.%d.0.1/foo.cgi"
		_, err := makeTestReq(fmt.Sprintf(testUrl, i), 404)
		if err != nil {
			t.Errorf(err.Error())
		}
	}
}

func Test404On192Dot168Net(t *testing.T) {
	t.Parallel()
	testUrl := "http://192.168.0.1/foo.cgi"
	_, err := makeTestReq(testUrl, 404)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestSupplyAcceptIfNoneGiven(t *testing.T) {
	t.Parallel()
	testUrl := "http://images.anandtech.com/doci/6673/OpenMoboAMD30_575px.png"
	req, err := makeReq(testUrl)
	req.Header.Del("Accept")
	if err != nil {
		t.Errorf(err.Error())
	}
	_, err = processRequest(req, 200)
	if err != nil {
		t.Errorf(err.Error())
	}
}
