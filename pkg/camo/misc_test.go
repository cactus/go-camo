// Copyright (c) 2012-2023 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package camo

import (
	"testing"

	"github.com/cactus/go-camo/v2/pkg/assert"
)

func Test_getMaxRangeByte(t *testing.T) {
	f := func(input string, expected int64, expectedErr any) {
		t.Helper()

		got, err := getMaxRangeByte(input)
		assert.Error(t, err, expectedErr)
		assert.Equal(t, got, expected)
	}

	f("bytesx1,2,4,100-200", -1, "improper prefix")
	f("bytes", -1, "improper format")
	f("bytes=", -1, "no values after prefix")
	f("bytes=100-200,-1024", -1, "empty value before '-'")

	f("bytes=1,2,4,100-200", 200, nil)
	f("bytes= 1, 2, 4, 100 - 200", 200, nil)
	f("bytes = 1, 2, 4, 100 - 200", 200, nil)
	f("bytes =1,2,4, 100 -200,300", 300, nil)
}
