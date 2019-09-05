// Copyright (c) 2012-2019 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package htrie

import (
	"github.com/xlab/treeprint"
)

func (gpc *GlobPathChecker) RenderTree() string {
	tree := treeprint.New()

	c := tree.AddBranch("case")
	gpc.csNode.printTree(c)
	i := tree.AddBranch("icase")
	gpc.ciNode.printTree(i)
	return tree.String()
}
