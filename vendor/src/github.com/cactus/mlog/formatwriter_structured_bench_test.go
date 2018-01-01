// Copyright (c) 2012-2016 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mlog

import (
	"io/ioutil"
	"testing"
)

func BenchmarkFormatWriterStructuredBase(b *testing.B) {
	logger := New(ioutil.Discard, 0)
	logWriter := &FormatWriterStructured{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logWriter.Emit(logger, 0, "this is a test", nil)
	}
}

func BenchmarkFormatWriterStructuredStd(b *testing.B) {
	logger := New(ioutil.Discard, Lstd)
	logWriter := &FormatWriterStructured{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logWriter.Emit(logger, 0, "this is a test", nil)
	}
}

func BenchmarkFormatWriterStructuredTime(b *testing.B) {
	logger := New(ioutil.Discard, Ltimestamp)
	logWriter := &FormatWriterStructured{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logWriter.Emit(logger, 0, "this is a test", nil)
	}
}

func BenchmarkFormatWriterStructuredTimeTAI64N(b *testing.B) {
	logger := New(ioutil.Discard, Ltai64n)
	logWriter := &FormatWriterStructured{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logWriter.Emit(logger, 0, "this is a test", nil)
	}
}

func BenchmarkFormatWriterStructuredShortfile(b *testing.B) {
	logger := New(ioutil.Discard, Lshortfile)
	logWriter := &FormatWriterStructured{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logWriter.Emit(logger, 0, "this is a test", nil)
	}
}

func BenchmarkFormatWriterStructuredLongfile(b *testing.B) {
	logger := New(ioutil.Discard, Llongfile)
	logWriter := &FormatWriterStructured{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logWriter.Emit(logger, 0, "this is a test", nil)
	}
}

func BenchmarkFormatWriterStructuredMap(b *testing.B) {
	logger := New(ioutil.Discard, 0)
	logWriter := &FormatWriterStructured{}
	m := Map{"x": 42}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logWriter.Emit(logger, 0, "this is a test", m)
	}
}

func BenchmarkFormatWriterStructuredHugeMapUnsorted(b *testing.B) {
	logger := New(ioutil.Discard, 0)
	logWriter := &FormatWriterStructured{}
	m := Map{}
	for i := 1; i <= 100; i++ {
		m[randString(6, false)] = randString(10, false)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logWriter.Emit(logger, 0, "this is a test", m)
	}
}

func BenchmarkFormatWriterStructuredHugeMapSorted(b *testing.B) {
	logger := New(ioutil.Discard, Lsort)
	logWriter := &FormatWriterStructured{}
	m := Map{}
	for i := 1; i <= 100; i++ {
		m[randString(6, false)] = randString(10, false)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logWriter.Emit(logger, 0, "this is a test", m)
	}
}
