// Copyright (c) 2012-2016 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mlog

import (
	"io/ioutil"
	"testing"
)

func BenchmarkFormatWriterPlainBase(b *testing.B) {
	logger := New(ioutil.Discard, 0)
	logWriter := &FormatWriterPlain{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logWriter.Emit(logger, 0, "this is a test", nil)
	}
}

func BenchmarkFormatWriterPlainStd(b *testing.B) {
	logger := New(ioutil.Discard, Lstd)
	logWriter := &FormatWriterPlain{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logWriter.Emit(logger, 0, "this is a test", nil)
	}
}

func BenchmarkFormatWriterPlainTime(b *testing.B) {
	logger := New(ioutil.Discard, Ltimestamp)
	logWriter := &FormatWriterPlain{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logWriter.Emit(logger, 0, "this is a test", nil)
	}
}

func BenchmarkFormatWriterPlainShortfile(b *testing.B) {
	logger := New(ioutil.Discard, Lshortfile)
	logWriter := &FormatWriterPlain{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logWriter.Emit(logger, 0, "this is a test", nil)
	}
}

func BenchmarkFormatWriterPlainLongfile(b *testing.B) {
	logger := New(ioutil.Discard, Llongfile)
	logWriter := &FormatWriterPlain{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logWriter.Emit(logger, 0, "this is a test", nil)
	}
}

func BenchmarkFormatWriterPlainMap(b *testing.B) {
	logger := New(ioutil.Discard, 0)
	logWriter := &FormatWriterPlain{}
	m := Map{"x": 42}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logWriter.Emit(logger, 0, "this is a test", m)
	}
}

func BenchmarkFormatWriterPlainHugeMapUnsorted(b *testing.B) {
	logger := New(ioutil.Discard, 0)
	logWriter := &FormatWriterPlain{}
	m := Map{}
	for i := 1; i <= 100; i++ {
		m[randString(6, false)] = randString(10, false)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logWriter.Emit(logger, 0, "this is a test", m)
	}
}

func BenchmarkFormatWriterPlainHugeMapSorted(b *testing.B) {
	logger := New(ioutil.Discard, Lsort)
	logWriter := &FormatWriterPlain{}
	m := Map{}
	for i := 1; i <= 100; i++ {
		m[randString(6, false)] = randString(10, false)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logWriter.Emit(logger, 0, "this is a test", m)
	}
}
