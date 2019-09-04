// Copyright (c) 2012-2018 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package camo

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
		w.Write([]byte("ok"))

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
