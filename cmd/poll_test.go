package cmd

import (
	"errors"
	"testing"
	"time"
)

// A poll that never returns (e.g. a kubectl exec hanging) must not hold the
// caller past its timeout.
func TestPollUntilTimesOutOnBlockedPoll(t *testing.T) {
	start := time.Now()
	err := pollUntil(100*time.Millisecond, 10*time.Millisecond, func() (bool, error) {
		select {}
	})
	if !errors.Is(err, errPollTimeout) {
		t.Errorf("expected errPollTimeout, got %v", err)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("returned after %s, expected about 100ms", elapsed)
	}
}

func TestPollUntilDone(t *testing.T) {
	calls := 0
	err := pollUntil(time.Second, time.Millisecond, func() (bool, error) {
		calls++
		return calls == 3, nil
	})
	if err != nil || calls != 3 {
		t.Errorf("got err %v after %d calls, expected nil after 3", err, calls)
	}
}

func TestPollUntilError(t *testing.T) {
	boom := errors.New("boom")
	err := pollUntil(time.Second, time.Millisecond, func() (bool, error) {
		return false, boom
	})
	if !errors.Is(err, boom) {
		t.Errorf("expected boom, got %v", err)
	}
}

func TestPollUntilTimesOut(t *testing.T) {
	err := pollUntil(50*time.Millisecond, 5*time.Millisecond, func() (bool, error) {
		return false, nil
	})
	if !errors.Is(err, errPollTimeout) {
		t.Errorf("expected errPollTimeout, got %v", err)
	}
}
