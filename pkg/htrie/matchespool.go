// Copyright (c) 2012-2023 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package htrie

import (
	"sync"
)

var matchesPool = sync.Pool{
	New: func() any {
		// starting backing array size of 8
		// that /seems/ like a pretty good initial value, without
		// being too crazy, and has the nice property of being a powler of 2. ;)
		matches := make([]*URLMatcher, 0, 8)
		return &matches
	},
}

func getMatchesPoolElem() *[]*URLMatcher {
	return matchesPool.Get().(*[]*URLMatcher)
}

func putMatchesPoolElem(s *[]*URLMatcher) {
	*s = (*s)[0:0]
	matchesPool.Put(s)
}
