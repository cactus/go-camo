// Copyright (c) 2012-2023 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package camo

import (
	"strings"
)

func containsOneOf(s string, substrs ...string) bool {
	for i := range substrs {
		if strings.Contains(s, substrs[i]) {
			return true
		}
	}
	return false
}
