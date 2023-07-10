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

func (r *Recorder) AssertWith(asserter Asserter) {
	r.addAssertionToRound(newAssertion(r, asserter))
}

// Test helpers

// AssertReceived checks if a message has been received (on each write).
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

// AssertFirstReceived checks first message and returns
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

// AssertLastReceived checks last message when recorder ends
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

// AssertNotReceived checks if a message has not been received (on timeout and close).
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

// AssertReceivedContains checks if a received message contains a given string (on each write).
// Messages that can't be converted to strings are JSON-marshalled
func (r *Recorder) AssertReceivedContains(substr string) {
	r.t.Helper()
	r.AssertWith(func(end bool, latestWrite any, _ []any) (done, passed bool, errorMessage string) {
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

// AssertClosed checks if conn has been closed (on timeout and close)
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

// AssertReceivedSparseSequence checks if a sparse sequence has been received (on each write).
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

// AssertReceivedAdjacentSequence checks if an adjacent sequence has been received (on each write).
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

// AssertReceivedExactSequence checks if this exact sequence has been received (on timeout and close).
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

// Run runs all Assert* methods that have been previously added
// on this recorder, with a timeout.
//
// If all the assertions succeeds before the timeout, or if one fails before it, timeout won't be reached.
//
// For instance, some assertions (like AssertNotReceived) always need to wait until the timeout has been reached
// to assert success, but may fail sooner.
//
// At the end of Run, the recorder keeps previously received messages but assertions
// are removed. It's then possible to add new Assert* methods and Run again.
func (r *Recorder) Run(timeout time.Duration) {
	r.startRound(timeout)
	r.waitForRound()
	r.stopRound()
}

// Run scoped to a test (t *testing.T) launches and waits for all Run of recorders
// that belong to this test.
func Run(t *testing.T, timeout time.Duration) {
	recs := getIndexedRecorders(t)
	wg := sync.WaitGroup{}

	for _, r := range recs {
		wg.Add(1)
		go func(r *Recorder) {
			defer wg.Done()
			r.Run(timeout)
		}(r)
	}
	wg.Wait()
}
