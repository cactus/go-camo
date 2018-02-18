package mlog

import (
	"runtime"
	"time"
	"unicode/utf8"
)

// FormatWriterStructured writes a plain text structured log line.
// Example:
//   time="2016-04-29T20:49:12Z" level="I" msg="this is a log"
type FormatWriterStructured struct{}

// Emit constructs and formats a plain text log line, then writes it to logger
func (l *FormatWriterStructured) Emit(logger *Logger, level int, message string, extra Map) {
	sb := bufPool.Get()
	defer bufPool.Put(sb)

	flags := logger.Flags()

	// if time is being logged, handle time as soon as possible
	if flags&(Ltimestamp|Ltai64n) != 0 {
		t := time.Now()
		sb.WriteString(`time="`)
		if flags&Ltai64n != 0 {
			writeTimeTAI64N(sb, &t, flags)
		} else {
			writeTime(sb, &t, flags)
		}
		sb.WriteString(`" `)
	}

	if flags&Llevel != 0 {
		sb.WriteString(`level="`)
		switch level {
		case -1:
			sb.WriteByte('D')
		case 1:
			sb.WriteByte('F')
		default:
			sb.WriteByte('I')
		}
		sb.WriteString(`" `)
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

		sb.WriteString(`caller="`)
		sb.WriteString(file)
		sb.WriteByte(':')
		sb.AppendIntWidth(line, 0)
		sb.WriteString(`" `)
	}

	sb.WriteString(`msg="`)
	encodeStringStructured(sb, message)
	sb.WriteByte('"')

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
func encodeStringStructured(e byteSliceWriter, s string) {
	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			i++
			if 0x20 <= b && b != '"' {
				e.WriteByte(b)
				continue
			}

			switch b {
			case '"':
				e.WriteByte('\\')
				e.WriteByte(b)
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
