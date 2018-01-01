// Copyright (c) 2012-2016 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mlog

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatWriterPlainEncodeString(t *testing.T) {
	var stringTests = map[string]struct {
		input  string
		output string
	}{
		"generic":           {`test`, `test`},
		"quote":             {`"this"`, `"this"`},
		"r&n":               {"te\r\nst", `te\r\nst`},
		"tab":               {"\t what", `\t what`},
		"weird chars":       {"\u2028 \u2029", "\u2028 \u2029"},
		"other weird chars": {`"\u003c\u0026\u003e"`, `"\u003c\u0026\u003e"`},
		"invalid utf8":      {"\xff\xff\xffhello", `\ufffd\ufffd\ufffdhello`},
	}

	b := &bytes.Buffer{}
	for name, tt := range stringTests {
		b.Truncate(0)
		encodeStringPlain(b, tt.input)
		assert.Equal(t, []byte(tt.output), b.Bytes(), "%s: did not match expectation", name)
	}
}
