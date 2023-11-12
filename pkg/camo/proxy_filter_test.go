// Copyright (c) 2012-2023 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package camo

import (
	"fmt"
	"net/url"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
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
	assert.Check(t, err)
	_, err = processRequest(req, 200, camoConfig, filters)
	assert.Check(t, err)
	assert.Check(t, called, "filter func wasn't called")
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
	assert.Check(t, err)
	_, err = processRequest(req, 404, camoConfig, filters)
	assert.Check(t, err)
	assert.Check(t, called, "filter func wasn't called")
}

func TestFilterListMatrixMultiples(t *testing.T) {
	t.Parallel()

	testURL := "http://www.google.com/images/srpr/logo11w.png"
	req, err := makeReq(camoConfig, testURL)
	assert.Check(t, err)
	type chkResponse struct {
		chk bool
		err error
	}

	mixtests := []struct {
		filterRuleAnswers  []chkResponse
		expectedCallMatrix []bool
		respcode           int
	}{
		// all rules return true, so all rules should have been called
		// so pass: http200
		{
			[]chkResponse{{true, nil}, {true, nil}, {true, nil}},
			[]bool{true, true, true},
			200,
		},

		// 3rd rule should not be called, because 2nd returned false
		// so no pass: http404
		{
			[]chkResponse{{true, nil}, {false, nil}, {true, nil}},
			[]bool{true, true, false},
			404,
		},
		// 3rd rule should not be called, because 2nd returned an error
		// so no pass: http404
		{
			[]chkResponse{{true, nil}, {true, fmt.Errorf("some error")}, {true, nil}},
			[]bool{true, true, false},
			404,
		},

		// 2nd, 3rd rules should not be called, because 1st returned false
		// so no pass: http404
		{
			[]chkResponse{{false, nil}, {false, nil}, {true, nil}},
			[]bool{true, false, false},
			404,
		},
		// 2nd, 3rd rules should not be called, because 1st returned an error
		// so no pass: http404
		{
			[]chkResponse{{true, fmt.Errorf("some error")}, {false, nil}, {true, nil}},
			[]bool{true, false, false},
			404,
		},

		// last rule returns false, but all rules should be called.
		// so no pass: http404
		{
			[]chkResponse{{true, nil}, {true, nil}, {false, nil}},
			[]bool{true, true, true},
			404,
		},
		// last rule returns an error, but all rules should be called.
		// so no pass: http404
		{
			[]chkResponse{{true, nil}, {true, nil}, {true, fmt.Errorf("some error")}},
			[]bool{true, true, true},
			404,
		},
	}

	for _, tt := range mixtests {
		callMatrix := []bool{false, false, false}
		filters := make([]FilterFunc, 0)
		for i := 0; i < 3; i++ {
			filters = append(
				filters, func(x int) FilterFunc {
					return func(*url.URL) (bool, error) {
						callMatrix[x] = true
						return tt.filterRuleAnswers[x].chk, tt.filterRuleAnswers[x].err
					}
				}(i),
			)
		}

		_, err = processRequest(req, tt.respcode, camoConfig, filters)
		assert.Check(t, err)
		for i := range callMatrix {
			assert.Check(t, is.Equal(callMatrix[i],
				tt.expectedCallMatrix[i]), fmt.Sprintf(
				"filter func called='%t' wanted '%t'",
				callMatrix[i], tt.expectedCallMatrix[i],
			))
		}
	}
}
