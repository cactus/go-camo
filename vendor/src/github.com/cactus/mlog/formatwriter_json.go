package mlog

import (
	"fmt"
	"runtime"
	"time"
	"unicode/utf8"
)

const hex = "0123456789abcdef"

// FormatWriterJSON writes a json structured log line.
// Example:
//   {"time": "2016-04-29T20:49:12Z", "level": "I", "msg": "this is a log"}
type FormatWriterJSON struct{}

// Emit constructs and formats a json log line, then writes it to logger
func (j *FormatWriterJSON) Emit(logger *Logger, level int, message string, extra Map) {
	sb := bufPool.Get()
	defer bufPool.Put(sb)

	flags := logger.Flags()

	sb.WriteByte('{')
	// if time is being logged, handle time as soon as possible
	if flags&(Ltimestamp|Ltai64n) != 0 {
		t := time.Now()
		sb.WriteString(`"time": "`)
		if flags&Ltai64n != 0 {
			writeTimeTAI64N(sb, &t, flags)
		} else {
			writeTime(sb, &t, flags)
		}
		sb.WriteString(`", `)
	}

	if flags&Llevel != 0 {
		sb.WriteString(`"level": "`)
		switch level {
		case -1:
			sb.WriteByte('D')
		case 1:
			sb.WriteByte('F')
		default:
			sb.WriteByte('I')
		}
		sb.WriteString(`", `)
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

		sb.WriteString(`"caller": "`)
		sb.WriteString(file)
		sb.WriteByte(':')
		sb.AppendIntWidth(line, 0)
		sb.WriteString(`", `)
	}

	sb.WriteString(`"msg": "`)
	encodeStringJSON(sb, message)
	sb.WriteByte('"')

	if extra != nil && len(extra) > 0 {
		sb.WriteString(`, "extra": {`)
		encodeLogMapJSON(sb, extra)
		sb.WriteByte('}')
	}

	sb.WriteByte('}')
	sb.WriteByte('\n')
	sb.WriteTo(logger)
}

func encodeLogMapJSON(w byteSliceWriter, m Map) {
	first := true
	for k, v := range m {
		if first {
			first = false
		} else {
			w.WriteString(`, `)
		}

		w.WriteByte('"')
		encodeStringJSON(w, k)
		w.WriteString(`": "`)
		encodeStringJSON(w, fmt.Sprint(v))
		w.WriteByte('"')
	}
}

// modified from Go stdlib: encoding/json/encode.go:787-862 (approx)
func encodeStringJSON(e byteSliceWriter, s string) {
	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			i++
			if 0x20 <= b && b != '\\' && b != '"' {
				e.WriteByte(b)
				continue
			}

			switch b {
			case '\\', '"':
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
				// This encodes bytes < 0x20 except for escapes above
				e.WriteString(`\u00`)
				e.WriteByte(hex[b>>4])
				e.WriteByte(hex[b&0xF])
			}
			continue
		}

		c, size := utf8.DecodeRuneInString(s[i:])
		if c == utf8.RuneError && size == 1 {
			e.WriteString(`\ufffd`)
			i++
			continue
		}

		// U+2028 is LINE SEPARATOR.
		// U+2029 is PARAGRAPH SEPARATOR.
		// They are both technically valid characters in JSON strings,
		// but don't work in JSONP, which has to be evaluated as JavaScript,
		// and can lead to security holes there. It is valid JSON to
		// escape them, so do so unconditionally.
		// See http://timelessrepo.com/json-isnt-a-javascript-subset
		if c == '\u2028' || c == '\u2029' {
			e.WriteString(`\u202`)
			e.WriteByte(hex[c&0xF])
			i += size
			continue
		}
		e.WriteString(s[i : i+size])
		i += size
	}
}
