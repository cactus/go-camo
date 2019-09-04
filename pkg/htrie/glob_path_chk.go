// Copyright (c) 2012-2019 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package htrie

import (
	"fmt"
	"net/url"
	"strings"
)

const charOffset int = 31

type GlobPathChecker struct {
	//subtree []*GlobPathChecker
	subtrees map[int]*GlobPathChecker
	// used to avoid map lookup when there is only one subtree candidate
	oneShot *GlobPathChecker
	// is this path component a glob
	isGlob bool
	// determines whether a node can be a match even if it isn't a leaf node;
	// this becomes necessary due to the possibility of longer and shorter
	// paths overlapping
	canMatch bool
	// optimization to avoid an extra map lookup on every char
	hasGlobChild bool
	// char for this node
	nodeChar rune
}

func (gpc *GlobPathChecker) parseRule(rule string) (string, string, error) {
	rule = strings.TrimRight(rule, "|")
	if strings.Count(rule, "|") != 2 {
		return "", "", fmt.Errorf("Bad rule format: %s", rule)
	}

	ruleset := make([]strings.Builder, 2)
	index := 0
	// start after first `|`
	for _, r := range rule[1:] {
		if r == '|' {
			index += 1
			continue
		}
		ruleset[index].WriteRune(r)
	}
	parts := make([]string, 2)
	for i, sb := range ruleset {
		parts[i] = strings.TrimSpace(sb.String())
	}
	return parts[0], parts[1], nil
}

