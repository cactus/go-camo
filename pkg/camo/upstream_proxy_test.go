// Copyright (c) 2012-2023 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package camo

import (
	"fmt"
	"net"
	"sort"
	"testing"

	"codeberg.org/dropwhile/assert"
	"golang.org/x/net/http/httpproxy"
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
	assert.Equal(
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
	assert.Equal(
		t,
		addresses,
		[]string{
			"127.0.0.1",
			"::1",
		},
	)
}

func TestUpstreamProxyMatching(t *testing.T) {
	t.Parallel()

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

	f := func(matchtype string, address string, port string, result bool) {
		t.Helper()
		var tRes bool
		switch matchtype {
		case "ip":
			tRes = uspc.matchesIP(net.ParseIP(address), port)
		case "host":
			tRes = uspc.matchesHost(address, port)
		case "any":
			tRes = uspc.matchesAny(address, port)
		default:
			tRes = !result
		}
		assert.Equal(t, tRes, result,
			fmt.Sprintf("match-%s check failed (%s, %s, %t)",
				matchtype, address, port, result))
	}

	// host matches
	f("host", "localhost", "80", true)
	f("host", "localhost", "80", true)
	f("host", "localhostx", "80", false)
	f("host", "loopback", "80", false)
	f("host", "localhost", "90", false)
	f("host", "127.0.0.1", "80", false)
	f("host", "::1", "80", false)
	f("host", "127.0.0.1", "90", false)
	f("host", "127.0.0.2", "80", false)
	f("host", "::1", "90", false)
	f("host", "::2", "80", false)

	// ip matches
	f("ip", "127.0.0.1", "80", true)
	f("ip", "::1", "80", true)

	f("ip", "127.0.0.2", "80", false)
	f("ip", "127.0.0.1", "90", false)
	f("ip", "::2", "80", false)
	f("ip", "::1", "90", false)

	// any matches
	f("any", "127.0.0.1", "80", true)
	f("any", "::1", "80", true)
	f("any", "localhost", "80", true)
	f("any", "loopback", "80", false)
	f("any", "localhostx", "80", false)
	f("any", "localhost", "90", false)
	f("any", "127.0.0.2", "80", false)
	f("any", "127.0.0.1", "90", false)
	f("any", "1.1.1.1", "80", false)
	f("any", "255.255.255.255", "80", false)
	f("any", "::1", "90", false)
}
