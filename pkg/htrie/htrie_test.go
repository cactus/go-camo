// Copyright (c) 2012-2023 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package htrie

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/dropwhile/assert"
)

func TestHTrieCheckURL(t *testing.T) {
	t.Parallel()

	rules := []string{
		"|s|example.org|i|*/test.png",
		"||*.example.com||*/test.png",
		"||example.net||*",
		"||foo.example.net||/",
		"||bar.example.net|i|*/test.png",
		"||bar.example.net|i|*/test.png.extra",
		"||bücher.example.net||",
	}

	testMatch := []string{
		"http://example.org/foo/test.png",
		"http://example.org/foo/TEST.png",
		"http://bar.example.com/foo/test.png",
		"http://example.net/test.png",
		"http://foo.example.net/",
		"http://bar.example.net/foo/test.png",
		"http://bar.example.net/foo/test.png.extra",
		"http://bücher.example.net/",
		"http://xn--bcher-kva.example.net/",
	}

	testNoMatch := []string{
		"http://example.com/foo/test.png",
		"http://example.org/foo/testx.png",
		"http://foo.example.net/nope",
		"http://bar.example.org/foo/testx.png",
		"http://bar.example.net/foo/test.png.ex",
		"http://bücher.example.com/",
	}

	dt := NewURLMatcher()
	for _, rule := range rules {
		err := dt.AddRule(rule)
		assert.Nil(t, err)
	}

	// fmt.Println(dt.RenderTree())

	for _, u := range testMatch {
		u, _ := url.Parse(u)
		chk, err := dt.CheckURL(u)
		assert.Nil(t, err)
		assert.True(t, chk, fmt.Sprintf("should have matched: %s", urlPathUnescape(u)))
	}
	for _, u := range testNoMatch {
		u, _ := url.Parse(u)
		chk, err := dt.CheckURL(u)
		assert.Nil(t, err)
		assert.False(t, chk, fmt.Sprintf("should not have matched: %s", urlPathUnescape(u)))
	}
}

func urlPathUnescape(u *url.URL) string {
	s := u.String()
	p, _ := url.PathUnescape(s)
	return p
}

func TestHTrieCheckHostname(t *testing.T) {
	t.Parallel()

	rules := []string{
		"|s|localhost||",
		"|s|localdomain||",
		"||bücher.example.net||",
	}

	testMatch := []string{
		"http://localhost/foo/test.png",
		"http://foo.localhost/foo/test.png",
		"http://bar.foo.localhost/foo/test.png",
		"http://localdomain/foo/TEST.png",
		"http://foo.localdomain/foo/test.png",
		"http://bar.foo.localdomain/foo/test.png",
		"http://bücher.example.net/",
		"http://xn--bcher-kva.example.net/",
	}

	testNoMatch := []string{
		"http://example.com/foo/test.png",
		"http://example.org/foo/testx.png",
		"http://foo.example.net/nope",
		"http://bar.example.org/foo/testx.png",
		"http://bar.example.net/foo/test.png.ex",
		"http://bücher.example.com/",
	}

	dt := NewURLMatcher()
	for _, rule := range rules {
		err := dt.AddRule(rule)
		if err != nil {
			t.Errorf("failed to add domain rule: %s", err)
		}
	}

	// fmt.Println(dt.RenderTree())

	for _, u := range testMatch {
		u, _ := url.Parse(u)
		result, err := dt.CheckHostname(u.Hostname())
		assert.Nil(t, err)
		assert.True(t, result, fmt.Sprintf("should have matched: %s", urlPathUnescape(u)))
	}
	for _, u := range testNoMatch {
		u, _ := url.Parse(u)
		result, err := dt.CheckHostname(u.Hostname())
		assert.Nil(t, err)
		assert.False(t, result, fmt.Sprintf("should not have matched: %s", urlPathUnescape(u)))
	}
}

func BenchmarkHTrieCreate(b *testing.B) {
	dt := NewURLMatcher()
	urls := []string{
		"||*.example.com||*/test.png",
		"|s|example.org|i|*/test.png",
		"||foo.example.net||/test.png",
		"||bar.example.net||/test.png",
		"||*.bar.example.net||/test.png",
		"||*.hodor.example.net||/*/test.png",
	}
	var err error

	for b.Loop() {
		for _, u := range urls {
			err = dt.AddRule(u)
			if err != nil {
				b.Errorf("%s", err)
			}
		}
	}
	_ = err
}

