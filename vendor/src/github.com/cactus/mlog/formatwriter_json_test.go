// Copyright (c) 2012-2016 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mlog

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testEncodeString(e byteSliceWriter, s string) {
	encoder := json.NewEncoder(e)
	encoder.Encode(s)
}

func TestFormatWriterJSONEncodeString(t *testing.T) {
	var jsonStringTests = map[string]string{
		"generic":           `test`,
		"quote":             `"this"`,
		"r&n":               "te\r\nst",
		"tab":               "\t what",
		"weird chars":       "\u2028 \u2029",
		"other weird chars": `"\u003c\u0026\u003e"`,
		"invalid utf8":      "\xff\xff\xffhello",
	}

	b := &bytes.Buffer{}
	for name, s := range jsonStringTests {
		e, err := json.Marshal(s)
		assert.Nil(t, err, "%s: json marshal failed", name)

		b.Truncate(0)
		b.WriteByte('"')
		encodeStringJSON(b, s)
		b.WriteByte('"')
		assert.Equal(t, string(e), b.String(), "%s: did not match expectation", name)
	}
}
