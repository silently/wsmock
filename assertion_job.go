package wsmock

import (
	"fmt"
	"time"
)

type assertionJob struct {
	rec   *Recorder
	index int // used in logs
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
	job.index = r.currentRound.addJob(job)
	return job
}

func (j *assertionJob) incPassed() {
	j.currentIndex++
}

func (j *assertionJob) allPassed() bool {
	return len(j.ab.conditions) == j.currentIndex
}

func (j *assertionJob) currentAsserter() Asserter {
	return j.ab.conditions[j.currentIndex]
}

func (j *assertionJob) addError(err string, end bool) {
	prefix := "error on write"
	if end {
		prefix = "error on end"
	}
	// add assert index to error log
	err = fmt.Sprintf(prefix+" in Assert#%v: ", j.index) + err
	j.rec.addError(err)
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
		j.addError(err, true)
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
			// log.Printf(">> %v latest %v", j.index, latest)
			if done {
				j.done = true
				if passed { // current passed
					j.incPassed()
					if j.allPassed() { // all passed
						return
					}
				} else {
					j.addError(err, false)
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
