// Copyright (c) 2012-2023 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package camo

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"unicode"
)

type LimitReadCloser struct {
	io.Reader
	io.Closer
}

func (l *LimitReadCloser) Read(p []byte) (int, error) {
	return l.Reader.Read(p)
}

func NewLimitReadCloser(r io.ReadCloser, n int64) *LimitReadCloser {
	return &LimitReadCloser{
		Reader: io.LimitReader(r, n),
		Closer: r,
	}
}

type sizedChunkWriter struct {
	flusher      http.Flusher
	dst          io.Writer
	flushCounter int
	// mu sync.Mutex
}

func (scw *sizedChunkWriter) Write(p []byte) (n int, err error) {
	//scw.mu.Lock()
	//defer scw.mu.Unlock()

	n, err = scw.dst.Write(p)
	if err != nil {
		return n, err
	}

	scw.flushCounter += n
	if scw.flushCounter >= bufSize {
		scw.flushCounter = 0
		scw.flusher.Flush()
	}
	return
}

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

func getMaxRangeByte(rangeReq string) int64 {
	// format: bytes=0-9,33,34-99
	rangeReq = strings.Join(strings.Fields(rangeReq), "")

	// malformed request
	if len(rangeReq) <= 6 || rangeReq[5] != '=' {
		return -1
	}

	rr := rangeReq[6:]
	maxSeen := int64(0)
	for elem := range strings.FieldsFuncSeq(rr, func(r rune) bool {
		return unicode.IsSpace(r) || r == ','
	}) {
		// skip negative offsets, as we don't know
		// how long the request actually would be.
		if len(elem) > 0 && elem[0] == '-' {
			return -1
		}
		for part := range strings.SplitSeq(elem, "-") {
			if n, err := strconv.ParseInt(part, 10, 64); err == nil {
				maxSeen = max(maxSeen, n)
			}
		}
	}
	return maxSeen
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

func containsOneOf(s string, substrs ...string) bool {
	for i := range substrs {
		if strings.Contains(s, substrs[i]) {
			return true
		}
	}
	return false
}

func hostnameToIPs(hostname string) ([]net.IP, error) {
	if ip := net.ParseIP(hostname); ip != nil {
		return []net.IP{ip}, nil
	} else {
		if ips, err := net.LookupIP(hostname); err == nil {
			return ips, nil
		}
	}
	return nil, fmt.Errorf("no ips for hostname %s", hostname)
}

const bufSize = 32 * 1024

var bufPool = sync.Pool{
	New: func() any {
		// note: 32 * 1024 is the size used by io.Copy by default.
		// Seems like a good starting point, just with a bit less garbage
		// (using a sync pool) to reduce some GC work.
		// ref: https://golang.org/src/io/io.go?s=13136:13214#L391
		buf := make([]byte, bufSize)
		return &buf
	},
}
