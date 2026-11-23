package cmd

import (
	"errors"
	"time"
)

// errPollTimeout means the condition was not met within the timeout.
var errPollTimeout = errors.New("timed out")

type pollResult struct {
	done bool
	err  error
}

// pollUntil calls poll every interval until it reports done, fails, or the
// timeout expires. Each call runs in a goroutine, so the timeout holds even
// when a call blocks (kubectl exec has no deadline of its own); a blocked call
// is then abandoned, which is fine for a command about to exit.
func pollUntil(timeout time.Duration, interval time.Duration, poll func() (bool, error)) error {
	deadline := time.After(timeout)
	for {
		result := make(chan pollResult, 1)
		go func() {
			done, err := poll()
			result <- pollResult{done, err}
		}()
		select {
		case r := <-result:
			if r.err != nil || r.done {
				return r.err
			}
		case <-deadline:
			return errPollTimeout
		}
		select {
		case <-time.After(interval):
		case <-deadline:
			return errPollTimeout
		}
	}
}
