// Copyright (c) 2012-2019 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package htrie

import (
	"fmt"
	"net/url"
	"strings"
)

// A GlobPathChecker represents a path checker that supports globbing comparisons
type GlobPathChecker struct {
	// case sensitive checker
	csNode *globPathNode
	// case insensitive checker
	ciNode *globPathNode
}

func (gpc *GlobPathChecker) parseRule(rule string) (string, string, error) {
	rule = strings.TrimRight(rule, "|")
	if strings.Count(rule, "|") != 2 {
		return "", "", fmt.Errorf("bad rule format: %s", rule)
	}

	ruleset := make([]strings.Builder, 2)
	index := 0
	// start after first `|`
	for _, r := range rule[1:] {
		if r == '|' {
			index++
			continue
		}
		_, err := ruleset[index].WriteRune(r)
		if err != nil {
			return "", "", err
		}
	}
	parts := make([]string, 2)
	for i, sb := range ruleset {
		parts[i] = strings.TrimSpace(sb.String())
	}
	return parts[0], parts[1], nil
}

// AddRule adds a rule to the GlobPathChecker.
// The expected rule format is:
//
//	<pipe-character><flags><pipe-character><match-url>
//
// Example:
//
//	|i|/some/subdir/*
//
// Allowed flags:
//
// * `i`: URL match string should be matched case insensitively
func (gpc *GlobPathChecker) AddRule(rule string) error {
	// expected format: |i|/some/subdir/*
	if gpc == nil {
		return fmt.Errorf("got nil <gpc> in receiver")
	}

	urlRuleFlags, urlRuleMatch, err := gpc.parseRule(rule)
	if err != nil {
		return err
	}

	icase := false
	if strings.Contains(urlRuleFlags, "i") {
		icase = true
	}

	if strings.ContainsAny(urlRuleMatch, "?#") {
		return fmt.Errorf("bad url: contains query string components")
	}

	// pipe encodes to %7C, and since pipe isn't valid unencoded in the rule
	// file (used as a separator), use it as a standin for `*` in the url so
	// we can rely on go's url parsing to get the escaped path for matching.
	// afterwards, we replace %7C back to `*`.
	// also handle the case where a url /does/ happen to include %7C already.
	hasGlob := false
	usePipe := false
	if strings.Contains(urlRuleMatch, "*") {
		hasGlob = true
		if strings.Contains(urlRuleMatch, "%7C") {
			usePipe = true
			urlRuleMatch = strings.ReplaceAll(urlRuleMatch, "*", "|")
		}
	}

	u, err := url.Parse(urlRuleMatch)
	if err != nil {
		return fmt.Errorf("bad url: %s", err)
	}

	escapedURL := u.EscapedPath()
	// note: `*` may or may not have been escaped, dependig on go url parsing
	// internals. for example, if the following evals to true:
	//    validEncodedPath(u.RawPath) and unescape(u.RawPath) == u.Path
	// ref: https://golang.org/src/net/url/url.go?s=20096:20130#704
	if hasGlob {
		if usePipe {
			// our url definitions can't contain `|`, since that is the separator,
			// so use that as a replace char..
			// note that it _should_ be escaped, but try to replace both anyway
			// for safety/future proofing...
			escapedURL = strings.ReplaceAll(escapedURL, "|", "*")
		} else {
			escapedURL = strings.ReplaceAll(escapedURL, "%7C", "*")
		}
	}

	if icase {
		if gpc.ciNode == nil {
			gpc.ciNode = newGlobPathNode(true)
		}
		err = gpc.ciNode.addPath(escapedURL)
	} else {
		if gpc.csNode == nil {
			gpc.csNode = newGlobPathNode(false)
		}
		err = gpc.csNode.addPath(escapedURL)
	}

	if err != nil {
		return err
	}
	return nil
}

// CheckPath checks the supplied path (as a string).
// Note: CheckPathString requires that the url path component is already escaped,
// in a similar way to `(*url.URL).EscapePath()`.
func (gpc *GlobPathChecker) CheckPath(url string) bool {
	url = strings.TrimSpace(url)
	ulen := len(url)

	// if we have a case sensitive checker, check that one first
	if gpc.csNode != nil && gpc.csNode.checkPath(url, 0, ulen, 0) {
		return true
	}

	// if we have a case insensitive checker, check that one next
	if gpc.ciNode != nil && gpc.ciNode.checkPath(url, 0, ulen, 0) {
		return true
	}

	// no matches, return false
	return false
}

// NewGlobPathChecker returns a new GlobPathChecker
func NewGlobPathChecker() *GlobPathChecker {
	return &GlobPathChecker{}
}
