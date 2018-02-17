// Copyright (c) 2012-2016 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mlog

import (
	"io/ioutil"
	"log"
	"math/rand"
	"testing"
)

const (
	letterBytes    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterBytesAlt = letterBytes + "\"\t\r\n"
	letterIdxBits  = 6                    // 6 bits to represent a letter index
	letterIdxMask  = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax   = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// uses unseeded rand (seed(1))...only use for testing!
func randString(n int, altchars bool) string {
	lb := letterBytes
	if altchars {
		lb = letterBytesAlt
	}
	b := make([]byte, n)
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(lb) {
			b[i] = lb[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func BenchmarkLoggingDebugWithDisabled(b *testing.B) {
	logger := New(ioutil.Discard, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Debug("this is a test")
	}
}

func BenchmarkLoggingDebugWithEnabled(b *testing.B) {
	logger := New(ioutil.Discard, Ldebug)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Debug("this is a test")
	}
}

func BenchmarkLoggingLikeStdlib(b *testing.B) {
	logger := New(ioutil.Discard, Lstd)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("this is a test")
	}
}

func BenchmarkLoggingStdlibLog(b *testing.B) {
	logger := log.New(ioutil.Discard, "info: ", log.LstdFlags)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Print("this is a test")
	}
}

func BenchmarkLoggingLikeStdlibShortfile(b *testing.B) {
	logger := New(ioutil.Discard, Lstd|Lshortfile)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("this is a test")
	}
}

func BenchmarkLoggingStdlibLogShortfile(b *testing.B) {
	logger := log.New(ioutil.Discard, "info: ", log.LstdFlags|log.Lshortfile)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Print("this is a test")
	}
}

func BenchmarkLoggingParallelLikeStdlib(b *testing.B) {
	logger := New(ioutil.Discard, Lstd)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("this is a test")
		}
	})
}

func BenchmarkLoggingParallelStdlibLog(b *testing.B) {
	logger := log.New(ioutil.Discard, "info: ", log.LstdFlags)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Print("this is a test")
		}
	})
}
