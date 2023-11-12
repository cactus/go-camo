// Copyright (c) 2012-2023 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package htrie

import (
	"strings"

	"github.com/xlab/treeprint"
)

func (dt *URLMatcher) printTree(stree treeprint.Tree) {
	meta := make([]string, 0)
	if dt.isWild {
		meta = append(meta, "wild")
	}
	if dt.hasWildChild {
		meta = append(meta, "wild-child")
	}
	if dt.pathChecker != nil {
		meta = append(meta, "has-urls")
	}
	if len(meta) > 0 {
		stree.SetMetaValue(strings.Join(meta, ","))
	}

	for k, v := range dt.subtrees {
		subTree := stree.AddBranch(k)
		v.printTree(subTree)
	}
}

func (dt *URLMatcher) RenderTree() string {
	tree := treeprint.New()
	dt.printTree(tree)
	return tree.String()
}
