package wsmock

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

// Used in conjunction with a WebSocket connection mock, a Recorder stores all messages written by
// the server to the mock, and its API is used to make assertions about these messages.
type Recorder struct {
	t            *testing.T
	index        int // used in logs
	currentRound *round
	// ws communication
	serverWriteCh chan any
	done          bool
	doneCh        chan struct{}
	// messages queue
	serverWrites []any
	// fails
	mu     sync.RWMutex
	errors []string
}

func newRecorder(t *testing.T) *Recorder {
	r := Recorder{
		t:             t,
		serverWriteCh: make(chan any, 256),
		doneCh:        make(chan struct{}),
	}
	r.index = indexRecorder(t, &r)
	r.resetRound()
	return &r
}

func (r *Recorder) resetRound() {
	r.serverWrites = nil
	r.currentRound = newRound()
}

// called when corresponding conn is closed
func (r *Recorder) stop() {
	if !r.done {
		r.done = true
		close(r.doneCh)
	}
}

func (r *Recorder) forwardWritesDuringRound() {
	for {
		select {
		case w := <-r.serverWriteCh:
			r.serverWrites = append(r.serverWrites, w)

			for job := range r.currentRound.jobIndex {
				if !job.done { // to prevent blocking channel
					job.latestWriteCh <- w
				}
			}
		case <-r.currentRound.doneCh:
			// stop forwarding when round ends, serverWriteCh buffers new messages waiting for next round
			return
		}
	}
}

func (r *Recorder) addError(err string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.errors = append(r.errors, err)
	r.currentRound.stop()
}

func formatErrorSection[T any](r *Recorder, label string, items []T) string {
	output := fmt.Sprintf("Recorder#%v ", r.index) + label + "\n"
	for _, item := range items {
		output = fmt.Sprintf("%v\t%+v\n", output, item)
	}
	return output
}

func (r *Recorder) outputError(err string, isFirst bool) {
	errorParts := strings.Split(err, "\n")
	label, rest := errorParts[0], errorParts[1:]
	errorOutput := formatErrorSection(r, "error: "+label, rest)

	if isFirst {
		num := len(r.serverWrites)
		label := fmt.Sprintf("%v messages received:", num)
		if num == 0 {
			label = "no message received"
		} else if num == 1 {
			label = "1 message received:"
		}
		stateOutput := formatErrorSection(r, label, r.serverWrites)
		r.t.Log("\n" + stateOutput)
	}

	r.t.Error("\n" + errorOutput)
}

func (r *Recorder) manageErrors() {
	r.mu.RLock()
	if len(r.errors) > 0 {
		for i, err := range r.errors {
			r.outputError(err, i == 0)
		}
	}
	r.mu.RUnlock()
}

// API

// Initialize a new chainable AssertionBuilder.
func (r *Recorder) Assert() *AssertionBuilder {
	p := &AssertionBuilder{rec: r}
	j := newAssertionJob(r, p)
	r.currentRound.addJob(j)
	return p
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
func (r *Recorder) RunAssertions(timeout time.Duration) {
	// start
	go r.forwardWritesDuringRound()
	r.currentRound.start(timeout)
	// wait
	r.currentRound.wg.Wait()
	// manage potential assert errors
	r.manageErrors()
	// stop and reset round
	r.currentRound.stop()
	r.resetRound()
}

// Launches and waits (till timeout) for the outcome of all assertions added to all recorders
// of this test.
func RunAssertions(t *testing.T, timeout time.Duration) {
	recs := getIndexedRecorders(t)
	wg := sync.WaitGroup{}

	for _, r := range recs {
		wg.Add(1)
		go func(r *Recorder) {
			defer wg.Done()
			r.RunAssertions(timeout)
		}(r)
	}
	wg.Wait()
}
