// Copyright (c) 2012-2023 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mlog

import "fmt"

type Attr struct {
	Key   string
	Value interface{}
}

func A(key string, value interface{}) *Attr {
	return &Attr{key, value}
}

func (attr *Attr) writeBuf(w byteSliceWriter) {
	if attr == nil {
		return
	}

	// scratch buffer for intermediate writes
	buf := bufPool.Get()
	defer bufPool.Put(buf)

	w.WriteString(attr.Key)
	w.WriteString(`="`)
	fmt.Fprint(buf, attr.Value)

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

func (attr *Attr) String() string {
	buf := bufPool.Get()
	defer bufPool.Put(buf)
	attr.writeBuf(buf)
	return buf.String()
}

func attrsWriteBuf(w byteSliceWriter, attrs []*Attr) {
	attrsLen := len(attrs)
	for i, attr := range filterAttrs(attrs) {
		attr.writeBuf(w)
		if i != attrsLen-1 {
			w.WriteByte(' ')
		}
	}
}

func filterAttrs(attrs []*Attr) []*Attr {
	hasNil := false
	for _, attr := range attrs {
		if attr == nil {
			hasNil = true
			break
		}
	}
	if !hasNil {
		return attrs
	}

	filteredAttrs := make([]*Attr, 0, len(attrs))
	for _, attr := range attrs {
		if attr != nil {
			filteredAttrs = append(filteredAttrs, attr)
		}
	}
	return filteredAttrs
}
