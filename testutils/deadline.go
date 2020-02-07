package testutils

import (
	"testing"
	"time"
)

func Deadline(t *testing.T, dur time.Duration, fChan <-chan func(t *testing.T)) {
	t.Helper()

	timer := time.NewTimer(dur)
	defer timer.Stop()

	select {
	case <-timer.C:
		t.Fatal("timeout waiting")
	case f := <-fChan:
		f(t)
	}
}
