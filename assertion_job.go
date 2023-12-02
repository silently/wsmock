package wsmock

import (
	"time"
)

type assertionJob struct {
	r *Recorder
	// configuration
	c *Checklist
	// events
	latestWriteCh chan any
	// state
	done         bool // means finished, as a success OR failure
	currentIndex int
	// optional
	err string
}

func newAssertionJob(r *Recorder, c *Checklist, err ...string) *assertionJob {
	job := &assertionJob{
		r:             r,
		c:             c,
		latestWriteCh: make(chan any, 256),
		done:          false,
		currentIndex:  0,
	}
	if len(err) == 1 {
		job.err = err[0]
	}
	return job
}

func (j *assertionJob) incPassed() {
	j.currentIndex++
}

func (j *assertionJob) allPassed() bool {
	return len(j.c.list) == j.currentIndex
}

func (j *assertionJob) currentAsserter() Asserter {
	return j.c.list[j.currentIndex]
}

func (job *assertionJob) assertOnEnd() {
	latest, _ := last(job.r.serverWrites)
	// on end, done is considered true anyway
	_, passed, err := job.currentAsserter().Try(true, latest, job.r.serverWrites)
	job.done = true

	if passed {
		job.incPassed()
	} else {
		if len(err) == 0 {
			err = job.err
		}
		job.r.addError(err)
	}
}

func (job *assertionJob) loopWithTimeout(timeout time.Duration) {
	// we found that using time.Sleep is more accurate (= less delay in addition to the specified timeout)
	// than using <-time.After directly on a for-select case
	timeoutCh := make(chan string, 1)
	go func() {
		time.Sleep(timeout)
		timeoutCh <- "timeout"
	}()

	for {
		select {
		case latest := <-job.latestWriteCh:
			done, passed, err := job.currentAsserter().Try(false, latest, job.r.serverWrites)
			if done {
				job.done = true
				if passed { // current passed
					job.incPassed()
					if job.allPassed() { // all passed
						return
					}
				} else {
					job.r.addError(err)
					return
				}
			}
		case <-job.r.currentRound.doneCh: // round is done because of another failing assertion
			job.assertOnEnd()
			return
		case <-job.r.doneCh: // conn is closed
			job.assertOnEnd()
			return
		case <-timeoutCh: // timeout is reached
			job.assertOnEnd()
			return
		}
	}
}
