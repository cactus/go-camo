// Copyright (c) 2015-2026 Eli Janssen
// Use of this source code is governed by an ISC-style
// license that can be found in the LICENSE file.

package strx

import "strings"

// ContainsOneOf reports whether one (or more) of any of the supplied substrs is within s.
func ContainsOneOf(s string, substrs []string) bool {
	for i := range substrs {
		if strings.Contains(s, substrs[i]) {
			return true
		}
	}
	return false
}
