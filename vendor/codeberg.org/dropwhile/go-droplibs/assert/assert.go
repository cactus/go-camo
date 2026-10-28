// Copyright (c) 2015-2026 Eli Janssen
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.
//
// Inspiration from https://github.com/nalgeon/be

// Package assert is a minimal Go test package.
package assert

import (
	"errors"
	"reflect"
	"regexp"
	"strings"

	"codeberg.org/dropwhile/go-droplibs/internal/asserty"
)

// TestingT is the subset of [testing.T] (see also [testing.TB]) used by the assert package.
type TestingT interface {
	Error(args ...any)
	Errorf(format string, args ...any)
	Fatal(args ...any)
	Fatalf(format string, args ...any)
	Cleanup(func())
}

// Assert is an alias for True
var Assert = True

// True fails assertion if got is false.
func True(t TestingT, got bool, msg ...string) {
	if ht, ok := t.(asserty.HelperT); ok {
		ht.Helper()
	}

	if !got {
		t.Fatalf("got: %t; want: %t;%s", got, true, asserty.FormatMsg(msg...))
	}
}

// False fails assertion if got is true.
func False(t TestingT, got bool, msg ...string) {
	if ht, ok := t.(asserty.HelperT); ok {
		ht.Helper()
	}

	if got {
		t.Fatalf("got: %t; want: %t;%s", got, false, asserty.FormatMsg(msg...))
	}
}

// Equal fails assertion if got and want are not equal.
func Equal[T any](t TestingT, got, want T, msg ...string) {
	if ht, ok := t.(asserty.HelperT); ok {
		ht.Helper()
	}

	if !asserty.IsEqual(got, want) {
		t.Fatalf("got: %#v; want: %#v;%s", got, want, asserty.FormatMsg(msg...))
	}
}

// NotEqual fails assertion if got and want are equal.
func NotEqual[T any](t TestingT, got, want T, msg ...string) {
	if ht, ok := t.(asserty.HelperT); ok {
		ht.Helper()
	}

	if asserty.IsEqual(got, want) {
		t.Fatalf("got: %#v; expected values to be different;%s", got, asserty.FormatMsg(msg...))
	}
}

// Nil fails assertion if got is not nil.
func Nil(t TestingT, got any, msg ...string) {
	if ht, ok := t.(asserty.HelperT); ok {
		ht.Helper()
	}

	if !asserty.IsNil(got) {
		t.Fatalf("got: %#v; want: <nil>;%s", got, asserty.FormatMsg(msg...))
	}
}

// NotNil fails assertion if got is nil.
func NotNil(t TestingT, got any, msg ...string) {
	if ht, ok := t.(asserty.HelperT); ok {
		ht.Helper()
	}

	if asserty.IsNil(got) {
		t.Fatalf("got: <nil>; expected non-nil;%s", asserty.FormatMsg(msg...))
	}
}

// Error fails assertion if got and want are not "error similar".
// want can be one of: nil, a string, an error, or an interface.
//
// Some examples...
//
// assert that an error value matches (string match)
//
//	assert.Error(t, err, "my bad error")
//
// assert error value matches (error match)
//
//	assert.Error(t, err, err2)
//
// assert that an error value matches, which
// works with wrapped errors, using errors.Is under the hood
//
//	wrappedErr := fmt.Errorf("wrapped: %w", err)
//	assert.Error(t, err, wrappedErr)
//
// can also check for error type, using errors.As under the hood
//
//	assert.Error(t, err, reflect.TypeFor[*fs.PathError]())
func Error(t TestingT, got error, want any, msg ...string) {
	if ht, ok := t.(asserty.HelperT); ok {
		ht.Helper()
	}

	switch w := want.(type) {
	case nil:
		if got != nil {
			t.Fatalf("unexpected error: %s;%s", got, asserty.FormatMsg(msg...))
		}
	case string:
		if got == nil || !strings.Contains(got.Error(), w) {
			t.Fatalf("got: %q; want: %q;%s", got, want, asserty.FormatMsg(msg...))
		}
	case error:
		if !errors.Is(got, w) {
			if asserty.IsNil(got) {
				t.Fatalf("got: <nil>; want: %T(%v);%s", w, w, asserty.FormatMsg(msg...))
			} else {
				t.Fatalf("got: %T(%v); want: %T(%v);%s", got, got, w, w, asserty.FormatMsg(msg...))
			}
		}
	case reflect.Type:
		target := reflect.New(w).Interface()
		if !errors.As(got, target) {
			t.Fatalf("got: %T; want: %v;%s", got, w, asserty.FormatMsg(msg...))
		}
	default:
		t.Fatalf("unsupported want type: %T", want)
	}
}

// MatchesRegex asserts that the string matches the regex pattern.
//
// Example
//
//	assert.MatchesRegexp(t, "abc123d", `abc[123]+$`)
func MatchesRegex(t TestingT, got, pattern string, msg ...string) {
	if ht, ok := t.(asserty.HelperT); ok {
		ht.Helper()
	}

	if matched, err := regexp.MatchString(pattern, got); err != nil {
		t.Fatalf("unable to parse regexp pattern %s: %s", pattern, err.Error())
	} else if !matched {
		t.Fatalf("got: %q; want to match %q;%s", got, pattern, asserty.FormatMsg(msg...))
	}
}
