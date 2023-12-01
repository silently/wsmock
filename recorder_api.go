package wsmock

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

func join(data []any) string {
	var output string
	for i, d := range data {
		output += fmt.Sprintf("%+v", d)
		if i != len(data)-1 {
			output += "\n"
		}
	}
	return output
}

// Adds custom Asserter
func (r *Recorder) AddAsserter(a Asserter) {
	r.addToRound(newAssertionJob(r, a))
}

// Adds custom AsserterFunc
func (r *Recorder) AddAsserterFunc(a AsserterFunc) {
	r.addToRound(newAssertionJob(r, a))
}

// Adds asserter that may succeed on receiving message, and fails if it dit not happen on end
func (r *Recorder) ToMatchOnReceive(a AsserterOnReceive, errorMessage string) {
	r.addToRound(newAssertionJob(r, a, errorMessage))
}

// Adds asserter that tried on end (timeout or connection closed)
func (r *Recorder) ToMatchOnEnd(a AsserterOnEnd, errorMessage string) {
	r.addToRound(newAssertionJob(r, a, errorMessage))
}

// Test helpers

// Asserts if a message has been received by recorder
func (r *Recorder) ToReceive(target any) {
	r.ToMatchOnReceive(func(msg any) bool {
		return msg == target
	}, fmt.Sprintf("message not received\nexpected: %+v", target))
}

// Asserts if a message received by recorder contains a given string.
// Messages that can't be converted to strings are JSON-marshalled
func (r *Recorder) ToReceiveContaining(substr string) {
	r.ToMatchOnReceive(func(msg any) bool {
		if str, ok := msg.(string); ok {
			return strings.Contains(str, substr)
		} else {
			b, err := json.Marshal(msg)
			if err == nil {
				return strings.Contains(string(b), substr)
			}
		}
		return false
	}, fmt.Sprintf("no message containing string\nexpected: %v", substr))
}

// Asserts first message (times out only if no message is received)
func (r *Recorder) ToReceiveFirst(target any) {
	r.AddAsserterFunc(func(_ bool, _ any, all []any) (done, passed bool, errorMessage string) {
		done = true
		hasReceivedOne := len(all) > 0
		passed = hasReceivedOne && all[0] == target
		if passed {
			return
		}
		if hasReceivedOne {
			errorMessage = fmt.Sprintf("incorrect first message\nexpected: %+v\nreceived: %+v", target, all[0])
		} else {
			errorMessage = fmt.Sprintf("incorrect first message\nexpected: %+v\nreceived none", target)
		}
		return
	})
}

// Asserts last message (always times out)
func (r *Recorder) ToReceiveLast(target any) {
	r.ToMatchOnEnd(func(all []any) bool {
		length := len(all)
		return length > 0 && all[length-1] == target
	}, fmt.Sprintf("incorrect last message on timeout\nexpected: %+v", target))
}

// Asserts if a message has not been received by recorder (can fail before time out)
func (r *Recorder) NotToReceive(target any) {
	r.AddAsserterFunc(func(end bool, latest any, _ []any) (done, passed bool, errorMessage string) {
		if end {
			done = true
			passed = true
		} else if latest == target {
			done = true
			passed = false
			errorMessage = fmt.Sprintf("message should not be received\nunexpected: %+v", target)
		}
		return
	})
}

// Asserts if conn has been closed
func (r *Recorder) ToBeClosed() {
	r.AddAsserterFunc(func(end bool, latest any, all []any) (done, passed bool, errorMessage string) {
		if end {
			passed = r.done // conn closed => recorder done
			errorMessage = "conn should be closed"
		}
		return
	})
}

// Asserts if an adjacent sequence has been received (possibly within a larger sequence)
//
// If the messages received are (1, 2, 3, 4, 5), included adjacent sequences are for instance
// (2, 3, 4, 5) or (1, 2), but neither (1, 3) nor (4, 5, 6).
func (r *Recorder) ToReceiveSequence(targets []any) {
	r.AddAsserterFunc(func(end bool, latest any, all []any) (done, passed bool, errorMessage string) {
		if end {
			done = true
			passed = false
			errorMessage = "adjacent sequence not received\n" + join(targets)
		} else {
			targetLength := len(targets)
			found := 0
			started := false
			for _, m := range all {
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

// Asserts if a sparse sequence has been received.
//
// If the messages received are (1, 2, 3, 4, 5), included sparse sequences are for instance
// (1, 2, 3, 5) or (2, 4), but neither (2, 4, 1) nor (1, 2, 6).
func (r *Recorder) ToReceiveSparseSequence(targets []any) {
	r.AddAsserterFunc(func(end bool, _ any, all []any) (done, passed bool, errorMessage string) {
		if end {
			done = true
			passed = false
			errorMessage = "sparse sequence not received\n" + join(targets)
		} else {
			targetLength := len(targets)
			found := 0
			for _, m := range all {
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

// Asserts if this exact sequence has been received (always times out).
//
// If the messages received are (1, 2, 3, 4, 5), the only valid sequence is (1, 2, 3, 4, 5).
func (r *Recorder) ToReceiveOnlySequence(targets []any) {
	// could be optimized by failing sooner if we know before the end the exact sequence does not match
	r.ToMatchOnEnd(func(all []any) bool {
		if len(targets) != len(all) {
			return false
		}
		for i := range targets {
			if targets[i] != all[i] {
				return false
			}
		}
		return true
	}, "exact sequence not received\n"+join(targets))
}

// Runs all Assert* methods that have been previously added on this recorder, with a timeout.
//
// If all the assertions succeeds before the timeout, or if one fails before it, timeout won't be reached.
//
// For instance, some assertions (like NotToReceive) always need to wait until the timeout has been reached
// to assert success, but may fail sooner.
//
// At the end of Run, the recorder previously received messages are flushed and assertions
// are removed. It's then possible to add new Assert* methods and Run again.
func (r *Recorder) RunAssertions(timeout time.Duration) {
	r.t.Helper()

	r.startRound(timeout)
	r.waitForRound()
	r.stopRound()
}

// Launches and waits (till timeout) for the outcome of all assertions added to all recorders
// of this test.
func RunAssertions(t *testing.T, timeout time.Duration) {
	t.Helper()

	recs := getIndexedRecorders(t)
	wg := sync.WaitGroup{}

	for _, r := range recs {
		wg.Add(1)
		go func(r *Recorder) {
			t.Helper()

			defer wg.Done()
			r.RunAssertions(timeout)
		}(r)
	}
	wg.Wait()
}
