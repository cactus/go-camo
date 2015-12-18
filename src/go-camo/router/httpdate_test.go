package router

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHTTPDateGoroutineUpdate(t *testing.T) {
	t.Parallel()
	d := newHTTPDate()
	n := d.String()
	time.Sleep(2 * time.Second)
	l := d.String()
	assert.NotEqual(t, n, l, "Date did not update as expected: %s == %s", n, l)
}

func TestHTTPDateManualUpdate(t *testing.T) {
	t.Parallel()
	d := &HTTPDate{dateStamp: time.Now().UTC().Format(timeFormat)}
	n := d.String()
	time.Sleep(2 * time.Second)
	d.Update()
	l := d.String()
	assert.NotEqual(t, n, l, "Date did not update as expected: %s == %s", n, l)
}
