// Copyright (c) 2012-2018 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package camo

import (
	"net"
	"os"
	"sync"
	"syscall"
)

func isBrokenPipe(err error) bool {
	if opErr, ok := err.(*net.OpError); ok {
		// >= go1.6
		if syscallErr, ok := opErr.Err.(*os.SyscallError); ok {
			switch syscallErr.Err {
			case syscall.EPIPE, syscall.ECONNRESET:
				return true
			default:
				return false
			}
		}

		// older go
		switch opErr.Err {
		case syscall.EPIPE, syscall.ECONNRESET:
			return true
		default:
			return false
		}
	}
	return false
}

func mustParseNetmask(s string) *net.IPNet {
	_, ipnet, err := net.ParseCIDR(s)
	if err != nil {
		panic(`misc: mustParseNetmask(` + s + `): ` + err.Error())
	}
	return ipnet
}

func mustParseNetmasks(networks []string) []*net.IPNet {
	nets := make([]*net.IPNet, 0)
	for _, s := range networks {
		ipnet := mustParseNetmask(s)
		nets = append(nets, ipnet)
	}
	return nets
}

func isRejectedIP(ip net.IP) bool {
	if !ip.IsGlobalUnicast() {
		return true
	}

	// test whether address is ipv4 or ipv6, to pick the proper filter list
	// (otherwise address may be 16 byte representation in go but not an actual
	// ipv6 address. this also helps avoid accidentally matching the
	// "::ffff:0:0/96" netblock
	checker := rejectIPv4Networks
	if ip.To4() == nil {
		checker = rejectIPv6Networks
	}

	for _, ipnet := range checker {
		if ipnet.Contains(ip) {
			return true
		}
	}

	return false
}

var bufPool = sync.Pool{
	New: func() interface{} {
		// note: 32 * 1024 is the size used by io.Copy by default.
		// Seems like a good starting point, just with a bit less garbage
		// (using a sync pool) to reduce some GC work.
		// ref: https://golang.org/src/io/io.go?s=13136:13214#L391
		buf := make([]byte, 32*1024)
		return &buf
	},
}