func (gpc *GlobPathChecker) AddRule(rule string) error {
	// expected format: |i|/some/subdir/*
	if gpc == nil {
		return fmt.Errorf("got nil <gpc> in receiver")
	}

	if gpc.subtrees == nil {
		return fmt.Errorf("got nil <gpc>.subtrees in receiver")
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

	gpc.addPath(escapedURL, icase)
	return nil
}

func (gpc *GlobPathChecker) addPath(s string, icase bool) error {
	if gpc.subtrees == nil {
		return fmt.Errorf("got nil <gpc>.subtrees in receiver")
	}

	curnode := gpc
	prevnode := curnode
	for _, part := range s {
		r := int(part)

		c := 0
		if part != '*' {
			c = r - charOffset
		}

		d := c
		if c != 0 && icase {
			switch {
			case 'a' <= part && part <= 'z':
				d = c - 32
			case 'A' <= part && part <= 'Z':
				d = c + 32
			default:
				// not a cased letter
				d = c
			}
		}

		subt := curnode.subtrees
		switch {
		case d == c:
			if subt[c] == nil {
				subt[c] = NewGlobPathChecker()
			}
		case subt[c] == nil && subt[d] == nil:
			subt[c] = NewGlobPathChecker()
			subt[d] = subt[c]
		case subt[c] != nil && subt[d] == nil:
			subt[d] = subt[c]
		case subt[c] == nil && subt[d] != nil:
			subt[c] = subt[d]
		}

		subt[c].nodeChar = part
		// setup oneshot as an optimizaiton if there is only one subcandidate...
		// except for cases where the subcandiate is glob!
		if len(subt) == 1 {
			curnode.oneShot = subt[c]
		} else {
			curnode.oneShot = nil
		}

		prevnode = curnode
		curnode = subt[c]
		if part == '*' {
			prevnode.hasGlobChild = true
			curnode.isGlob = true
		}
	}

	// this node can be a match, even if future urls add children
	curnode.canMatch = true
	return nil
}

func (gpc *GlobPathChecker) globConsume(s string, index int) bool {
	// we have a glob and no follow-on chars, so we can consume
	// till the end and then match. early return
	if gpc.canMatch {
		return true
	}

	oneShotLookahead := false
	oneShotStep := true
	// optimize common single char after * globbing
	// eg. .../*/...
	if gpc.oneShot != nil {
		oneShotLookahead = true
		oneShotStep = true
	}

	// otherwise we have some work to do...
	curnode := gpc
	mlen := len(s[index:])
	for i, part := range s[index:] {
		// we know the glob has one one subcandidate (next char), so consume until
		// we hit one of those
		if oneShotStep {
			if part != gpc.oneShot.nodeChar {
				continue
			}
			// got the oneshot expected char finally, so unset oneshot
			// stepping, and proceed
			oneShotStep = false
		}

		r := int(part) - charOffset
		if v, ok := curnode.subtrees[r]; ok {
			// found a candidate. follow it with normal branch logic.
			// if it matches, we're done!
			// increment index value for walkBranch because we consumed a char
			if v.walkBranch(s, i+index+1) {
				return true
			}
		}

		// was this the last char in path?
		if i == mlen-1 {
			// reached the end without a match, and the glob wasn't at the end
			// of the line...
			if !curnode.canMatch {
				return false
			}
			// this should be covered by the test in the start of the function,
			// but add it here in case the code changes in the future.
			return true
		}

		// if we walked the branch, and got no match, we may just need to consume
		// some more... reset oneshot stepping and continue onwards
		if oneShotLookahead {
			oneShotStep = true
		}
	}

	// exhausted the string, but never found a match
	return false
}

func (gpc *GlobPathChecker) walkBranch(s string, index int) bool {
	curnode := gpc
	for i, part := range s[index:] {
		// node has a glob child candidate (consumes), check that first
		if curnode.hasGlobChild {
			if v, ok := curnode.subtrees[0]; ok && v.globConsume(s, index+i) {
				return true
			}
		}

		// oneshot means we only have one child candidate -- an optimization (fastpath)
		// to avoid the slow path map fallback
		if curnode.oneShot != nil {
			// only one candidate, and it _was_ the glob we tried.
			// we're done!
			if curnode.oneShot.nodeChar == '*' {
				return false
			}

			// if oneshot matches, use it
			if curnode.oneShot.nodeChar == part {
				curnode = curnode.oneShot
				continue
			}

			// we had once chance, and it wasn't a glob or a match
			// work is done on this branch
			return false
		}

		// more than one candidate, so fallback to map lookup, since we don't
		// know anything else
		r := int(part) - charOffset
		v, ok := curnode.subtrees[r]
		if !ok {
			return false
		}
		curnode = v
	}

	// reached the end of the string.. check if curnode is a leaf or globby
	// note: unlikely we would end up with a globby here, but maybe possible.
	if curnode.canMatch || curnode.isGlob {
		return true
	}

	// didn't hit a leaf, and didn't find a match
	return false
}

// CheckPath checks the path component of the supplied url.
func (gpc *GlobPathChecker) CheckPath(url *url.URL) bool {
	return gpc.walkBranch(url.EscapedPath(), 0)
}

// CheckPathString checks the supplied path (as a string).
// Note: CheckPathString requires that the url path component is already escaped,
// in a similar way to `(*url.URL).EscapePath()`, as well as TrimSpace'd.
func (gpc *GlobPathChecker) CheckPathString(url string) bool {
	return gpc.walkBranch(url, 0)
}

func NewGlobPathChecker() *GlobPathChecker {
	// refs for valid tree chars
	// https://www.w3.org/TR/2011/WD-html5-20110525/urls.html (refers to RFC 3986)
	// https://en.wikipedia.org/wiki/Uniform_Resource_Identifier#Generic_syntax
	// http://www.asciitable.com
	//
	// omit: less than or equal to 0x0020 or greater than or equal to 0x007F
	// omit: 0x0022, 0x003C, 0x003E, 0x005B, 0x005E, 0x0060, and 0x007B to 0x007D
	// ... so set is:
	//   0x0021             33
	//   0x0023...0x003B    35-59
	//   0x003D             61
	//   0x003F...0x005A    63-90
	//   0x005C             92
	//   0x005D             93
	//   0x005F             94
	//   0x0061...0x007A    97-122
	//   0x007E             126
	// so a total possible of 85 chars, but spread out over 94 slots
	// since there are quite a few slots, let's use a map for now...
	// web searches say a map is faster in go above a certain size. benchmark later...

	// for now, just use the top bounds minus low bound for array size: 126-32=94
	// plus one for our globbing
	return &GlobPathChecker{subtrees: make(map[int]*GlobPathChecker, 95)}
}
