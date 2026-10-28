// Copyright (c) 2015-2026 Eli Janssen
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.
//
// Inspiration from https://github.com/nalgeon/be

// Package asserty is an internal collection of helper code for the assert packages
package asserty

import (
	"bytes"
	"reflect"
	"strings"
)

// HelperT is the interface that wraps the Helper method
type HelperT interface {
	Helper()
}

// Equaler is the interface that wraps the Equal method
type Equaler[T any] interface {
	Equal(T) bool
}

// IsEqual checks if values are equal
func IsEqual[T any](got, want T) bool {
	if IsNil(got) && IsNil(want) {
		return true
	}

	if equalable, ok := any(got).(Equaler[T]); ok {
		return equalable.Equal(want)
	}

	// Special case for byte slices.
	if aBytes, ok := any(got).([]byte); ok {
		bBytes := any(want).([]byte)
		return bytes.Equal(aBytes, bBytes)
	}

	// Fallback to reflective comparison.
	return reflect.DeepEqual(got, want)
}

// IsNil checks if value is nil
func IsNil(v any) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	}
	return false
}

// FormatMsg joins supplied strings (if present) for message formatting.
// If no msg values supplied, returns empty string.
func FormatMsg(msg ...string) string {
	if len(msg) == 0 {
		return ""
	}

	if len(msg[0]) == 0 {
		return ""
	}

	return " " + strings.Join(msg, "; ")
}
