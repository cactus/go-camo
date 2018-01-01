package mlog

import (
	"time"

	"github.com/cactus/tai64"
)

func writeTime(sb intSliceWriter, t *time.Time, flags FlagSet) {
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
		sb.AppendIntWidth(offset/3600, 2)
		sb.WriteByte(':')
		sb.AppendIntWidth(offset%3600, 2)
	}
}

func writeTimeTAI64N(sb intSliceWriter, t *time.Time, flags FlagSet) {
	tu := t.UTC()
	tux := tu.Unix()
	offset := tai64.GetOffsetUnix(tux)
	sb.WriteString("@4")
	sb.AppendIntWidthHex(tux+offset, 15)
	sb.AppendIntWidthHex(int64(tu.Nanosecond()), 8)
}
