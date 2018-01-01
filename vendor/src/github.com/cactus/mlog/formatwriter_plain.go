package mlog

import (
	"runtime"
	"time"
	"unicode/utf8"
)

// FormatWriterPlain a plain text structured log line.
// Example:
//   2016-04-29T20:49:12Z INFO this is a log
type FormatWriterPlain struct{}

// Emit constructs and formats a plain text log line, then writes it to logger
func (l *FormatWriterPlain) Emit(logger *Logger, level int, message string, extra Map) {
	sb := bufPool.Get()
	defer bufPool.Put(sb)

	flags := logger.Flags()

	// if time is being logged, handle time as soon as possible
	if flags&(Ltimestamp|Ltai64n) != 0 {
		t := time.Now()
		if flags&Ltai64n != 0 {
			writeTimeTAI64N(sb, &t, flags)
		} else {
			writeTime(sb, &t, flags)
		}
		sb.WriteByte(' ')
	}

	if flags&Llevel != 0 {
		switch level {
		case -1:
			sb.WriteString(`DEBUG `)
		case 1:
			sb.WriteString(`FATAL `)
		default:
			sb.WriteString(`INFO  `)
		}
	}

	if flags&(Lshortfile|Llongfile) != 0 {
		_, file, line, ok := runtime.Caller(3)
		if !ok {
			file = "???"
			line = 0
		}

		if flags&Lshortfile != 0 {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
		}

		sb.WriteString(file)
		sb.WriteByte(':')
		sb.AppendIntWidth(line, 0)
		sb.WriteByte(' ')
	}

	encodeStringPlain(sb, message)

	if extra != nil && len(extra) > 0 {
		sb.WriteByte(' ')
		if flags&Lsort != 0 {
			extra.sortedWriteBuf(sb)
		} else {
			extra.unsortedWriteBuf(sb)
		}
	}

	sb.WriteByte('\n')
	sb.WriteTo(logger)
}

// modified from Go stdlib: encoding/json/encode.go:787-862 (approx)
func encodeStringPlain(e byteSliceWriter, s string) {
	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			i++
			if 0x20 <= b {
				e.WriteByte(b)
				continue
			}

			switch b {
			case '\n':
				e.WriteByte('\\')
				e.WriteByte('n')
			case '\r':
				e.WriteByte('\\')
				e.WriteByte('r')
			case '\t':
				e.WriteByte('\\')
				e.WriteByte('t')
			default:
				e.WriteByte(b)
			}
			continue
		}

		c, size := utf8.DecodeRuneInString(s[i:])
		if c == utf8.RuneError && size == 1 {
			e.WriteString(`\ufffd`)
			i++
			continue
		}

		e.WriteString(s[i : i+size])
		i += size
	}
}
