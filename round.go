package wsmock

import (
	"sync"
	"time"
)

type round struct {
	wg       sync.WaitGroup // track if assertions are finished
	jobIndex map[*assertionJob]bool
}

func newRound() *round {
	return &round{
		wg:       sync.WaitGroup{},
		jobIndex: make(map[*assertionJob]bool),
	}
}

func (r *round) addJob(j *assertionJob) (index int) {
	index = len(r.jobIndex) // index is length before adding new job (or new length minus one)
	r.jobIndex[j] = true
	r.wg.Add(1)
	return
}

func (r *round) start(timeout time.Duration) {
	for j := range r.jobIndex {
		go func(j *assertionJob) {
			defer r.wg.Done()
			j.loopWithTimeout(timeout)
		}(j)
	}
}
