// Copyright (c) 2012-2023 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package htrie

import (
	"fmt"
)

const globChar uint32 = 1

// A globPathNode represents a path checker that supports globbing comparisons
type globPathNode struct {
	// go maps are optimized for only certain int types:
	//  -- results as of go 1.13 on my slow laptop --
	//  BenchmarkInt        297391227    3.99 ns/op
	//  BenchmarkInt8        68107761   17.90 ns/op
	//  BenchmarkInt16       65628482   18.30 ns/op
	//  BenchmarkInt32      292725417    4.08 ns/op
	//  BenchmarkInt64      293602374    4.11 ns/op
	//  BenchmarkUInt       298711089    3.99 ns/op
	//  BenchmarkUInt8       68173198   17.80 ns/op
	//  BenchmarkUInt16      67566312   18.10 ns/op
	//  BenchmarkUInt32     298597942    3.99 ns/op
	//  BenchmarkUInt64     300239860    4.02 ns/op
	//
	// Since we would /want/ to use uint8 here, use uint32 instead
	// Ugly and wasteful, but quite a bit faster for now...
	//
	// Further, using a map of uint32 is also slightly faster than using
	// a [][]uint list to hold indexes into another []uint8 list
	// eg. SOA (Struct of Array) vs this AOS (Array of Struct) method.
	// Go maps are pretty efficient, and these structs are pointers,
	// which reduces some overhead.
	// Note that as far as memory usage goes, SOA does come out ahead,
	// but tree setup is a one-time startup cost, and we are willing to
	// trade off some additional memory and/or slowness there
	// for faster rules processing speed.

	// subtree of nodes
	subtrees map[uint32]*globPathNode
	// used to avoid map lookup when there is only one subtree candidate
	oneShot *globPathNode
	// char for this node
	nodeChar uint32

	// note: Using a bitmask instead of separate boolean slots, uses
	// less memory, but we again make the tradeoff of slight increase
	// in memory for slightly faster rules processing speed.
	// A boolean check is faster than a bitcheck+equality check

	// whether this node is a glob node
	isGlob bool
	// determines whether a node can be a match even if it isn't a leaf node;
	// this becomes necessary due to the possibility of longer and shorter
	// paths overlapping
	canMatch bool
	// whether the node has a wildcard/glob descendent
	// this is an optimization to avoid an extra map lookup on every char
	hasGlobChild bool
	// is this a case insensitive comparison tree?
	icase bool
}

func (gpn *globPathNode) addPath(s string) error {
	if gpn.subtrees == nil {
		return fmt.Errorf("got nil <gpn>.subtrees in receiver")
	}

	curnode := gpn
	prevnode := curnode
	mlen := len(s)
	for i := range mlen {
		part := uint32(s[i])

		// if icase, use lowercase letters for comparisons
		// 'A' == 65; 'Z' == 90
		if gpn.icase && 65 <= part && part <= 90 {
			part = part + 32
		}

		var c uint32
		// '*' == 42
		if part == 42 {
			c = globChar
		} else {
			c = part
		}

		subt := curnode.subtrees
		if subt[c] == nil {
			subt[c] = newGlobPathNode(gpn.icase)
		}

		subt[c].nodeChar = part

		// setup oneshot as an optimizaiton if there is only one subcandidate...
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

	// this is the end of the path, so this node can be a match, even if future
	// urls add children to it (a longer url).
	curnode.canMatch = true
	return nil
}

func (gpn *globPathNode) globConsume(s string, index, mlen int) bool {
	// we have a glob and no follow-on chars, so we can consume
	// till the end and then match. early return
	if gpn.canMatch {
		return true
	}

	oneShotLookahead := false
	oneShotStep := false
	// optimize common single char after * globbing
	// eg. .../*/...
	if gpn.oneShot != nil {
		oneShotLookahead = true
		oneShotStep = true
	}

	// otherwise we have some work to do...
	curnode := gpn
	// don't need to iter runes since we have ascii
	for i := index; i < mlen; i++ {
		part := uint32(s[i])

		// if icase, use lowercase letters for comparisons
		// 'A' == 65; 'Z' == 90
		if gpn.icase && 65 <= part && part <= 90 {
			part = part + 32
		}

		// we know the glob has one one subcandidate (next char), so consume until
		// we hit one of those
		if oneShotStep {
			if part != curnode.oneShot.nodeChar {
				continue
			}
			// got the oneshot expected char finally, so unset oneshot
			// stepping, and proceed
			oneShotStep = false
		}

		if v, ok := curnode.subtrees[part]; ok {
			// found a candidate. follow it with normal branch logic.
			// if it matches, we're done!
			// increment index value for checkPath because we consumed a char
			// by following oneShot
			if v.checkPath(s, i+1, mlen) {
				return true
			}
		}

		// was this the last char in path?
		if i == mlen-1 {
			// reached the end without a match, and the glob wasn't at the end
			// of the line... return whether the curnode can match or not,
			// to determine overall success or failure
			return curnode.canMatch
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

func (gpn *globPathNode) checkPath(s string, index, mlen int) bool {
	curnode := gpn
	// don't need to iter runes since we have ascii
	for i := index; i < mlen; i++ {
		part := uint32(s[i])

		// if icase, use lowercase letters for comparisons
		// 'A' == 65; 'Z' == 90
		if gpn.icase && 65 <= part && part <= 90 {
			part = part + 32
		}

		// node may have a glob child candidate (consumes), check that first
		if curnode.hasGlobChild {
			// get glob node, and call globconsume.
			// don't advance string pointer here though, as a glob is a match
			// node and not a standard char node (it can also match zero characters)
			if v, ok := curnode.subtrees[globChar]; ok && v.globConsume(s, i, mlen) {
				return true
			}
		}

		// oneshot means we only have one child candidate -- an optimization (fastpath)
		// to avoid the slow path map fallback
		if curnode.oneShot != nil {
			// only one candidate, and it _was_ the glob we tried.
			// we're done!
			if curnode.oneShot.nodeChar == globChar {
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
		v, ok := curnode.subtrees[part]
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

func newGlobPathNode(icase bool) *globPathNode {
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
	// since there are quite a few possible slots, let's use a map for now...
	// web searches say a map is faster in go above a certain size. benchmark later...
	//
	// for now, since realloc cost is paid at creation, and we want to RSS size
	// and since we only /really/ care about lookup costs, just start with 0 initial
	// map size and let it grow as needed
	return &globPathNode{
		subtrees: make(map[uint32]*globPathNode),
		icase:    icase,
	}
}
