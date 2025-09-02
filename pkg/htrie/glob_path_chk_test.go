// Copyright (c) 2012-2023 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package htrie

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/dropwhile/assert"
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
		"http://example.org/yalp/base/llab/img.png",
		"http://example.org/yalpllab/img.png",
	}

	testNoMatch := []string{
		"http://bar.example.com/foo/testx.png",
		"http://example.net/something/to/see/here/file.png",
		"http://example.org/hodor/test.png.long",
	}

	gpc := NewGlobPathChecker()
	for _, rule := range rules {
		err := gpc.AddRule(rule)
		assert.Nil(t, err)
	}

	// fmt.Println(gpc.RenderTree())

	for _, u := range testMatch {
		u, _ := url.Parse(u)
		assert.True(t, gpc.CheckPath(u.EscapedPath()),
			fmt.Sprintf("should have matched: %s", u))

	}
	for _, u := range testNoMatch {
		u, _ := url.Parse(u)
		assert.False(t, gpc.CheckPath(u.EscapedPath()), fmt.Sprintf("should NOT have matched: %s", u))

	}
}

func TestGlobPathCheckerPathsMisc(t *testing.T) {
	t.Parallel()

	rules := []string{
		"|i|image/*",
		"||video/mp4",
		"||audio/ogg",
		"||pickle/dill+brine",
	}

	testMatch := []string{
		"image/png",
		"video/mp4",
		"audio/ogg",
		"pickle/dill+brine",
	}

	testNoMatch := []string{
		"imagex/png",
		"/imagex/png",
		"ximage/png",
		"\nximage/png",
		"ximage/png\n",
		"VIDEO/mp4",
		"xVIDEO/mp4",
		"AUDIO/ogg",
		"xAUDIO/ogg",
		"pickle/dill+briney",
		"pickley/dilly+brine",
	}

	gpc := NewGlobPathChecker()
	for _, rule := range rules {
		err := gpc.AddRule(rule)
		assert.Nil(t, err)
	}

	// fmt.Println(gpc.RenderTree())

	for _, u := range testMatch {
		assert.True(t, gpc.CheckPath(u), fmt.Sprintf("should have matched: %s", u))
	}
	for _, u := range testNoMatch {
		assert.False(t, gpc.CheckPath(u), fmt.Sprintf("should NOT have matched: %s", u))
	}
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

	testIters := 10000

	// avoid inlining optimization
	var x bool
	b.ResetTimer()

	for _, u := range testMatch {
		u, _ := url.Parse(u)
		z := u.EscapedPath()
		for i := 0; i < testIters; i++ {
			x = gpc.CheckPath(z)
		}
	}
	_ = x
}
