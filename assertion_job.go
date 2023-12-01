package wsmock

import (
	"time"
)

type assertionJob struct {
	r *Recorder
	// configuration
	list []Asserter
	// events
	latestWriteCh chan any
	// state
	done bool
	// optional
	errorMessage string
}

func newAssertionJob(r *Recorder, a Asserter, errorMessage ...string) *assertionJob {
	job := &assertionJob{
		r:             r,
		list:          []Asserter{a},
		latestWriteCh: make(chan any),
		done:          false,
	}
	if len(errorMessage) == 1 {
		job.errorMessage = errorMessage[0]
	}
	return job
}

func (job *assertionJob) assertOnEnd() {
	latest, _ := last(job.r.serverWrites)
	// on end, done is considered true anyway
	_, passed, errorMessage := job.list[0].Try(true, latest, job.r.serverWrites)

	if !passed {
		if len(errorMessage) == 0 {
			errorMessage = job.errorMessage
		}
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
			done, passed, errorMessage := job.list[0].Try(false, latest, job.r.serverWrites)
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
