package wsmock

import (
	"sync"
	"testing"
)

type Recorder struct {
	// testing logic
	t              *testing.T
	timeoutCh      chan struct{}
	assertionWG    sync.WaitGroup // track if assertions are finished
	assertionIndex map[*assertion]bool
	// ws communication
	serverReadCh  chan any
	serverWriteCh chan any
	closed        bool
	closedCh      chan struct{}
	// messages queue
	serverWrites []any
}

func newRecorder(t *testing.T) *Recorder {
	r := Recorder{
		t:             t,
		serverReadCh:  make(chan any, 32),
		serverWriteCh: make(chan any),
		closedCh:      make(chan struct{}),
	}
	r.resetAssertions()
	indexRecorder(t, &r)
	go r.loop() // TODO start loop on StartAssert
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

func (r *Recorder) resetAssertions() {
	r.timeoutCh = make(chan struct{})
	r.assertionWG = sync.WaitGroup{}
	r.assertionIndex = make(map[*assertion]bool)
}

func (r *Recorder) startAssertions() {
	for a := range r.assertionIndex {
		go a.loop()
	}
}

func (r *Recorder) loop() {
	for w := range r.serverWriteCh {
		r.serverWrites = append(r.serverWrites, w)
		for a := range r.assertionIndex {
			go func(a *assertion) { // we don't want to block the loop while assertions haven't started
				a.latestWriteCh <- w
			}(a)
		}
	}
}
