// Copyright (c) 2012-2019 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build ignore

package htrie

import (
	"strings"

	"github.com/xlab/treeprint"
)

func (gpc *GlobPathChecker) printTree(stree treeprint.Tree) {
	for i, x := range gpc.subtrees {
		if x == nil {
			continue
		}
		c := "*"
		if i != 0 {
			c = string(i + charOffset)
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

func (gpc *GlobPathChecker) RenderTree() string {
	tree := treeprint.New()

	meta := make([]string, 0)
	if gpc.isGlob {
		meta = append(meta, "glob")
	}
	if gpc.hasGlobChild {
		meta = append(meta, "glob-child")
	}
	if gpc.canMatch {
		meta = append(meta, "$")
	}
	if len(meta) > 0 {
		tree.SetMetaValue(strings.Join(meta, ","))
	}

	gpc.printTree(tree)
	return tree.String()
}
