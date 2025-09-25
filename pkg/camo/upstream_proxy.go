// Copyright (c) 2012-2023 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package camo provides an HTTP proxy server with content type
// restrictions as well as regex host allow list support.
package camo

import (
	"fmt"
	"net"
	"net/url"

	"golang.org/x/net/http/httpproxy"
)

type innerUpstreamProxyConfig struct {
	scheme    string
	host      string
	port      string
	addresses []net.IP
}

func (ic *innerUpstreamProxyConfig) matchesIP(ip net.IP, port string) bool {
	if ic == nil {
		return false
	}

	if ic.port != port {
		return false
	}

	if len(ic.addresses) == 0 {
		return false
	}

	for i := range ic.addresses {
		if ip.Equal(ic.addresses[i]) {
			return true
		}
	}
	return false
}

func (ic *innerUpstreamProxyConfig) matchesHost(host string, port string) bool {
	if ic == nil {
		return false
	}

	if ic.port != port {
		return false
	}

	if ic.host == host {
		return true
	}
	return false
}

type upstreamProxyConfig struct {
	httpProxy  *innerUpstreamProxyConfig
	httpsProxy *innerUpstreamProxyConfig
	hasProxy   bool
}

func (c *upstreamProxyConfig) matchesIP(ip net.IP, port string) bool {
	if c == nil || !c.hasProxy {
		return false
	}

	if c.httpProxy != nil && c.httpProxy.matchesIP(ip, port) {
		return true
	}
	if c.httpsProxy != nil && c.httpsProxy.matchesIP(ip, port) {
		return true
	}
	return false
}

func (c *upstreamProxyConfig) matchesHost(host string, port string) bool {
	if c == nil || !c.hasProxy {
		return false
	}

	if c.httpProxy != nil && c.httpProxy.matchesHost(host, port) {
		return true
	}
	if c.httpsProxy != nil && c.httpsProxy.matchesHost(host, port) {
		return true
	}
	return false
}

func (c *upstreamProxyConfig) matchesAny(address string, port string) bool {
	if c == nil || !c.hasProxy {
		return false
	}

	if port == "" {
		port = "80"
	}

	if c.matchesHost(address, port) {
		return true
	}

	if ips, err := hostnameToIPs(address); err == nil {
		for _, ip := range ips {
			if c.matchesIP(ip, port) {
				return true
			}
		}
	}

	return false
}

func (c *upstreamProxyConfig) parseProxy(proxy string) (*innerUpstreamProxyConfig, error) {
	// parts of this function from golang stdlib
	//   http/httproxy/proxy.go
	//
	// Copyright 2017 The Go Authors. All rights reserved.
	// Use of this source code is governed by a BSD-style
	// license that can be found in the LICENSE file here:
	//   https://github.com/golang/go/blob/master/LICENSE
	if proxy == "" {
		return nil, nil
	}

	proxyURL, err := url.Parse(proxy)
	if err != nil ||
		(proxyURL.Scheme != "http" &&
			proxyURL.Scheme != "https" &&
			proxyURL.Scheme != "socks5") {
		// proxy was bogus. Try prepending "http://" to it and
		// see if that parses correctly. If not, we fall
		// through and complain about the original one.
		proxyURL, err = url.Parse("http://" + proxy)
	}
	if err != nil {
		return nil, fmt.Errorf("invalid proxy address %q: %v", proxy, err)
	}

	ic := &innerUpstreamProxyConfig{
		scheme: proxyURL.Scheme,
		host:   proxyURL.Hostname(),
		port:   proxyURL.Port(),
	}

	if ic.port == "" {
		ic.port = "80"
	}

	if ip := net.ParseIP(ic.host); ip != nil {
		ic.addresses = append(ic.addresses, ip)
	} else {
		if ips, err := net.LookupIP(ic.host); err == nil {
			ic.addresses = append(ic.addresses, ips...)
		}
	}

	return ic, nil
}

func (c *upstreamProxyConfig) initFromConfig(config *httpproxy.Config) {
	if parsed, err := c.parseProxy(config.HTTPProxy); err == nil {
		c.httpProxy = parsed
		c.hasProxy = true
	}
	if parsed, err := c.parseProxy(config.HTTPSProxy); err == nil {
		c.httpsProxy = parsed
		c.hasProxy = true
	}
}

func (c *upstreamProxyConfig) init() {
	c.initFromConfig(httpproxy.FromEnvironment())
}
