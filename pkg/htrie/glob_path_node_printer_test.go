// Copyright (c) 2012-2023 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package htrie

import (
	"fmt"
	"io"
	"strings"
)

func (gpn *globPathNode) printTree(out io.Writer, depth int, prefix string) {
	subTreeCount := len(gpn.subtrees)
	iter := 0
	for i, x := range gpn.subtrees {
		if x == nil {
			continue
		}
		iter += 1

		c := "*"
		if i != 0 && i != 1 {
			// we use uint32 for performance, and don't care about
			// truncation at all here (just printing anyway), so
			// just convert.
			c = string(uint8(i))
		}

		subprefix := prefix
		if subTreeCount > 1 {
			if iter < subTreeCount {
				subprefix += "├── "
			} else {
				subprefix += "└── "
			}
		} else {
			subprefix += "└── "
		}
		fmt.Fprint(out, subprefix)
		if before, ok := strings.CutSuffix(subprefix, "├── "); ok {
			subprefix = before + "│   "
		}
		if before, ok := strings.CutSuffix(subprefix, "└── "); ok {
			subprefix = before + "    "
		}

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

		fmt.Fprintf(out, "%s", c)

		if len(meta) > 0 {
			fmt.Fprintf(out, " [%s]", strings.Join(meta, ","))
		}
		fmt.Fprint(out, "\n")

		x.printTree(out, depth+1, subprefix)
	}
}

func (gpn *globPathNode) RenderTree() string {
	var out strings.Builder
	gpn.printTree(&out, 0, "")
	return out.String()
}
