// Copyright (c) 2012-2019 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package htrie

import (
	"fmt"
)

const charOffset int = 31

// A globPathNode represents a path checker that supports globbing comparisons
type globPathNode struct {
	subtrees map[int]*globPathNode
	// used to avoid map lookup when there is only one subtree candidate
	oneShot *globPathNode
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

func (gpn *globPathNode) addPath(s string) error {
	if gpn.subtrees == nil {
		return fmt.Errorf("got nil <gpn>.subtrees in receiver")
	}

	curnode := gpn
	prevnode := curnode
	for _, part := range s {
		r := int(part)

		var c int
		if part == '*' {
			c = 0
		} else {
			c = r - charOffset
		}

		subt := curnode.subtrees
		if subt[c] == nil {
			subt[c] = newGlobPathNode()
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

func (gpn *globPathNode) globConsume(s string, index int) bool {
	// we have a glob and no follow-on chars, so we can consume
	// till the end and then match. early return
	if gpn.canMatch {
		return true
	}

	oneShotLookahead := false
	oneShotStep := true
	// optimize common single char after * globbing
	// eg. .../*/...
	if gpn.oneShot != nil {
		oneShotLookahead = true
		oneShotStep = true
	}

	// otherwise we have some work to do...
	curnode := gpn
	mlen := len(s[index:])
	for i, part := range s[index:] {
		// we know the glob has one one subcandidate (next char), so consume until
		// we hit one of those
		if oneShotStep {
			if part != gpn.oneShot.nodeChar {
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

func (gpn *globPathNode) walkBranch(s string, index int) bool {
	curnode := gpn
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

func newGlobPathNode() *globPathNode {
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

	// NOTE: since realloc cost is paid at creation, and we want to reduce size
	// and we only care about lookup costs, just start with 0 and let it grow
	// as needed.
	// return &globPathNode{subtrees: make(map[int]*globPathNode, 95)}
	return &globPathNode{subtrees: make(map[int]*globPathNode, 0)}
}
