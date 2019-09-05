// Copyright (c) 2012-2019 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package htrie

import (
	"strings"

	"github.com/xlab/treeprint"
)

func (gpn *globPathNode) printTree(stree treeprint.Tree) {
	for i, x := range gpn.subtrees {
		if x == nil {
			continue
		}
		c := "*"
		if i != 0 {
			c = string(i)
		}

		subTree := stree.AddBranch(c)
		meta := make([]string, 0)
		if x.isGlob {
			meta = append(meta, "glob")
		}
		if x.hasGlobChild {
			meta = append(meta, "glob-child")
		}
		if x.canMatch {
			meta = append(meta, "$")
		}
		if len(meta) > 0 {
			subTree.SetMetaValue(strings.Join(meta, ","))
		}

		x.printTree(subTree)
	}
}

func (gpn *globPathNode) RenderTree() string {
	tree := treeprint.New()

	meta := make([]string, 0)
	if gpn.isGlob {
		meta = append(meta, "glob")
	}
	if gpn.hasGlobChild {
		meta = append(meta, "glob-child")
	}
	if gpn.canMatch {
		meta = append(meta, "$")
	}
	if len(meta) > 0 {
		tree.SetMetaValue(strings.Join(meta, ","))
	}

	gpn.printTree(tree)
	return tree.String()
}
