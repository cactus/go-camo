// Copyright (c) 2012-2023 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package camo

import (
	"encoding/json"
	"fmt"
	"net/url"
	"testing"

	"github.com/dropwhile/assert"
)

func TestFilterListAcceptSimple(t *testing.T) {
	t.Parallel()

	called := false
	filters := []FilterFunc{
		func(*url.URL) (bool, error) {
			called = true
			return true, nil
		},
	}
	testURL := "http://www.google.com/images/srpr/logo11w.png"
	req, err := makeReq(camoConfig, testURL)
	assert.Nil(t, err)
	_, err = processRequest(req, 200, camoConfig, filters)
	assert.Nil(t, err)
	assert.True(t, called, "filter func wasn't called")
}

func TestFilterListAcceptSimpleWithFilterError(t *testing.T) {
	t.Parallel()

	called := false
	filters := []FilterFunc{
		func(*url.URL) (bool, error) {
			called = true
			return true, fmt.Errorf("bad hostname")
		},
	}
	testURL := "http://www.google.com/images/srpr/logo11w.png"
	req, err := makeReq(camoConfig, testURL)
	assert.Nil(t, err)
	_, err = processRequest(req, 404, camoConfig, filters)
	assert.Nil(t, err)
	assert.True(t, called, "filter func wasn't called")
}

func TestFilterListMatrixMultiples(t *testing.T) {
	t.Parallel()

	testURL := "http://www.google.com/images/srpr/logo11w.png"
	req, err := makeReq(camoConfig, testURL)
	assert.Nil(t, err)

	type errResp []Tuple[bool, error]

	f := func(input string, expectedCallMatrix []bool, respcode int) {
		t.Helper()
		callMatrix := []bool{false, false, false}
		filters := make([]FilterFunc, 0)
		var responses errResp
		if err := json.Unmarshal([]byte(input), &responses); err != nil {
			t.Fatal(err)
		}

		for i := 0; i < 3; i++ {
			filters = append(
				filters, func(x int) FilterFunc {
					return func(*url.URL) (bool, error) {
						callMatrix[x] = true
						return responses[x].a, responses[x].b
					}
				}(i),
			)
		}
		_, err = processRequest(req, respcode, camoConfig, filters)
		assert.Nil(t, err)
		for i := range callMatrix {
			assert.Equal(t, callMatrix[i], expectedCallMatrix[i],
				fmt.Sprintf(
					"filter func called='%t'[%d] wanted '%t'",
					callMatrix[i], i, expectedCallMatrix[i],
				),
			)
		}
	}
	f(
		"[ [true, null], [true, null], [true, null] ]",
		[]bool{true, true, true},
		200,
	)

	// all rules return true, so all rules should have been called
	// so pass: http200
	f(
		"[ [true, null], [true, null], [true, null] ]",
		[]bool{true, true, true},
		200,
	)

	// 3rd rule should not be called, because 2nd returned false
	// so no pass: http404
	f(
		"[ [true, null], [false, null], [true, null] ]",
		[]bool{true, true, false},
		404,
	)

	// 3rd rule should not be called, because 2nd returned an error
	// so no pass: http404
	f(
		"[ [true, null], [true, \"some error\"], [true, null] ]",
		[]bool{true, true, false},
		404,
	)

	// 2nd, 3rd rules should not be called, because 1st returned false
	// so no pass: http404
	f(
		"[ [false, null], [false, null], [true, null] ]",
		[]bool{true, false, false},
		404,
	)

	// 2nd, 3rd rules should not be called, because 1st returned an error
	// so no pass: http404
	f(
		"[ [true, \"some error\"], [false, null], [true, null] ]",
		[]bool{true, false, false},
		404,
	)

	// last rule returns false, but all rules should be called.
	// so no pass: http404
	f(
		"[ [true, null], [true, null], [false, null] ]",
		[]bool{true, true, true},
		404,
	)

	// last rule returns an error, but all rules should be called.
	// so no pass: http404
	f(
		"[ [true, null], [true, null], [true, \"some error\"] ]",
		[]bool{true, true, true},
		404,
	)
}
