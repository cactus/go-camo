// Copyright (c) 2012-2016 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package router

import (
	"sync"
	"time"
)

const timeFormat = "Mon, 02 Jan 2006 15:04:05 GMT"

// HTTPDate holds current date stamp formatting for HTTP date header
type HTTPDate struct {
	dateStamp   string
	mu          sync.RWMutex
	onceUpdater sync.Once
}

func (h *HTTPDate) String() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.dateStamp
}

func (h *HTTPDate) Update() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.dateStamp = time.Now().UTC().Format(timeFormat)
}

func newHTTPDate() *HTTPDate {
	d := &HTTPDate{dateStamp: time.Now().UTC().Format(timeFormat)}
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

var formattedDate = newHTTPDate()
