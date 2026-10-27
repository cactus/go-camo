// Copyright (c) 2012-2023 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package htrie

import (
	"fmt"
	"net/url"
	"testing"

	"codeberg.org/dropwhile/go-droplibs/assert"
)

func TestGlobPathChecker(t *testing.T) {
	t.Parallel()

	rules := []string{
		"|i|*/test.png",
		"||/hodor/test.png",
		"||/hodor/test.png.longer",
		"||/hodor/bar*",
		"||/hodor/ütest.png",
		"||/no*/to/s*/here",
		"||/i/can/s*/it*",
		"||/play/*/ball/img.png",
		"||/yalp*llab/img.png",
	}

	gpc := NewGlobPathChecker()
	for _, rule := range rules {
		err := gpc.AddRule(rule)
		assert.Nil(t, err)
	}

	f := func(s string, expected bool) {
		t.Helper()
		u, _ := url.Parse(s)
		msg := "should NOT match"
		if expected {
			msg = "should match"
		}
		assert.Equal(
			t,
			gpc.CheckPath(CleanPath(u)),
			expected,
			fmt.Sprintf("%s: %s", msg, s),
		)
	}

	// fmt.Println(gpc.RenderTree())

	// test match
	f("http://bar.example.com/foo/TEST.png", true)
	f("http://example.org/foo/test.png", true)
	f("http://example.org/hodor/test.png", true)
	f("http://example.org/hodor/ütest.png", true)
	f("http://example.org/hodor/test.png.longer", true)
	f("http://example.org/hodor/bartholemew", true)
	f("http://example.org/hodor/bart/homer.png", true)
	f("http://example.net/nothing/to/see/here", true)
	f("http://example.net/i/can/see/it/in/the/clouds/file.png", true)
	f("http://example.org/play/base/ball/img.png", true)
	f("http://example.org/yalp/base/llab/img.png", true)
	f("http://example.org/yalpllab/img.png", true)

	// test no match
	f("http://bar.example.com/foo/testx.png", false)
	f("http://example.net/something/to/see/here/file.png", false)
	f("http://example.org/hodor/test.png.long", false)
}

func TestGlobPathCheckerEscapes(t *testing.T) {
	t.Parallel()

	rules := []string{
		"|i|/images/private/secret.png",
	}

	gpc := NewGlobPathChecker()
	for _, rule := range rules {
		err := gpc.AddRule(rule)
		assert.Nil(t, err)
	}

	f := func(s string, expected bool) {
		t.Helper()
		u, _ := url.Parse(s)
		msg := "should NOT match"
		if expected {
			msg = "should match"
		}
		assert.Equal(
			t,
			gpc.CheckPath(CleanPath(u)),
			expected,
			fmt.Sprintf("%s: %s", msg, s),
		)
	}

	// fmt.Println(gpc.RenderTree())

	// test match
	f("http://example.net/images/private/secret.png", true)
	f("http://example.net/images/private//secret.png", true)
	f("http://example.net/images/private/./secret.png", true)
	f("http://example.net/images/private/secret.png?", true)
	f("http://example.net/images/private%2Fsecret.png", true)
	f("http://example.net/images/private%2fsecret.png", true)

	// test no match
	f("http://example.net/images/private/../public/public.png", false)
}

func TestGlobPathCheckerPathsMisc(t *testing.T) {
	t.Parallel()

	rules := []string{
		"|i|image/*",
		"||video/mp4",
		"||audio/ogg",
		"||pickle/dill+brine",
	}

	gpc := NewGlobPathChecker()
	for _, rule := range rules {
		err := gpc.AddRule(rule)
		assert.Nil(t, err)
	}

	f := func(s string, expected bool) {
		t.Helper()
		msg := "should NOT match"
		if expected {
			msg = "should match"
		}
		assert.Equal(
			t,
			gpc.CheckPath(s),
			expected,
			fmt.Sprintf("%s: %s", msg, s),
		)
	}

	// fmt.Println(gpc.RenderTree())

	// test match
	f("image/png", true)
	f("video/mp4", true)
	f("audio/ogg", true)
	f("pickle/dill+brine", true)

	// test no match
	f("imagex/png", false)
	f("/imagex/png", false)
	f("ximage/png", false)
	f("\nximage/png", false)
	f("ximage/png\n", false)
	f("VIDEO/mp4", false)
	f("xVIDEO/mp4", false)
	f("AUDIO/ogg", false)
	f("xAUDIO/ogg", false)
	f("pickle/dill+briney", false)
	f("pickley/dilly+brine", false)
}

func BenchmarkGlobPathChecker(b *testing.B) {
	rules := []string{
		"|i|*/test.png",
		"||/hodor/test.png",
		"||/hodor/test.png.longer",
		"||/hodor/bar*",
		"||/hodor/ütest.png",
		"||/no*/to/s*/here",
		"||/i/can/s*/it*",
		"||/play/*/ball/img.png",
		"||/yalp*llab/img.png",
	}

	testMatch := []string{
		"http://bar.example.com/foo/TEST.png",
		"http://example.org/foo/test.png",
		"http://example.org/hodor/test.png",
		"http://example.org/hodor/test.png.longer",
		"http://example.org/hodor/bartholemew",
		"http://example.org/hodor/bart/homer.png",
		"http://example.net/nothing/to/see/here",
		"http://example.net/i/can/see/it/in/the/clouds/file.png",
		"http://example.org/play/base/ball/img.png",
		"http://example.org/yalp/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/base/llab/img.png",
		"http://example.org/yalpllab/img.png",
	}

	gpc := NewGlobPathChecker()
	for _, rule := range rules {
		err := gpc.AddRule(rule)
		assert.Nil(b, err)
	}

	preparsedTestMatch := make([]string, 0, len(testMatch))
	for _, u := range testMatch {
		u, _ := url.Parse(u)
		z := u.EscapedPath()
		preparsedTestMatch = append(preparsedTestMatch, z)
	}

	for b.Loop() {
		// avoid inlining optimization
		var x bool
		for _, z := range preparsedTestMatch {
			x = gpc.CheckPath(z)
		}
		_ = x
	}
}
