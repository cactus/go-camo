// Copyright (c) 2012-2018 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package camo

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterListAcceptSimple(t *testing.T) {
	t.Parallel()

	called := false
	filters := []FilterFunc{
		func(*url.URL) bool {
			called = true
			return true
		},
	}
	testURL := "http://www.google.com/images/srpr/logo11w.png"
	req, err := makeReq(camoConfig, testURL)
	assert.Nil(t, err)
	_, err = processRequest(req, 200, camoConfig, filters)
	assert.Nil(t, err)
	assert.True(t, called, "filter func wasn't called")
}

func TestFilterListMatrixMultiples(t *testing.T) {
	t.Parallel()

	testURL := "http://www.google.com/images/srpr/logo11w.png"
	req, err := makeReq(camoConfig, testURL)
	assert.Nil(t, err)

	var mixtests = []struct {
		filterRuleAnswers  []bool
		expectedCallMatrix []bool
		respcode           int
	}{
		// all rules return true, so all rules should have been called
		// so pass: http200
		{
			[]bool{true, true, true},
			[]bool{true, true, true},
			200,
		},
		// 3rd rule should not be called, because 2nd returned false
		// so no pass: http404
		{
			[]bool{true, false, true},
			[]bool{true, true, false},
			404,
		},
		// 2nd, 3rd rules should not be called, because 1st returned false
		// so no pass: http404
		{
			[]bool{false, false, true},
			[]bool{true, false, false},
			404,
		},
		// last rule returns false, but all rules should be called.
		// so no pass: http404
		{
			[]bool{true, true, false},
			[]bool{true, true, true},
			404,
		},
	}

	for _, tt := range mixtests {
		callMatrix := []bool{false, false, false}
		filters := make([]FilterFunc, 0)
		for i := 0; i < 3; i++ {
			filters = append(
				filters, func(x int) func(*url.URL) bool {
					return func(*url.URL) bool {
						callMatrix[x] = true
						return tt.filterRuleAnswers[x]
					}
				}(i),
			)
		}

		_, err = processRequest(req, tt.respcode, camoConfig, filters)
		assert.Nil(t, err)
		for i := range callMatrix {
			assert.Equal(t,
				callMatrix[i],
				tt.expectedCallMatrix[i],
				fmt.Sprintf(
					"filter func called='%t' wanted '%t'",
					callMatrix[i], tt.expectedCallMatrix[i],
				),
			)
		}
	}
}
