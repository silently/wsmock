package wsmock

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

// type Finder func(index int, write any) (ok bool)

// func (r *Recorder) FindOneAfter(f Finder, timeout time.Duration) (found any, ok bool) {
// 	time.Sleep(timeout)
// 	for i, w := range r.serverWrites {
// 		if f(i, w) {
// 			return w, true
// 		}
// 	}
// 	return nil, false
// }

// func (r *Recorder) FindAllAfter(f Finder, timeout time.Duration) (founds []any, ok bool) {
// 	time.Sleep(timeout)
// 	for i, w := range r.serverWrites {
// 		if f(i, w) {
// 			founds = append(founds, w)
// 			ok = true
// 		}
// 	}
// 	return
// }

// Adds custom assertions
func (r *Recorder) AssertWith(asserter Asserter) {
	r.t.Helper()

	r.addAssertionToRound(newAssertion(r, asserter))
}

// Test helpers

// Asserts if a message has been received by recorder
func (r *Recorder) AssertReceived(target any) {
	r.t.Helper()

	r.AssertWith(func(end bool, latestWrite any, _ []any) (done, passed bool, errorMessage string) {
		if end {
			done = true
			passed = false
			errorMessage = fmt.Sprintf("[wsmock] message not received: %v", target)
		} else if latestWrite == target {
			done = true
			passed = true
		}
		return
	})
}

// Asserts first message (times out only if no message is received)
func (r *Recorder) AssertFirstReceived(target any) {
	r.t.Helper()

	r.AssertWith(func(_ bool, _ any, allWrites []any) (done, passed bool, errorMessage string) {
		done = true
		hasReceivedOne := len(allWrites) > 0
		passed = hasReceivedOne && allWrites[0] == target
		if passed {
			return
		}
		if hasReceivedOne {
			errorMessage = fmt.Sprintf("[wsmock] first message should be: %v, received: %v", target, allWrites[0])
		} else {
			errorMessage = fmt.Sprintf("[wsmock] first message should be: %v, received none", target)
		}
		return
	})
}

// Asserts last message (always times out)
func (r *Recorder) AssertLastReceivedOnTimeout(target any) {
	r.t.Helper()

	r.AssertWith(func(end bool, latestWrite any, allWrites []any) (done, passed bool, errorMessage string) {
		if end {
			done = true
			hasReceivedOne := len(allWrites) > 0
			passed = hasReceivedOne && latestWrite == target
			if passed {
				return
			}
			if hasReceivedOne {
				errorMessage = fmt.Sprintf("[wsmock] last message should be: %v, received: %v", target, latestWrite)
			} else {
				errorMessage = fmt.Sprintf("[wsmock] last message should be: %v, received none", target)
			}
		}
		return
	})
}

// Asserts if a message has not been received by recorder
func (r *Recorder) AssertNotReceived(target any) {
	r.t.Helper()

	r.AssertWith(func(end bool, latestWrite any, _ []any) (done, passed bool, errorMessage string) {
		if end {
			done = true
			passed = true
		} else if latestWrite == target {
			done = true
			passed = false
			errorMessage = fmt.Sprintf("[wsmock] message should not have been received: %v", target)
		}
		return
	})
}

// Asserts if a message received by recorder contains a given string.
// Messages that can't be converted to strings are JSON-marshalled
func (r *Recorder) AssertReceivedContains(substr string) {
	r.t.Helper()

	r.AssertWith(func(end bool, latestWrite any, _ []any) (done, passed bool, errorMessage string) {
		r.t.Helper()

		if end {
			done = true
			passed = false
			errorMessage = fmt.Sprintf("[wsmock] no message containing: %v", substr)
		} else if str, ok := latestWrite.(string); ok {
			if strings.Contains(str, substr) {
				done = true
				passed = true
			}
		} else {
			b, err := json.Marshal(latestWrite)
			if err == nil {
				if strings.Contains(string(b), substr) {
					done = true
					passed = true
				}
			}
		}
		return
	})
}

