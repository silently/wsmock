package wsmock

import (
	"sync"
	"testing"
	"time"
)

// Used in conjunction with a WebSocket connection mock, a Recorder stores all messages written by
// the WebSocket server handler to the mock.
//
// Its API is used to define assertions about these messages.
type Recorder struct {
	t            *testing.T
	index        int // used in logs
	currentRound *round
	// ws communication
	writeCh chan any
	done    bool
	doneCh  chan struct{}
	// when fails
	mu     sync.RWMutex
	errors []string
}

func newRecorder(t *testing.T) *Recorder {
	r := Recorder{
		t:       t,
		writeCh: make(chan any, 512),
		doneCh:  make(chan struct{}),
	}
	r.index = indexRecorder(t, &r)
	r.resetRound()
	return &r
}

func (r *Recorder) resetRound() {
	r.currentRound = newRound()
}

// called when corresponding conn is closed
func (r *Recorder) stop() {
	if !r.done {
		r.done = true
		close(r.doneCh)
	}
}

// forward to assertionJobs
func (r *Recorder) forwardWritesDuringRound() {
	for {
		select {
		case w := <-r.writeCh:
			for job := range r.currentRound.jobIndex {
				if !job.done { // to prevent blocking channel
					job.writeCh <- w
				}
			}
		case <-r.doneCh:
			// stop forwarding when round ends, serverWriteCh buffers new messages waiting for next round
			return
		}
	}
}

func (r *Recorder) addError(err string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.errors = append(r.errors, err)
}

func (r *Recorder) outputError(err string) {
	r.t.Helper()
	r.t.Error(err)
}

func (r *Recorder) manageErrors() {
	r.t.Helper()

	r.mu.RLock()
	if len(r.errors) > 0 {
		for _, err := range r.errors {
			r.outputError(err)
		}
	}
	r.mu.RUnlock()
}

// API

// Initialize a new chainable Assertion
func (r *Recorder) NewAssertion() *Assertion {
	p := &Assertion{}
	newAssertionJob(r, p)
	return p
}

// type Selector func(w any) (val any, ok bool)

// func WaitFor[T any](rec *Recorder, sel Selector) (val T, ok bool) {
// 	for {
// 		select {
// 		case w := <-rec.serverWriteCh:
// 			if v, ok := sel(w); ok {
// 				if s, ok := v.(T); ok {
// 					return s, true
// 				}
// 			}
// 		case <-rec.doneCh:
// 			return *new(T), false
// 		}
// 	}
// }

// func Try() {
// 	x, ok := WaitFor[float32](nil, func(w any) (v any, ok bool) {
// 		return 1, true
// 	})
// // other implemention could look like
// rec.WaitFor(pointer, filter, timeout)
// }

// Runs all the assertions added on this recorder with NewAssertion() and waits for their outcome.
//
// The specified timeout is not reached in the following cases:
// - all of the assertions succeed
// - one assertion fails
// - the conn is closed
// - (or if there is no assertion on the recorder)
//
// For instance some conditions (like NoneToBe) always need to wait until the timeout is reached
// to succeed, but may fail sooner.
//
// At the end of RunAssertions, the recorder message history is flushed and assertions
// are removed. It's then possible to add new assertions and run them with a fresh history
// on the same recorder.
func (r *Recorder) RunAssertions(timeout time.Duration) {
	r.t.Helper()

	// start
	go r.forwardWritesDuringRound()
	r.currentRound.start(timeout)
	// wait
	r.currentRound.wg.Wait()
	// manage potential assert errors
	r.manageErrors()
	// stop and reset round
	r.resetRound()
}

// Runs and waits for the outcome of all the assertions added to all the recorders
// of this T test.
func RunAssertions(t *testing.T, timeout time.Duration) {
	t.Helper()

	recs := getIndexedRecorders(t)
	wg := sync.WaitGroup{}

	for _, r := range recs {
		wg.Add(1)
		go func(r *Recorder) {
			r.t.Helper()

			defer wg.Done()
			r.RunAssertions(timeout)
		}(r)
	}
	wg.Wait()
}
