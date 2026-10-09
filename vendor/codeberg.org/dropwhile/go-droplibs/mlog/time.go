// Copyright (c) 2012-2026 Eli Janssen
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mlog

import (
	"time"

	"codeberg.org/dropwhile/go-droplibs/tai64"
)

func writeTime(sb intSliceWriter, t *time.Time) {
	year, month, day := t.Date()
	sb.AppendIntWidth(year, 4)
	sb.WriteByte('-')
	sb.AppendIntWidth(int(month), 2)
	sb.WriteByte('-')
	sb.AppendIntWidth(day, 2)

	sb.WriteByte('T')

	thour, tmin, tsec := t.Clock()
	sb.AppendIntWidth(thour, 2)
	sb.WriteByte(':')
	sb.AppendIntWidth(tmin, 2)
	sb.WriteByte(':')
	sb.AppendIntWidth(tsec, 2)

	sb.WriteByte('.')
	sb.AppendIntWidth(t.Nanosecond(), 9)

	_, offset := t.Zone()
	if offset == 0 {
		sb.WriteByte('Z')
	} else {
		if offset < 0 {
			sb.WriteByte('-')
			offset = -offset
		} else {
			sb.WriteByte('+')
		}
		offset := offset / 60
		sb.AppendIntWidth(offset/60, 2)
		sb.WriteByte(':')
		sb.AppendIntWidth(offset%60, 2)
	}
}

func writeTimeTAI64N(sb intSliceWriter, t *time.Time) {
	tu := t.UTC()
	tux := tu.Unix()
	offset := tai64.GetOffsetUnix(tux)
	sb.WriteString("@4")
	sb.AppendIntWidthHex(tux+offset, 15)
	sb.AppendIntWidthHex(int64(tu.Nanosecond()), 8)
}
