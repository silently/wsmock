package wsmock

import (
	"log"
	"sync"
	"testing"
	"time"
)

type round struct {
	wg             sync.WaitGroup // track if assertions are finished
	timeoutCh      chan struct{}
	doneCh         chan struct{} // caused by timeout or outcome known before timeout (wg passed)
	assertionIndex map[*assertion]bool
}

type Recorder struct {
	t            *testing.T
	currentRound round
	// ws communication
	serverReadCh  chan any
	serverWriteCh chan any
	closed        bool
	closedCh      chan struct{}
	// messages queue
	serverWrites []any
}

func (r *Recorder) newRound() {
	r.currentRound = round{
		wg:             sync.WaitGroup{},
		timeoutCh:      make(chan struct{}),
		doneCh:         make(chan struct{}),
		assertionIndex: make(map[*assertion]bool),
	}
}

func newRecorder(t *testing.T) *Recorder {
	r := Recorder{
		t:             t,
		serverReadCh:  make(chan any, 512),
		serverWriteCh: make(chan any, 512),
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
	if !r.closed {
		r.closed = true
		close(r.closedCh)
	}
	return nil
}

func (r *Recorder) addAssertionToRound(a *assertion) {
	r.currentRound.assertionIndex[a] = true
}

func (r *Recorder) startRound(timeout time.Duration) {
	log.Printf(">>> %+v", r.currentRound.assertionIndex)
	go func() {
		<-time.After(timeout)
		close(r.currentRound.timeoutCh)
	}()

	go r.forwardWritesDuringRound()
	for a := range r.currentRound.assertionIndex {
		go a.loop()
	}
}

func (r *Recorder) waitForRound() {
	r.currentRound.wg.Wait()
}

func (r *Recorder) stopRound() {
	close(r.currentRound.doneCh)
	r.newRound()
}

func (r *Recorder) forwardWritesDuringRound() {
	for {
		select {
		case w := <-r.serverWriteCh:
			r.serverWrites = append(r.serverWrites, w)
			for a := range r.currentRound.assertionIndex {
				go func(a *assertion, w any) { // we don't want to block the loop while assertions haven't started
					a.latestWriteCh <- w
				}(a, w)
			}
		case <-r.currentRound.doneCh:
			// stop forwarding when round ends, serverWriteCh buffers new messages waiting for next round
			return
		}
	}
}
