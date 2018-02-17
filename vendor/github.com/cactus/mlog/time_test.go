package mlog

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTime(t *testing.T) {
	loc := time.FixedZone("PDT", -25200)
	cases := []struct {
		F FlagSet
		T time.Time
		R string
	}{
		{
			Lnanoseconds,
			time.Date(2016, time.November, 1, 2, 3, 4, 5, loc),
			`2016-11-01T02:03:04.000000005-07:00`,
		},
		{
			Lnanoseconds,
			time.Date(2016, time.January, 11, 12, 13, 14, 15, time.UTC),
			`2016-01-11T12:13:14.000000015Z`,
		},
		{
			Lmicroseconds,
			time.Date(2016, time.November, 1, 2, 3, 4, 5000, loc),
			`2016-11-01T02:03:04.000005-07:00`,
		},
		{
			Lmicroseconds,
			time.Date(2016, time.January, 11, 12, 13, 14, 15000, time.UTC),
			`2016-01-11T12:13:14.000015Z`,
		},
	}

	b := &sliceBuffer{make([]byte, 0, 1024)}
	for _, tc := range cases {
		b.Truncate(0)
		writeTime(b, &(tc.T), tc.F)
		assert.Equal(t, tc.R, b.String(), "time written incorrectly")
	}
}
