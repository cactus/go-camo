// Copyright (c) 2012-2016 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mlog

import (
	"sort"
	"strings"
)

const (
	// Bits or'ed together to control what's printed.

	// Ltimestamp specifies to log the date+time stamp
	Ltimestamp FlagSet = 1 << iota
	// Lmicroseconds specifies to use microsecond timestamp granularity in
	// Ltimestamp.
	Lmicroseconds
	// Lnanoseconds specifies to use nanosecond timestamp granularity in
	// Ltimestamp. overrides Lmicroseconds.
	Lnanoseconds
	// Llevel specifies to log message level.
	Llevel
	// Llongfile specifies to log file path and line number: /a/b/c/d.go:23
	Llongfile
	// Lshortfile specifies to log file name and line number: d.go:23.
	// overrides Llongfile.
	Lshortfile
	// Lsort specifies to sort Map key value pairs in output.
	Lsort
	// Ldebug specifies to enable debug level logging.
	Ldebug
	// Lstd is the standard log format if none is specified.
	Lstd = Ltimestamp | Llevel | Lsort
)

var flagNames = map[FlagSet]string{
	Ltimestamp:    "Ltimestamp",
	Lmicroseconds: "Lmicroseconds",
	Lnanoseconds:  "Lnanoseconds",
	Llevel:        "Llevel",
	Llongfile:     "Llongfile",
	Lshortfile:    "Lshortfile",
	Lsort:         "Lsort",
	Ldebug:        "Ldebug",
}

// FlagSet defines the output formatting flags (bitfield) type, which define
// certainly fields to appear in the output.
type FlagSet uint64

// Has returns true if the FlagSet argument is in the set of flags (binary &)
func (f *FlagSet) Has(p FlagSet) bool {
	if *f&p != 0 {
		return true
	}
	return false
}

// GoString fulfills the GoStringer interface, defining the format used for
// the %#v format string.
func (f FlagSet) GoString() string {
	s := make([]byte, 0, len(flagNames))
	var p uint64
	for p = 256; p > 0; p >>= 1 {
		if f&FlagSet(p) != 0 {
			s = append(s, '1')
		} else {
			s = append(s, '0')
		}
	}
	return string(s)
}

// String fulfills the Stringer interface, defining the format used for
// the %s format string.
func (f FlagSet) String() string {
	flags := make([]string, 0, len(flagNames))
	for k, v := range flagNames {
		if f&k != 0 {
			flags = append(flags, v)
		}
	}
	sort.Strings(flags)
	return "FlagSet(" + strings.Join(flags, "|") + ")"
}
