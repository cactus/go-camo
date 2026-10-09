// Copyright (c) 2015-2026 Eli Janssen
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package chrono provides a unix epoch time structure that is updated in 1 second intervals
package chrono

import (
	"sync"
	"sync/atomic"
	"time"
)

// TimeNow is a unix epoch time structure that is updated in 1 second
// intervals.
type TimeNow struct {
	tEpoch      int64
	onceUpdater sync.Once
}

// Get returns the current time value as unix epoc.
func (t *TimeNow) Get() int64 {
	return atomic.LoadInt64(&t.tEpoch)
}

// Update forces an update to the current time.
func (t *TimeNow) Update() {
	atomic.StoreInt64(&t.tEpoch, time.Now().UTC().Unix())
}

// NewTimeNow returns a new TimeNow struct
func NewTimeNow() *TimeNow {
	t := &TimeNow{tEpoch: time.Now().UTC().Unix()}
	t.onceUpdater.Do(func() {
		go func() {
			for range time.Tick(1 * time.Second) {
				t.Update()
			}
		}()
	})
	return t
}

// TimeNowString is a formatted utc time string that is updated in 1 second
// intervals.
type TimeNowString struct {
	dateValue   atomic.Value
	format      string
	onceUpdater sync.Once
}

// String fulfills the stringer interface.
// returns the current time value as a string
func (t *TimeNowString) String() string {
	stamp := t.dateValue.Load()
	if stamp == nil {
		ts := time.Now().UTC().Format(t.format)
		t.dateValue.Store(ts)
		return ts
	}
	return stamp.(string)
}

// Update forces an update to the current time.
func (t *TimeNowString) Update() {
	t.dateValue.Store(time.Now().UTC().Format(t.format))
}

// NewTimeNowString returns a new TimeNowString struct
func NewTimeNowString(format string) *TimeNowString {
	t := &TimeNowString{format: format}
	t.Update()
	t.onceUpdater.Do(func() {
		go func() {
			for range time.Tick(1 * time.Second) {
				t.Update()
			}
		}()
	})
	return t
}
