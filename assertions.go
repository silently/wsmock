package wsmock

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

func toAnySlice[T any](expecteds []T) []any {
	expectedsAny := make([]any, len(expecteds))
	for i, v := range expecteds {
		expectedsAny[i] = v
	}
	return expectedsAny
}

type assertion struct {
	recorder *Recorder
	// configuration
	assertOnWrite bool
	errMessage    string
	finder        Finder
	// events
	newWriteCh chan struct{}
}

func newAssertion(r *Recorder, fastExit bool, errMessage string, f Finder) *assertion {
	a := assertion{
		recorder:      r,
		assertOnWrite: fastExit,
		errMessage:    errMessage,
		finder:        f,
		newWriteCh:    make(chan struct{}),
	}
	r.assertionWG.Add(1)
	if fastExit {
		r.fastExitWG.Add(1)
	}
	go func() {
		// first try if serverWrites already contains messages
		a.newWriteCh <- struct{}{}
	}()
	return &a
}

func (a assertion) loop() {
	defer a.recorder.assertionWG.Done()

assertLoop:
	// breaking loop causes test to fail
	// returning prevents test from faililng
	for {
		select {
		case <-a.newWriteCh: // for assertions that are tried after every new write
			if a.assertOnWrite && a.finder(a.recorder.serverWrites) {
				a.recorder.fastExitWG.Done()
				return
			}
		case <-a.recorder.timeoutCh: // for assertions that are tried on timeout
			if !a.assertOnWrite && a.finder(a.recorder.serverWrites) {
				return
			} else {
				break assertLoop
			}
		case <-a.recorder.closedCh: // for assertions that are tried on timeout
			if !a.assertOnWrite && a.finder(a.recorder.serverWrites) {
				return
			} else {
				a.recorder.fastExitWG.Done()
				break assertLoop
			}
		}
	}

	a.recorder.t.Error(a.errMessage)
}

func (r *Recorder) addAssert(fastExit bool, errMessage string, f Finder) {
	r.t.Helper()
	r.fastExit = r.fastExit && fastExit
	a := newAssertion(r, fastExit, errMessage, f)
	r.assertionIndex[a] = true
}

// generic assertions

func (r *Recorder) AssertOnWrite(errMessage string, f Finder) {
	r.t.Helper()
	r.addAssert(true, errMessage, f)
}

func (r *Recorder) AssertOnTimeoutOrClose(errMessage string, f Finder) {
	r.t.Helper()
	r.addAssert(false, errMessage, f)
}

// assertions provided predefined Finder functions

// AssertReceived checks if a message has been received (on each write).
func (r *Recorder) AssertReceived(expected any) {
	r.t.Helper()
	r.AssertOnWrite(fmt.Sprintf("[wsmock] message not received: %v", expected), func(messages []any) bool {
		for _, m := range messages {
			if m == expected {
				return true
			}
		}
		return false
	})
}

// AssertReceivedContains checks if a received message contains a given string (on each write).
func (r *Recorder) AssertReceivedContains(expected string) {
	r.t.Helper()
	r.AssertOnWrite(fmt.Sprintf("[wsmock] no message contained: %v", expected), func(messages []any) bool {
		for _, m := range messages {
			if str, ok := m.(string); ok {
				if strings.Contains(str, expected) {
					return true
				}
			}
		}
		return false
	})
}

// AssertReceivedSparseSequence checks if a sparse sequence has been received (on each write).
//
// If the messages received are (1, 2, 3, 4, 5), included sparse sequences are for instance
// (1, 2, 3, 5) or (2, 4), but neither (2, 4, 1) nor (1, 2, 6).
func (r *Recorder) AssertReceivedSparseSequence(expecteds []any) {
	r.t.Helper()

	r.AssertOnWrite("[wsmock] sparse message sequence not received", func(messages []any) bool {
		length := len(expecteds)
		found := 0
		for _, m := range messages {
			if m == expecteds[found] {
				found++
			}
			if found >= length {
				break
			}
		}
		return found == length
	})
}

// AssertReceivedAdjacentSequence checks if an adjacent sequence has been received (on each write).
//
// If the messages received are (1, 2, 3, 4, 5), included adjacent sequences are for instance
// (2, 3, 4, 5) or (1, 2), but neither (1, 3) nor (4, 5, 6).
func (r *Recorder) AssertReceivedAdjacentSequence(expecteds []any) {
	r.t.Helper()

	r.AssertOnWrite("[wsmock] adjacent message sequence not received", func(messages []any) bool {
		length := len(expecteds)
		found := 0
		started := false
		for _, m := range messages {
			if m == expecteds[found] {
				found++
				started = true
				if found >= length {
					break
				}
			} else if started {
				break
			}
		}
		return found == length
	})
}

// AssertReceivedExactSequence checks if this exact sequence has been received (on timeout and close).
//
// If the messages received are (1, 2, 3, 4, 5), the only valid sequence is (1, 2, 3, 4, 5).
func (r *Recorder) AssertReceivedExactSequence(expecteds []any) {
	r.t.Helper()

	r.AssertOnTimeoutOrClose("[wsmock] exact message sequence not received", func(messages []any) bool {
		if len(expecteds) != len(messages) {
			return false
		}
		for i := range expecteds {
			if expecteds[i] != messages[i] {
				return false
			}
		}
		return true
	})
}

// AssertNotReceived checks if a message has not been received (on timeout and close).
func (r *Recorder) AssertNotReceived(expected any) {
	r.t.Helper()
	r.AssertOnTimeoutOrClose("[wsmock] message received but should not", func(messages []any) bool {
		for _, m := range messages {
			if m == expected {
				return false
			}
		}
		return true
	})
}

// AssertClosed checks if conn has been closed (on timeout and close)
func (r *Recorder) AssertClosed() {
	r.t.Helper()
	r.AssertOnTimeoutOrClose("[wsmock] should be closed", func(messages []any) bool {
		return r.closed
	})
}

// RunAssertions runs all Assert* methods that have been previously added (invoked)
// on this recorder, with a timeout.
//
// When all the assertions occur on write, they may all succeed before the timeout: in that
// case RunAssertions returns before timeout has been reached.
//
// If at least one assertion does not succeed on writes that occur within the timeout,
// RunAssertions will go till this timeout.
//
// Some assertions (like AssertNotReceived) always wait until the timeout has been reached
// to assert success, but may fail sooner.
//
// At the end of RunAssertions, the recorder keeps previously received messages but assertions
// are removed. It's then possible to add new Assert* methods and RunAssertions again.
func (r *Recorder) RunAssertions(timeout time.Duration) {
	r.startAssertions()
	fastExitCh := make(chan struct{})
	if r.fastExit {
		go func() {
			r.fastExitWG.Wait()
			close(fastExitCh)
		}()
	}
	// wait for end of assert group
	select {
	case <-time.After(timeout):
		close(r.timeoutCh)
		break
	case <-fastExitCh:
		break
	}
	// wait gracefully for assertions to be indeed finished
	r.assertionWG.Wait()
	r.resetAssertions()
}

// RunAssertions scoped to a test (t *testing.T) launches and waits for all RunAssertions of recorders
// that belong to this test.
func RunAssertions(t *testing.T, timeout time.Duration) {
	recs := getIndexedRecorders(t)
	wg := sync.WaitGroup{}

	for _, r := range recs {
		wg.Add(1)
		go func(r *Recorder) {
			r.RunAssertions(timeout)
			wg.Done()
		}(r)
	}
	wg.Wait()
}
