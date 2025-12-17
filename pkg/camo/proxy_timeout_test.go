// Copyright (c) 2012-2023 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package camo

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"codeberg.org/dropwhile/assert"
	"codeberg.org/dropwhile/mlog"
	"github.com/cactus/go-camo/v2/pkg/camo/encoding"
	"github.com/cactus/go-camo/v2/pkg/router"
)

func TestTimeout(t *testing.T) {
	t.Parallel()
	c := Config{
		HMACKey:        []byte("0x24FEEDFACEDEADBEEFCAFE"),
		MaxSize:        5120 * 1024,
		RequestTimeout: time.Duration(500) * time.Millisecond,
		MaxRedirects:   3,
		ServerName:     "go-camo",
		noIPFiltering:  true,
	}
	cc := make(chan bool, 1)
	received := make(chan bool)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received <- true
		<-cc
		r.Close = true
		_, err := w.Write([]byte("ok"))
		assert.Nil(t, err)
	}))
	defer ts.Close()

	req, err := makeReq(c, ts.URL)
	assert.Nil(t, err)

	errc := make(chan error, 1)
	go func() {
		code := 504
		_, err := processRequest(req, code, c, nil)
		errc <- err
	}()

	select {
	case <-received:
		select {
		case e := <-errc:
			assert.Nil(t, e)
			cc <- true
		case <-time.After(1 * time.Second):
			cc <- true
			t.Errorf("timeout didn't fire in time")
		}
	case <-time.After(1 * time.Second):
		var err error
		select {
		case e := <-errc:
			err = e
		default:
		}
		if err != nil {
			assert.Nil(t, err, "test didn't hit backend as expected")
		}
		t.Errorf("test didn't hit backend as expected")
	}

	close(cc)
}

func TestClientCancelEarly(t *testing.T) {
	t.Parallel()
	c := Config{
		HMACKey:        []byte("0x24FEEDFACEDEADBEEFCAFE"),
		MaxSize:        5120 * 1024,
		RequestTimeout: time.Duration(500) * time.Millisecond,
		MaxRedirects:   3,
		ServerName:     "go-camo",
		noIPFiltering:  true,
	}

	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "image/png")
			w.Header().Set("Connection", "close")
			flusher, ok := w.(http.Flusher)
			assert.True(t, ok)
			for i := 1; i <= 500; i++ {
				_, err := fmt.Fprintf(w, "Chunk #%d\n", i)
				// conn closed/broken pipe
				if err != nil {
					mlog.Debugx("write error", mlog.A("err", err), mlog.A("i", i))
					break
				}
				flusher.Flush() // Trigger "chunked" encoding and send a chunk...
			}
		},
	))
	defer ts.Close()

	camoServer, err := New(c)
	assert.Nil(t, err)
	router := &router.DumbRouter{
		ServerName:  c.ServerName,
		CamoHandler: camoServer,
	}

	tsCamo := httptest.NewServer(router)
	defer tsCamo.Close()

	conn, err := net.Dial("tcp", tsCamo.Listener.Addr().String())
	assert.Nil(t, err)
	defer conn.Close()

	req := fmt.Appendf(nil,
		"GET %s HTTP/1.1\r\nHost: foo.com\r\nConnection: close\r\n\r\n",
		encoding.B64EncodeURL(c.HMACKey, ts.URL+"/image.png"),
	)
	_, err = conn.Write(req)
	assert.Nil(t, err)
	conn.Close()
	time.Sleep(100 * time.Millisecond)
	// fmt.Printf("done\n")
}

