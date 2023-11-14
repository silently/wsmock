package wsmock

import (
	"time"
)

type assertionJob struct {
	r *Recorder
	// configuration
	a asserter
	// events
	latestWriteCh chan any
	// state
	done bool
}

func newAssertionJob(r *Recorder, a asserter) *assertionJob {
	return &assertionJob{
		r:             r,
		a:             a,
		latestWriteCh: make(chan any),
		done:          false,
	}
}

func (job *assertionJob) assertOnEnd() {
	latest, _ := last(job.r.serverWrites)
	// on end, done is considered true anyway
	_, passed, errorMessage := job.a.assert(true, latest, job.r.serverWrites)

	if !passed {
		job.r.addError(errorMessage)
	}

	job.done = true
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
			done, passed, errorMessage := job.a.assert(false, latest, job.r.serverWrites)
			if done {
				if !passed {
					job.r.addError(errorMessage)
				}
				job.done = true
				return
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
