// Copyright (c) 2012-2016 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mlog

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var update = flag.Bool("update", false, "update golden files")

func TestLoggerMsgs(t *testing.T) {
	var infoTests = map[string]struct {
		flags   FlagSet
		method  string
		message string
		extra   interface{}
	}{
		"infom1": {Llevel | Lsort, "infom", "test", Map{"x": "y"}},
		"infom2": {Llevel | Lsort, "infom", "test", Map{"x": "y", "y": "z", "t": "u", "u": "v"}},
		"infom3": {Llevel | Lsort, "infom", "test", Map{"y": "z", "x": "y", "u": "v", "t": "u"}},
		"infom4": {Llevel | Lsort, "infom", "test", Map{"x": 1, "y": 2, "z": 3, "haz_string": "such tests"}},
		"debug1": {Llevel | Lsort | Ldebug, "debugm", "test", nil},
		"debug2": {Llevel | Lsort | Ldebug, "debugm", "test", nil},
		"infof1": {Llevel, "infof", "test: %d", 5},
		"infof2": {Llevel, "infof", "test: %s", "test"},
		"infof3": {Llevel, "infof", "test: %s %s", []interface{}{"test", "pickles"}},
	}

	buf := &bytes.Buffer{}
	logger := New(ioutil.Discard, Llevel|Lsort)
	logger.out = buf

	for name, tt := range infoTests {
		buf.Truncate(0)
		logger.flags = uint64(tt.flags)

		switch tt.method {
		case "debugm":
			m, ok := tt.extra.(Map)
			if !ok && tt.extra != nil {
				t.Errorf("%s: failed type assertion", name)
				continue
			}
			logger.Debugm(tt.message, m)
		case "infom":
			m, ok := tt.extra.(Map)
			if !ok && tt.extra != nil {
				t.Errorf("%s: failed type assertion", name)
				continue
			}
			logger.Infom(tt.message, m)
		case "debug":
			logger.Debug(tt.message)
		case "info":
			logger.Info(tt.message)
		case "debugf":
			if i, ok := tt.extra.([]interface{}); ok {
				logger.Debugf(tt.message, i...)
			} else {
				logger.Debugf(tt.message, tt.extra)
			}
		case "infof":
			if i, ok := tt.extra.([]interface{}); ok {
				logger.Infof(tt.message, i...)
			} else {
				logger.Infof(tt.message, tt.extra)
			}
		default:
			t.Errorf("%s: not sure what to do", name)
			continue
		}
		actual := buf.Bytes()
		golden := filepath.Join("test-fixtures", fmt.Sprintf("test_logger_msgs.%s.golden", name))
		if *update {
			ioutil.WriteFile(golden, actual, 0644)
		}
		expected, _ := ioutil.ReadFile(golden)
		assert.Equal(t, string(expected), string(actual), "%s: did not match expectation", name)
	}

}

func TestLoggerTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}

	// test nanoseconds
	logger := New(buf, Lstd|Lnanoseconds)
	tnow := time.Now()
	logger.Info("test this")
	ts := bytes.Split(buf.Bytes()[6:], []byte{'"'})[0]
	tlog, err := time.Parse(time.RFC3339Nano, string(ts))
	assert.Nil(t, err, "Failed to parse time from log")
	assert.WithinDuration(t, tnow, tlog, 2*time.Second, "Time not even close")

	buf.Truncate(0)

	// test microeconds
	logger.SetFlags(Lstd | Lmicroseconds)
	tnow = time.Now()
	logger.Info("test this")
	ts = bytes.Split(buf.Bytes()[6:], []byte{'"'})[0]
	tlog, err = time.Parse(time.RFC3339Nano, string(ts))
	assert.Nil(t, err, "Failed to parse time from log")
	assert.WithinDuration(t, tnow, tlog, 2*time.Second, "Time not even close")

	buf.Truncate(0)

	// test standard (seconds)
	logger.SetFlags(Lstd)
	tnow = time.Now()
	logger.Info("test this")
	ts = bytes.Split(buf.Bytes()[6:], []byte{'"'})[0]
	tlog, err = time.Parse(time.RFC3339Nano, string(ts))
	assert.Nil(t, err, "Failed to parse time from log")
	assert.WithinDuration(t, tnow, tlog, 2*time.Second, "Time not even close")
}
