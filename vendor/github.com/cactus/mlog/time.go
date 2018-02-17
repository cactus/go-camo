package mlog

import "time"

func writeTime(sb sliceWriter, t *time.Time, flags FlagSet) {
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

	switch {
	case flags&Lnanoseconds != 0:
		sb.WriteByte('.')
		sb.AppendIntWidth(t.Nanosecond(), 9)
	case flags&Lmicroseconds != 0:
		sb.WriteByte('.')
		sb.AppendIntWidth(t.Nanosecond()/1e3, 6)
	}

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
