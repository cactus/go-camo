// Copyright (c) 2012-2023 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package htrie

import (
	"strings"
)

func (gpc *GlobPathChecker) RenderTree() string {
	var out strings.Builder
	out.WriteString("case\n")
	out.WriteString(gpc.csNode.RenderTree())
	out.WriteString("icase\n")
	out.WriteString(gpc.ciNode.RenderTree())
	return out.String()
}
