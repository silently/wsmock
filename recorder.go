package wsmock

import (
	"sync"
	"testing"
	"time"
)

type round struct {
	wg             sync.WaitGroup // track if assertions are finished
	doneCh         chan struct{}  // caused by timeout or outcome known before timeout (wg passed)
	assertionIndex map[*assertion]bool
}

type Recorder struct {
	t            *testing.T
	currentRound *round
	// ws communication
	serverReadCh  chan any
	serverWriteCh chan any
	closed        bool
	closedCh      chan struct{}
	// messages queue
	serverWrites []any
}

func (r *Recorder) newRound() {
	r.t.Helper()

	r.currentRound = &round{
		wg:             sync.WaitGroup{},
		doneCh:         make(chan struct{}),
		assertionIndex: make(map[*assertion]bool),
	}
}

func newRecorder(t *testing.T) *Recorder {
	t.Helper()

	r := Recorder{
		t:             t,
		serverReadCh:  make(chan any, 256),
		serverWriteCh: make(chan any, 256),
		closedCh:      make(chan struct{}),
	}
	r.newRound()
	indexRecorder(t, &r)
	t.Cleanup(func() {
		unindexRecorder(t, &r)
	})
	return &r
}

func (r *Recorder) close() error {
	r.t.Helper()

	if !r.closed {
		r.closed = true
		close(r.closedCh)
	}
	return nil
}

func (r *Recorder) addAssertionToRound(a *assertion) {
	r.t.Helper()

	r.currentRound.assertionIndex[a] = true
	r.currentRound.wg.Add(1)
}

func (r *Recorder) startRound(timeout time.Duration) {
	r.t.Helper()

	go r.forwardWritesDuringRound()
	for a := range r.currentRound.assertionIndex {
		go func(a *assertion) {
			a.recorder.t.Helper()

			defer r.currentRound.wg.Done()
			a.loopWithTimeout(timeout)
		}(a)
	}
}

func (r *Recorder) waitForRound() {
	r.t.Helper()

	r.currentRound.wg.Wait()
}

func (r *Recorder) stopRound() {
	r.t.Helper()

	close(r.currentRound.doneCh)
	r.newRound()
}

func (r *Recorder) forwardWritesDuringRound() {
	r.t.Helper()

	for {
		select {
		case w := <-r.serverWriteCh:
			r.serverWrites = append(r.serverWrites, w)
			for a := range r.currentRound.assertionIndex {
				a.latestWriteCh <- w
			}
		case <-r.currentRound.doneCh:
			// stop forwarding when round ends, serverWriteCh buffers new messages waiting for next round
			return
		}
	}
}
