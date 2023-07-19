package wsmock

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

type round struct {
	wg             sync.WaitGroup // track if assertions are finished
	done           bool
	doneCh         chan struct{} // caused by timeout or outcome known before timeout (wg passed)
	assertionIndex map[*assertion]bool
}

// closed when corresponding conn is
func (rd *round) stop() error {
	if !rd.done {
		rd.done = true
		close(rd.doneCh)
	}
	return nil
}

// Used in conjunction with a WebSocket connection mock, a Recorder stores all messages written by
// the server to the mock, and its API (Assert* methods) is used to make assertions about these messages.
type Recorder struct {
	t            *testing.T
	index        int // used in logs
	currentRound *round
	// ws communication
	serverReadCh  chan any
	serverWriteCh chan any
	done          bool
	doneCh        chan struct{}
	// messages queue
	serverWrites []any
	// fails
	mu     sync.RWMutex
	errors []string
}

func (r *Recorder) newRound() {
	r.serverWrites = nil
	r.currentRound = &round{
		wg:             sync.WaitGroup{},
		doneCh:         make(chan struct{}),
		assertionIndex: make(map[*assertion]bool),
	}
}

func newRecorder(t *testing.T) *Recorder {
	r := Recorder{
		t:             t,
		serverReadCh:  make(chan any, 256),
		serverWriteCh: make(chan any, 256),
		doneCh:        make(chan struct{}),
	}
	r.index = indexRecorder(t, &r)
	r.newRound()
	return &r
}

// closed when corresponding conn is
func (r *Recorder) stop() error {
	if !r.done {
		r.done = true
		close(r.doneCh)
	}
	return nil
}

func (r *Recorder) addAssertionToRound(a *assertion) {
	r.currentRound.assertionIndex[a] = true
	r.currentRound.wg.Add(1)
}

func (r *Recorder) addError(errorMessage string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.errors = append(r.errors, errorMessage)
	r.currentRound.stop()
}

func (r *Recorder) startRound(timeout time.Duration) {
	go r.forwardWritesDuringRound()
	for a := range r.currentRound.assertionIndex {
		go func(a *assertion) {
			defer r.currentRound.wg.Done()
			a.loopWithTimeout(timeout)
		}(a)
	}
}

func formatErrorSection[T any](r *Recorder, label string, items []T) string {
	output := fmt.Sprintf("Recorder#%v ", r.index) + label + "\n"
	for _, item := range items {
		output = fmt.Sprintf("%v\t%+v\n", output, item)
	}
	return output
}

func (r *Recorder) error(errorMessage string, isFirst bool) {
	r.t.Helper()

	errorParts := strings.Split(errorMessage, "\n")
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

func (r *Recorder) waitForRound() {
	r.t.Helper()

	r.currentRound.wg.Wait()

	if len(r.errors) > 0 {
		r.mu.RLock()
		for i, errorMessage := range r.errors {
			r.error(errorMessage, i == 0)
		}
		r.mu.RUnlock()
	}
}

func (r *Recorder) stopRound() {
	r.currentRound.stop()
	r.newRound()
}

func (r *Recorder) forwardWritesDuringRound() {
	for {
		select {
		case w := <-r.serverWriteCh:
			r.serverWrites = append(r.serverWrites, w)

			for a := range r.currentRound.assertionIndex {
				if !a.done { // to prevent blocking channel
					a.latestWriteCh <- w
				}
			}
		case <-r.currentRound.doneCh:
			// stop forwarding when round ends, serverWriteCh buffers new messages waiting for next round
			return
		}
	}
}
