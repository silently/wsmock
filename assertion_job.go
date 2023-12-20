package wsmock

import (
	"time"
)

type assertionJob struct {
	rec *Recorder
	// configuration
	ab *AssertionBuilder
	// events
	latestWriteCh chan any
	// state
	done         bool // means finished, as a success OR failure
	currentIndex int
	// optional
	err string
}

func newAssertionJob(r *Recorder, ab *AssertionBuilder) *assertionJob {
	job := &assertionJob{
		rec:           r,
		ab:            ab,
		latestWriteCh: make(chan any, 256),
		done:          false,
		currentIndex:  0,
	}
	return job
}

func (j *assertionJob) incPassed() {
	j.currentIndex++
}

func (j *assertionJob) allPassed() bool {
	return len(j.ab.list) == j.currentIndex
}

func (j *assertionJob) currentAsserter() Asserter {
	return j.ab.list[j.currentIndex]
}

func (j *assertionJob) assertOnEnd() {
	latest, _ := last(j.rec.serverWrites)
	// on end, done is considered true anyway
	_, passed, err := j.currentAsserter().Try(true, latest, j.rec.serverWrites)
	j.done = true

	if passed {
		j.incPassed()
	} else {
		if len(err) == 0 {
			err = j.err
		}
		j.rec.addError(err)
	}
}

// Deals with messages forwarded by recorder, send them to condition and manage condition progress,
// also dealing with ending logic.
func (j *assertionJob) loopWithTimeout(timeout time.Duration) {
	// we found that using time.Sleep is more accurate (= less delay in addition to the specified timeout)
	// than using <-time.After directly on a for-select case
	timeoutCh := make(chan string, 1)
	go func() {
		time.Sleep(timeout)
		timeoutCh <- "timeout"
	}()

	for {
		select {
		case latest := <-j.latestWriteCh:
			done, passed, err := j.currentAsserter().Try(false, latest, j.rec.serverWrites)
			if done {
				j.done = true
				if passed { // current passed
					j.incPassed()
					if j.allPassed() { // all passed
						return
					}
				} else {
					j.rec.addError(err)
					return
				}
			}
		case <-j.rec.doneCh: // conn is closed
			j.assertOnEnd()
			return
		case <-timeoutCh: // timeout is reached
			j.assertOnEnd()
			return
		}
	}
}
