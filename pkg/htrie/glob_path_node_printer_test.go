// Copyright (c) 2012-2019 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package htrie

import (
	"fmt"
	"strings"

	"github.com/xlab/treeprint"
)

func (gpn *globPathNode) printTree(stree treeprint.Tree, nodeIndex int) {

	curnode := nodeIndex
	b := &strings.Builder{}
	if gpn == nil || len(gpn.nodeAttrs) < curnode {
		return
	}
	for j := range gpn.nodeTree[curnode] {
		idx := gpn.nodeTree[curnode][j]

		c := gpn.nodeChars[idx]
		if c == globChar {
			c = '*'
		}

		meta := make([]string, 0)
		if gpn.nodeAttrs[idx][0] {
			meta = append(meta, "glob")
		}
		if gpn.nodeAttrs[idx][2] {
			meta = append(meta, "glob-child")
		}
		/*
			if gpn.nodeAttrs[idx][3] {
				meta = append(meta, "1shot")
			}
		*/
		if gpn.nodeAttrs[idx][1] {
			meta = append(meta, "$")
		}

		b.WriteRune(rune(c))

		if len(meta) > 0 {
			fmt.Fprintf(b, " [%s]", strings.Join(meta, ","))
		}
		subTree := stree.AddBranch(b.String())
		b.Reset()

		gpn.printTree(subTree, idx)
	}
}

func (gpn *globPathNode) RenderTree() string {
	tree := treeprint.New()

	meta := make([]string, 0)
	if gpn.nodeAttrs[0][2] {
		meta = append(meta, "glob-child")
	}
	if len(meta) > 0 {
		tree.SetMetaValue(strings.Join(meta, ","))
	}

	if len(gpn.nodeChars) > 1 {
		gpn.printTree(tree, 1)
	}
	return tree.String()
}
