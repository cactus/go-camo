// Copyright (c) 2012-2023 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package htrie

import (
	"fmt"
	"io"
	"strings"
)

func (dt *URLMatcher) printTree(out io.Writer, depth int, prefix string) {
	subTreeCount := len(dt.subtrees)
	iter := 0
	for k, v := range dt.subtrees {
		iter += 1
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
		if v.isWild {
			meta = append(meta, "wild")
		}
		if v.hasWildChild {
			meta = append(meta, "wild-child")
		}
		if v.pathChecker != nil {
			meta = append(meta, "has-urls")
		}

		fmt.Fprintf(out, "%s", k)
		if len(meta) > 0 {
			fmt.Fprintf(out, " [%s]", strings.Join(meta, ","))
		}
		fmt.Fprint(out, "\n")

		v.printTree(out, depth+1, subprefix)
	}
}

func (dt *URLMatcher) RenderTree() string {
	var out strings.Builder
	out.WriteString(".\n")
	dt.printTree(&out, 0, "")
	return out.String()
}
