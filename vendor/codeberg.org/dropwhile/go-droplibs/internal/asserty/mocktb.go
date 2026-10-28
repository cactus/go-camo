// Copyright (c) 2015-2026 Eli Janssen
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.
//
// Inspiration from https://github.com/nalgeon/be

package asserty

import (
	"fmt"
	"testing"
)

// MockTB is a mock implementation of testing.TB
// to capture test failures.
type MockTB struct {
	testing.TB
	Msg     string
	Failed  bool
	Fataled bool
}

// Helper fulfills the Helper interface
func (m *MockTB) Helper() {}

// Cleanup fulfills the TestingT interface
func (m *MockTB) Cleanup(func()) {}

// Fatal fulfills the TestingT interface
func (m *MockTB) Fatal(args ...any) {
	m.Fataled = true
	m.Error(args...)
}

// Fatalf fulfills the TestingT interface
func (m *MockTB) Fatalf(format string, args ...any) {
	m.Fataled = true
	m.Errorf(format, args...)
}

// Error fulfills the TestingT interface
func (m *MockTB) Error(args ...any) {
	m.Failed = true
	m.Msg = fmt.Sprint(args...)
}

// Errorf fulfills the TestingT interface
func (m *MockTB) Errorf(format string, args ...any) {
	m.Failed = true
	m.Msg = fmt.Sprintf(format, args...)
}
