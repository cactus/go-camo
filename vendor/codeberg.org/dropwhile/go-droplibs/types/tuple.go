// Copyright (c) 2015-2026 Eli Janssen
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package types provides some helpful types.
package types

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
)

// Tuple is a Pair container type
type Tuple[L any, R any] struct {
	Left  L
	Right R
}

// Unpack returns all values
func (t Tuple[L, R]) Unpack() (L, R) {
	return t.Left, t.Right
}

// Swap returns the Left value in Right and vice versa.
func (t Tuple[L, R]) Swap() Tuple[R, L] {
	return Tuple[R, L]{t.Right, t.Left}
}

// MarshalJSONTo encodes t as a JSON array into enc.
func (t Tuple[L, R]) MarshalJSONTo(enc *jsontext.Encoder) error {
	if err := enc.WriteToken(jsontext.BeginArray); err != nil {
		return err
	}

	if err := json.MarshalEncode(enc, t.Left); err != nil {
		return err
	}

	if err := json.MarshalEncode(enc, t.Right); err != nil {
		return err
	}

	return enc.WriteToken(jsontext.EndArray)
}

// UnmarshalJSONFrom decodes a JSON array from dec into t.
func (t *Tuple[L, R]) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	if k := dec.PeekKind(); k != '[' {
		return &json.SemanticError{JSONKind: k}
	}

	if _, err := dec.ReadToken(); err != nil {
		return err
	}

	if dec.PeekKind() == ']' {
		return fmt.Errorf("list ends before Left")
	}

	var left L
	if err := json.UnmarshalDecode(dec, &left); err != nil {
		return err
	}

	if dec.PeekKind() == ']' {
		return fmt.Errorf("list ends before Right")
	}

	var right R
	if err := json.UnmarshalDecode(dec, &right); err != nil {
		return err
	}

	if dec.PeekKind() != ']' {
		return fmt.Errorf("list does not end after Right")
	}

	if _, err := dec.ReadToken(); err != nil {
		return err
	}

	t.Left = left
	t.Right = right
	return nil
}
