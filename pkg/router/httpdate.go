// Copyright (c) 2012-2018 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package router

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/cactus/mlog"
)

const timeFormat = "Mon, 02 Jan 2006 15:04:05 GMT"

// HTTPDate holds current date stamp formatting for HTTP date header
type iHTTPDate struct {
	dateValue   atomic.Value
	onceUpdater sync.Once
}

func (h *iHTTPDate) String() string {
	stamp := h.dateValue.Load()
	if stamp == nil {
		mlog.Print("got a nil datesamp. Trying to recover...")
		h.Update()
		return time.Now().UTC().Format(timeFormat)
	}
	return stamp.(string)
}

func (h *iHTTPDate) Update() {
	h.dateValue.Store(time.Now().UTC().Format(timeFormat))
}

func newiHTTPDate() *iHTTPDate {
	d := &iHTTPDate{}
	d.Update()
	// spawn a single formattedDate updater
	d.onceUpdater.Do(func() {
		go func() {
			for range time.Tick(1 * time.Second) {
				d.Update()
			}
		}()
	})
	return d
}

var formattedDate = newiHTTPDate()
