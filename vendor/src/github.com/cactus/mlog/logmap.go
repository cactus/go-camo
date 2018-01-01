// Copyright (c) 2012-2016 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mlog

import (
	"fmt"
	"sort"
)

// Map is a key value element used to pass
// data to the Logger functions.
type Map map[string]interface{}

// Keys returns an unsorted list of keys in the Map as a []string.
func (m Map) Keys() []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// unsortedWriteBuf writes an unsorted string representation of
// the Map's key value pairs to w.
func (m Map) unsortedWriteBuf(w byteSliceWriter) {
	// scratch buffer for intermediate writes
	buf := bufPool.Get()
	defer bufPool.Put(buf)

	first := true
	for k, v := range m {
		if first {
			first = false
		} else {
			w.WriteByte(' ')
		}

		w.WriteString(k)
		w.WriteString(`="`)

		fmt.Fprint(buf, v)
		// pull out byte slice from buff
		b := buf.Bytes()
		blen := buf.Len()
		p := 0
		for i := 0; i < blen; i++ {
			switch b[i] {
			case '"':
				w.Write(b[p:i])
				w.WriteString(`\"`)
				p = i + 1
			case '\t':
				w.Write(b[p:i])
				w.WriteString(`\t`)
				p = i + 1
			case '\r':
				w.Write(b[p:i])
				w.WriteString(`\r`)
				p = i + 1
			case '\n':
				w.Write(b[p:i])
				w.WriteString(`\n`)
				p = i + 1
			}
		}
		if p < blen {
			w.Write(b[p:blen])
		}

		w.WriteByte('"')
		// truncate intermediate buf so it is clean for next loop
		buf.Truncate(0)
	}
}

// sortedWriteBuf writes a sorted string representation of
// the Map's key value pairs to w.
func (m Map) sortedWriteBuf(w byteSliceWriter) {
	// scratch buffer for intermediate writes
	buf := bufPool.Get()
	defer bufPool.Put(buf)

	keys := m.Keys()
	sort.Strings(keys)

	first := true
	for _, k := range keys {
		if first {
			first = false
		} else {
			w.WriteByte(' ')
		}

		w.WriteString(k)
		w.WriteString(`="`)

		fmt.Fprint(buf, m[k])
		b := buf.Bytes()
		blen := buf.Len()
		p := 0
		for i := 0; i < blen; i++ {
			switch b[i] {
			case '"':
				w.Write(b[p:i])
				w.WriteString(`\"`)
				p = i + 1
			case '\t':
				w.Write(b[p:i])
				w.WriteString(`\t`)
				p = i + 1
			case '\r':
				w.Write(b[p:i])
				w.WriteString(`\r`)
				p = i + 1
			case '\n':
				w.Write(b[p:i])
				w.WriteString(`\n`)
				p = i + 1
			}
		}
		if p < blen {
			w.Write(b[p:blen])
		}

		w.WriteByte('"')
		// truncate intermediate buf so it is clean for next loop
		buf.Truncate(0)
	}
}

// String returns an unsorted string representation of
// the Map's key value pairs.
func (m Map) String() string {
	buf := bufPool.Get()
	defer bufPool.Put(buf)
	m.unsortedWriteBuf(buf)
	return buf.String()
}

// SortedString returns a sorted string representation of
// the Map's key value pairs.
func (m Map) SortedString() string {
	buf := bufPool.Get()
	defer bufPool.Put(buf)
	m.sortedWriteBuf(buf)
	return buf.String()
}
