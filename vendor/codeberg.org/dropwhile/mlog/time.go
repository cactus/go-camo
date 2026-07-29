package mlog

import (
	"time"

	"codeberg.org/dropwhile/tai64"
)

func writeTime(sb intSliceWriter, t *time.Time) {
	year, month, day := t.Date()
	sb.AppendIntWidth(year, 4)
	sb.WriteByte('-')
	sb.AppendIntWidth(int(month), 2)
	sb.WriteByte('-')
	sb.AppendIntWidth(day, 2)

	sb.WriteByte('T')

	hour, min, sec := t.Clock()
	sb.AppendIntWidth(hour, 2)
	sb.WriteByte(':')
	sb.AppendIntWidth(min, 2)
	sb.WriteByte(':')
	sb.AppendIntWidth(sec, 2)

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
