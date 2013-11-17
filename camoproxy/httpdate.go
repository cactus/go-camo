package camoproxy

import (
	"sync"
	"time"
)

const TimeFormat = "Mon, 02 Jan 2006 15:04:05 GMT"

// holds current date stamp formatting for http Date header
type HttpDate struct {
	dateStamp   string
	mu          sync.RWMutex
	onceUpdater sync.Once
}

func (h *HttpDate) String() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.dateStamp
}

func (h *HttpDate) Update() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.dateStamp = time.Now().UTC().Format(TimeFormat)
}

func newHttpDate() *HttpDate {
	d := &HttpDate{dateStamp: time.Now().UTC().Format(TimeFormat)}
	// spawn a single formattedDate updater
	d.onceUpdater.Do(func() {
		go func() {
			<-time.After(time.Second)
			d.Update()
		}()
	})
	return d
}

var formattedDate = newHttpDate()