// Asserts if conn has been closed (on timeout and close)
func (r *Recorder) AssertClosed() {
	r.t.Helper()

	r.AssertWith(func(end bool, latestWrite any, allWrites []any) (done, passed bool, errorMessage string) {
		if end {
			passed = r.closed
			errorMessage = "[wsmock] should be closed"
		}
		return
	})
}

// Asserts if a sparse sequence has been received.
//
// If the messages received are (1, 2, 3, 4, 5), included sparse sequences are for instance
// (1, 2, 3, 5) or (2, 4), but neither (2, 4, 1) nor (1, 2, 6).
func (r *Recorder) AssertReceivedSparseSequence(targets []any) {
	r.t.Helper()

	r.AssertWith(func(end bool, _ any, allWrites []any) (done, passed bool, errorMessage string) {
		if end {
			done = true
			passed = false
			errorMessage = fmt.Sprintf("[wsmock] sparse sequence not received: %v", targets)
		} else {
			targetLength := len(targets)
			found := 0
			for _, m := range allWrites {
				if m == targets[found] {
					found++
				}
				if found >= targetLength {
					break
				}
			}
			if found == targetLength {
				done = true
				passed = true
			}
		}
		return
	})
}

// Asserts if an adjacent sequence has been received.
//
// If the messages received are (1, 2, 3, 4, 5), included adjacent sequences are for instance
// (2, 3, 4, 5) or (1, 2), but neither (1, 3) nor (4, 5, 6).
func (r *Recorder) AssertReceivedAdjacentSequence(targets []any) {
	r.t.Helper()

	r.AssertWith(func(end bool, latestWrite any, allWrites []any) (done, passed bool, errorMessage string) {
		if end {
			done = true
			passed = false
			errorMessage = fmt.Sprintf("[wsmock] adjacent sequence not received: %v", targets)
		} else {
			targetLength := len(targets)
			found := 0
			started := false
			for _, m := range allWrites {
				if m == targets[found] {
					found++
					started = true
					if found >= targetLength {
						break
					}
				} else if started {
					break
				}
			}
			if found == targetLength {
				done = true
				passed = true
			}
		}
		return
	})
}

// Asserts if this exact sequence has been received (always times out).
//
// If the messages received are (1, 2, 3, 4, 5), the only valid sequence is (1, 2, 3, 4, 5).
func (r *Recorder) AssertReceivedExactSequence(targets []any) {
	r.t.Helper()

	r.AssertWith(func(end bool, latestWrite any, allWrites []any) (done, passed bool, errorMessage string) {
		if end {
			done = true
			errorMessage = fmt.Sprintf("[wsmock] exact sequence not received: %v", targets)
			if len(targets) != len(allWrites) {
				passed = false
				return
			}
			for i := range targets {
				if targets[i] != allWrites[i] {
					passed = false
					return
				}
			}
			passed = true
		}
		return
	})
}

// Runs all Assert* methods that have been previously added on this recorder, with a timeout.
//
// If all the assertions succeeds before the timeout, or if one fails before it, timeout won't be reached.
//
// For instance, some assertions (like AssertNotReceived) always need to wait until the timeout has been reached
// to assert success, but may fail sooner.
//
// At the end of Run, the recorder previously received messages are flushed and assertions
// are removed. It's then possible to add new Assert* methods and Run again.
func (r *Recorder) Run(timeout time.Duration) {
	r.t.Helper()

	r.startRound(timeout)
	r.waitForRound()
	r.stopRound()
}

// Launches and waits (till timeout) for the outcome of all assertions added to all recorders
// of this test.
func Run(t *testing.T, timeout time.Duration) {
	t.Helper()

	recs := getIndexedRecorders(t)
	wg := sync.WaitGroup{}

	for _, r := range recs {
		wg.Add(1)
		go func(r *Recorder) {
			r.t.Helper()

			defer wg.Done()
			r.Run(timeout)
		}(r)
	}
	wg.Wait()
}