func TestClientCancelLate(t *testing.T) {
	t.Parallel()
	c := Config{
		HMACKey:        []byte("0x24FEEDFACEDEADBEEFCAFE"),
		MaxSize:        5120 * 1024,
		RequestTimeout: time.Duration(500) * time.Millisecond,
		MaxRedirects:   3,
		ServerName:     "go-camo",
		noIPFiltering:  true,
	}

	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "image/png")
			w.Header().Set("Connection", "close")
			flusher, ok := w.(http.Flusher)
			assert.True(t, ok)
			for i := 1; i <= 500; i++ {
				_, err := fmt.Fprintf(w, "Chunk #%d\n", i)
				// conn closed/broken pipe
				if err != nil {
					mlog.Debugx("write error", mlog.A("err", err), mlog.A("i", i))
					break
				}
				flusher.Flush() // Trigger "chunked" encoding and send a chunk...
			}
		},
	))
	defer ts.Close()

	camoServer, err := New(c)
	assert.Nil(t, err)
	router := &router.DumbRouter{
		ServerName:  c.ServerName,
		CamoHandler: camoServer,
	}

	tsCamo := httptest.NewServer(router)
	defer tsCamo.Close()

	conn, err := net.Dial("tcp", tsCamo.Listener.Addr().String())
	assert.Nil(t, err)
	defer conn.Close()

	req := fmt.Appendf(nil,
		"GET %s HTTP/1.1\r\nHost: foo.com\r\nConnection: close\r\n\r\n",
		encoding.B64EncodeURL(c.HMACKey, ts.URL+"/image.png"),
	)
	_, err = conn.Write(req)
	assert.Nil(t, err)

	// partial read
	cReader := bufio.NewReaderSize(conn, 32)
	for {
		data, err := cReader.ReadBytes('\n')
		assert.Nil(t, err)
		if bytes.Contains(data, []byte("Chunk #2")) {
			break
		} else if bytes.Contains(data, []byte("404 Not Found")) {
			fmt.Printf("got 404!\n")
			for {
				d, err := cReader.ReadBytes('\n')
				if err == io.EOF {
					mlog.Debug("got eof")
					break
				}
				assert.Nil(t, err)
				mlog.Debugf("got: %s", string(d))
			}
			break
		} else {
			mlog.Debugf("data: %s", string(data))
		}
	}
	conn.Close()
	// fmt.Printf("done\n")
}

func TestServerEarlyEOF(t *testing.T) {
	t.Parallel()
	c := Config{
		HMACKey:        []byte("0x24FEEDFACEDEADBEEFCAFE"),
		MaxSize:        5120 * 1024,
		RequestTimeout: time.Duration(500) * time.Millisecond,
		MaxRedirects:   3,
		ServerName:     "go-camo",
		noIPFiltering:  true,
	}

	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "image/png")
			w.Header().Set("Connection", "close")
			w.Header().Set("Content-Length", "100")
			w.WriteHeader(200)
		},
	))
	defer ts.Close()

	req, err := makeReq(c, ts.URL)
	assert.Nil(t, err)
	// response is a 200, not much we can do about that since we response
	// streaming (chunked)...
	resp, err := processRequest(req, 200, c, nil)
	assert.Nil(t, err)

	body, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, len(body), 0)
}

func TestServerChunkTooBig(t *testing.T) {
	t.Parallel()
	c := Config{
		HMACKey:        []byte("0x24FEEDFACEDEADBEEFCAFE"),
		MaxSize:        1024,
		RequestTimeout: time.Duration(500) * time.Millisecond,
		MaxRedirects:   3,
		ServerName:     "go-camo",
		noIPFiltering:  true,
	}

	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "image/png")
			w.Header().Set("Connection", "close")
			flusher, ok := w.(http.Flusher)
			assert.True(t, ok)
			for i := 1; i <= 500; i++ {
				// all done
				if r.Context().Err() != nil {
					// camo aborted reading the rest, we're done!
					return
				}
				_, err := fmt.Fprintf(w, "Chunk #%d\n", i)
				if err != nil {
					assert.Nil(t, err)
					break
				}
				flusher.Flush() // Trigger "chunked" encoding and send a chunk...
			}
		},
	))
	defer ts.Close()

	req, err := makeReq(c, ts.URL)
	assert.Nil(t, err)
	// response is a 200, not much we can do about that since we response
	// streaming (chunked)...
	resp, err := processRequest(req, 200, c, nil)
	assert.Nil(t, err)

	// partial read
	cReader := bufio.NewReaderSize(resp.Body, 100)
	total := 0
	for {
		discarded, err := cReader.Discard(100)
		total += discarded
		if err == io.EOF {
			break
		}
		assert.Nil(t, err)
	}
	// at least we should have only read the MaxSize amount...
	assert.Equal(t, total, 1024)
}
