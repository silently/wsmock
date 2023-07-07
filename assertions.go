package wsmock

import (
	"time"
)

// Asserter functions are added to Recorders with AssertWith and are called possibly several times
// during the same test to determine the outcome of the test.
//
// Case 1: when a new write occurs (on the underlying Conn, and thus on the associated Recorder),
// Asserter is called with the following arguments: false, latestWrite, allWritesIncludingLatest.
//
// Possible outcomes of Case 1 are:
// - if a decision can't be made about the assertion being true or not (e.g. if more data or timeout needed)
// *done* is false and the other return values don't matter
// - the test succeeds, *done* and *passed* are true, *errorMessage* does not matter
// - the test fails, *done* is true, *passed* is false and *errorMessage* will be used to print an error
//
// Case 2: when the recorder ends (the underlying Conn is closed or the test times out),
// Asserter is called with the following arguments: true, latestWrite, allWritesIncludingLatest. Contrary to Case 1,
// we don't know if there was indeed a *latestWrite* (could be nil, like *allWritesIncludingLatest*).
//
// Case 2 is the last time Asserter will be called (*done* return value should not be taken into account, but considered true),
// *passed* gives the test outcome and *errorMessage* is used if test failed.
type Asserter func(end bool, latestWrite any, allWrites []any) (done, passed bool, errorMessage string)

type assertion struct {
	recorder *Recorder
	// configuration
	asserter Asserter
	// events
	latestWriteCh chan any
}

func newAssertion(r *Recorder, asserter Asserter) *assertion {
	return &assertion{
		recorder:      r,
		asserter:      asserter,
		latestWriteCh: make(chan any),
	}
}

func (a assertion) assertOnEnd() {
	latest, _ := Last(a.recorder.serverWrites)
	// on end, done is considered true anyway
	_, passed, errorMessage := a.asserter(true, latest, a.recorder.serverWrites)

	if !passed {
		a.recorder.t.Error(errorMessage)
	}
}

func (a assertion) loopWithTimeout(timeout time.Duration) {
	// we found that using time.Sleep is more accurate (= less delay in addition to the specified timeout)
	// than using <-time.After directly on a for-select case
	timeoutCh := make(chan string, 1)
	go func() {
		time.Sleep(timeout)
		timeoutCh <- "timeout"
	}()

	for {
		select {
		case latest := <-a.latestWriteCh:
			done, passed, errorMessage := a.asserter(false, latest, a.recorder.serverWrites)
			if done {
				if !passed {
					a.recorder.t.Error(errorMessage)
				}
				return
			}
		case <-timeoutCh:
			a.assertOnEnd()
			return
		case <-a.recorder.closedCh:
			a.assertOnEnd()
			return
		}
	}
}
