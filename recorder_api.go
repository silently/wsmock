package wsmock

import (
	"sync"
	"testing"
	"time"
)

func (r *Recorder) NewChecklist() *Checklist {
	c := &Checklist{r: r}
	r.addToRound(c)
	return c
}

// Adds custom Asserter
func (r *Recorder) AddAsserter(a Asserter) {
	r.NewChecklist().AddAsserter(a)
}

// Adds custom AsserterFunc
func (r *Recorder) Check(f AsserterFunc) {
	r.NewChecklist().Check(f)
}

// Test helpers

// Adds asserter that may succeed on receiving message, and fails if it dit not happen on end
func (r *Recorder) OneToCheck(f FailOnEnd) *Checklist {
	return r.NewChecklist().OneToCheck(f)
}

// Asserts if a message has been received by recorder
func (r *Recorder) OneToBe(target any) *Checklist {
	return r.NewChecklist().OneToBe(target)
}

// Asserts if a message received by recorder contains a given string.
// Messages that can't be converted to strings are JSON-marshalled
func (r *Recorder) OneToContain(sub string) *Checklist {
	return r.NewChecklist().OneToContain(sub)
}

// Asserts first message (times out only if no message is received)
func (r *Recorder) FirstToBe(target any) *Checklist {
	return r.NewChecklist().FirstToBe(target)
}

// Asserts last message (always times out)
func (r *Recorder) LastToBe(target any) *Checklist {
	return r.NewChecklist().LastToBe(target)
}

// Asserts if a message has not been received by recorder (can fail before time out)
func (r *Recorder) OneNotToBe(target any) *Checklist {
	return r.NewChecklist().OneNotToBe(target)
}

// Asserts if conn has been closed
func (r *Recorder) ConnClosed() *Checklist {
	return r.NewChecklist().ConnClosed()
}

// Runs all Assert* methods that have been previously added on this recorder, with a timeout.
//
// If all the assertions succeeds before the timeout, or if one fails before it, timeout won't be reached.
//
// For instance, some assertions (like OneNotToBe) always need to wait until the timeout has been reached
// to assert success, but may fail sooner.
//
// At the end of Run, the recorder previously received messages are flushed and assertions
// are removed. It's then possible to add new Assert* methods and Run again.
func (r *Recorder) RunChecks(timeout time.Duration) {
	r.t.Helper()

	r.startRound(timeout)
	r.waitForRound()
	r.stopRound()
}

// Launches and waits (till timeout) for the outcome of all assertions added to all recorders
// of this test.
func RunChecks(t *testing.T, timeout time.Duration) {
	t.Helper()

	recs := getIndexedRecorders(t)
	wg := sync.WaitGroup{}

	for _, r := range recs {
		wg.Add(1)
		go func(r *Recorder) {
			t.Helper()

			defer wg.Done()
			r.RunChecks(timeout)
		}(r)
	}
	wg.Wait()
}
