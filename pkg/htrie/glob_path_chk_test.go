// Copyright (c) 2012-2019 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package htrie

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobPathChecker(t *testing.T) {
	t.Parallel()

	rules := []string{
		"|i|*/test.png",
		"||/hodor/test.png",
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
		"http://example.org/hodor/bartholemew",
		"http://example.org/hodor/bart/homer.png",
		"http://example.net/nothing/to/see/here",
		"http://example.net/i/can/see/it/in/the/clouds/file.png",
		"http://example.org/play/base/ball/img.png",
		"http://example.org/yalp/base/llab/img.png",
	}

	testNoMatch := []string{
		"http://bar.example.com/foo/testx.png",
		"http://example.net/something/to/see/here/file.png",
	}

	gpc := NewGlobPathChecker()
	for _, rule := range rules {
		err := gpc.AddRule(rule)
		assert.Nil(t, err)
	}

	//fmt.Println(gpc.RenderTree())

	for _, u := range testMatch {
		u, _ := url.Parse(u)
		assert.True(t, gpc.CheckPath(u.EscapedPath()),
			fmt.Sprintf("should have matched: %s", u),
		)
	}
	for _, u := range testNoMatch {
		u, _ := url.Parse(u)
		assert.False(t, gpc.CheckPath(u.EscapedPath()),
			fmt.Sprintf("should NOT have matched: %s", u),
		)
	}
}

func TestGlobPathCheckerPathsMisc(t *testing.T) {
	t.Parallel()

	rules := []string{
		"|i|image/*",
		"||video/mp4",
		"||pickle/dill+brine",
	}

	testMatch := []string{
		"image/png",
		"video/mp4",
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
		"pickle/dill+briney",
		"pickley/dilly+brine",
	}

	gpc := NewGlobPathChecker()
	for _, rule := range rules {
		err := gpc.AddRule(rule)
		assert.Nil(t, err)
	}

	//fmt.Println(gpc.RenderTree())

	for _, u := range testMatch {
		assert.True(t, gpc.CheckPath(u), fmt.Sprintf("should have matched: %s", u))
	}
	for _, u := range testNoMatch {
		assert.False(t, gpc.CheckPath(u), fmt.Sprintf("should NOT have matched: %s", u))
	}
}
