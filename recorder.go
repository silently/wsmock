package wsmock

import (
	"sync"
	"testing"
)

type Recorder struct {
	// testing logic
	t              *testing.T
	timeoutCh      chan struct{}
	fastExit       bool
	fastExitWG     sync.WaitGroup // if fastExit is enabled, we may not have to wait for timeoutCh to be closed
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

func NewRecorder(t *testing.T) *Recorder {
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

func (r *Recorder) Close() error {
	if !r.closed {
		r.closed = true
		close(r.closedCh)
	}
	return nil
}

func (r *Recorder) resetAssertions() {
	r.timeoutCh = make(chan struct{})
	r.fastExit = true
	r.fastExitWG = sync.WaitGroup{}
	r.assertionWG = sync.WaitGroup{}
	r.assertionIndex = make(map[*assertion]bool)
}

func (r *Recorder) startAssertions() {
	for a := range r.assertionIndex {
		go a.loop()
	}
}

func (r *Recorder) loop() {
	for m := range r.serverWriteCh {
		r.serverWrites = append(r.serverWrites, m)
		for a := range r.assertionIndex {
			go func(a *assertion) { // we don't want to block the loop while assertions haven't started
				a.newWriteCh <- struct{}{}
			}(a)
		}
	}
}