func BenchmarkRegexCreate(b *testing.B) {
	urls := []string{
		`^.*\.example.com/.*/test.png`,
		`^(.*\.)?example.org/(?:i.*)/test.png`,
	}

	var r *regexp.Regexp
	var err error

	for b.Loop() {
		for _, u := range urls {
			r, err = regexp.Compile(u)
			if err != nil {
				b.Errorf("%s", err)
			}
		}
	}
	_ = r
	_ = err
}

var urlMatchTestURLs = []string{
	"http://example.com/foo/test.png",
	"http://bar.example.com/foo/test.png",
	"http://bar.example.com/foo/testx.png",
	"http://bar.example.com/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/test.png",
	"http://bar.example.com/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/testx.png",
	// this one kills the regex pretty bad. :(
	"http://bar.example.com/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/testx.png",
}

func BenchmarkHTrieMatch(b *testing.B) {
	rules := []string{
		"||foo.example.net||/test.png",
		"||bar.example.net||/test.png",
		"||*.bar.example.net||/test.png",
		"||*.hodor.example.net||/*/test.png",
		"||*.example.com||*/test.png",
		"|s|example.org|i|*/test.png",
	}

	testIters := 10000

	dt := NewURLMatcher()
	for _, rule := range rules {
		err := dt.AddRule(rule)
		assert.Nil(b, err)
	}

	parsed := make([]*url.URL, 0)
	for _, u := range urlMatchTestURLs {
		u, _ := url.Parse(u)
		parsed = append(parsed, u)
	}

	// avoid inlining optimization
	var x bool
	b.ResetTimer()

	for _, u := range parsed {
		for i := 0; i < testIters; i++ {
			x, _ = dt.CheckURL(u)
		}
	}
	_ = x
}

func BenchmarkRegexMatch(b *testing.B) {
	rules := []string{
		// giving regex lots of help here, putting this rule first
		`^.*\.example.com/.*/test.png`,
		`^bar.example.net/test.png`,
		`^foo.example.net/test.png`,
		`^.*\.bar.example.net/test.png`,
		`^.*\.hodor.example.net/.*/test.png`,
		`^(.*\.)?example.org/(?:i.*/test.png)`,
	}

	testIters := 10000

	rexes := make([]*regexp.Regexp, 0)
	for _, r := range rules {
		rx, _ := regexp.Compile(r)
		rexes = append(rexes, rx)
	}

	// strip protocol prefix to make regex matches easier
	testUrls := make([]string, len(urlMatchTestURLs))
	for _, u := range urlMatchTestURLs {
		testUrls = append(testUrls, strings.TrimPrefix(u, "http://"))
	}

	// avoid inlining optimization
	var x bool
	b.ResetTimer()

	for _, u := range testUrls {
		for i := 0; i < testIters; i++ {
			// walk regexes in order. first match wins
			for _, rx := range rexes {
				if rx.MatchString(u) {
					x = true
					break
				}
			}
		}
	}
	_ = x
}

func BenchmarkHTrieMatchHostname(b *testing.B) {
	rules := []string{
		"||foo.example.net||/test.png",
		"||bar.example.net||/test.png",
		"||*.bar.example.net||/test.png",
		"||*.hodor.example.net||/*/test.png",
		"||*.example.com||*/test.png",
		"|s|example.org|i|*/test.png",
	}

	testIters := 10000

	dt := NewURLMatcher()
	for _, rule := range rules {
		err := dt.AddRule(rule)
		assert.Nil(b, err)
	}

	parsed := make([]string, 0)
	for _, u := range urlMatchTestURLs {
		u, _ := url.Parse(u)
		parsed = append(parsed, u.Hostname())
	}

	// avoid inlining optimization
	var x bool
	b.ResetTimer()

	b.Run("CheckHostname", func(b *testing.B) {
		for _, u := range parsed {
			for i := 0; i < testIters; i++ {
				x, _ = dt.CheckHostname(u)
			}
		}
	})

	b.Run("CheckCleanHostname", func(b *testing.B) {
		for _, u := range parsed {
			for i := 0; i < testIters; i++ {
				x = dt.CheckCleanHostname(u)
			}
		}
	})

	_ = x
}
