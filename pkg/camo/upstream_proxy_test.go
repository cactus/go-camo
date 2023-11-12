// Copyright (c) 2012-2023 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package camo

import (
	"fmt"
	"net"
	"sort"
	"testing"

	"golang.org/x/net/http/httpproxy"
	"gotest.tools/v3/assert"
)

func IPListToStringList(sliceList []net.IP) []string {
	list := make([]string, 0, len(sliceList))
	for _, item := range sliceList {
		list = append(list, item.String())
	}
	return list
}

func TestUpstreamProxyParsing(t *testing.T) {
	t.Parallel()

	config := &httpproxy.Config{
		HTTPProxy:  "127.0.0.1",
		HTTPSProxy: "socks5://localhost:9999",
		NoProxy:    "",
		CGI:        false,
	}
	uspc := &upstreamProxyConfig{}
	uspc.initFromConfig(config)
	assert.Equal(t, uspc.hasProxy, true)
	assert.Equal(t, uspc.httpProxy.scheme, "http")
	assert.Equal(t, uspc.httpProxy.host, "127.0.0.1")
	assert.Equal(t, uspc.httpProxy.port, "80")
	assert.DeepEqual(
		t,
		uspc.httpProxy.addresses,
		[]net.IP{
			net.ParseIP("127.0.0.1"),
		},
	)
	assert.Equal(t, uspc.httpsProxy.scheme, "socks5")
	assert.Equal(t, uspc.httpsProxy.host, "localhost")
	assert.Equal(t, uspc.httpsProxy.port, "9999")

	addresses := IPListToStringList(uspc.httpsProxy.addresses)
	sort.Strings(addresses)
	assert.DeepEqual(
		t,
		addresses,
		[]string{
			"127.0.0.1",
			"::1",
		},
	)
}

func TestUpstreamProxyMatching(t *testing.T) {
	uspc := &upstreamProxyConfig{
		hasProxy: true,
		httpProxy: &innerUpstreamProxyConfig{
			addresses: []net.IP{
				net.ParseIP("::1"),
				net.ParseIP("127.0.0.1"),
			},
			scheme: "http",
			host:   "localhost",
			port:   "80",
		},
	}

	// matches host
	testElemsHost := []struct {
		matchtype string
		address   string
		port      string
		result    bool
	}{
		// host matches
		{"host", "localhost", "80", true},
		{"host", "localhost", "80", true},

		{"host", "localhostx", "80", false},
		{"host", "loopback", "80", false},
		{"host", "localhost", "90", false},
		{"host", "127.0.0.1", "80", false},
		{"host", "::1", "80", false},
		{"host", "127.0.0.1", "90", false},
		{"host", "127.0.0.2", "80", false},
		{"host", "::1", "90", false},
		{"host", "::2", "80", false},

		// ip matches
		{"ip", "127.0.0.1", "80", true},
		{"ip", "::1", "80", true},

		{"ip", "127.0.0.2", "80", false},
		{"ip", "127.0.0.1", "90", false},
		{"ip", "::2", "80", false},
		{"ip", "::1", "90", false},

		// any matches
		{"any", "127.0.0.1", "80", true},
		{"any", "::1", "80", true},
		{"any", "localhost", "80", true},

		{"any", "loopback", "80", false},
		{"any", "localhostx", "80", false},
		{"any", "localhost", "90", false},
		{"any", "127.0.0.2", "80", false},
		{"any", "127.0.0.1", "90", false},
		{"any", "1.1.1.1", "80", false},
		{"any", "255.255.255.255", "80", false},
		{"any", "::1", "90", false},
	}

	errMsg := func(a string, b string, c string, d bool) string {
		return fmt.Sprintf("match-%s check failed (%s, %s, %t)", a, b, c, d)
	}

	for _, elem := range testElemsHost {
		var tRes bool
		switch elem.matchtype {
		case "ip":
			tRes = uspc.matchesIP(net.ParseIP(elem.address), elem.port)
		case "host":
			tRes = uspc.matchesHost(elem.address, elem.port)
		case "any":
			tRes = uspc.matchesAny(elem.address, elem.port)
		default:
			tRes = !elem.result
		}
		assert.Check(
			t, tRes == elem.result,
			errMsg(elem.matchtype, elem.address, elem.port, elem.result),
		)
	}
}
