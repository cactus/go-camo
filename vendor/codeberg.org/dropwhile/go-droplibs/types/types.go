// Copyright (c) 2015-2026 Eli Janssen
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package types provides some helpful types.
package types

// AnyInteger is a type constraint for the list of interger types
type AnyInteger interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}
