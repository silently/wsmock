package wsmock

import (
	"fmt"
	"time"
)

type assertionJob struct {
	rec   *Recorder
	index int // used in logs
	// configuration
	a *Assertion
	// events
	writeCh chan any
	// message writes history
	writes []any
	// state
	done         bool // means finished, as a success OR failure
	currentIndex int
}

func last[T any](slice []T) (T, bool) {
	if len(slice) == 0 {
		var zero T
		return zero, false
	}
	return slice[len(slice)-1], true
}

func newAssertionJob(r *Recorder, a *Assertion) *assertionJob {
	job := &assertionJob{
		rec:          r,
		a:            a,
		writeCh:      make(chan any, 512),
		done:         false,
		currentIndex: 0,
	}
	job.index = r.currentRound.addJob(job)
	return job
}

func (j *assertionJob) incPassed() {
	j.currentIndex++
}

func (j *assertionJob) allPassed() bool {
	return len(j.a.conditions) == j.currentIndex
}

func (j *assertionJob) currentCondition() Condition {
	return j.a.conditions[j.currentIndex]
}

func (j *assertionJob) addError(err string, end bool) {
	// introduction
	numMessages := len(j.writes)
	messagesLabel := fmt.Sprintf("%v messages received:", numMessages)
	if numMessages == 0 {
		messagesLabel = "no message received"
	} else if numMessages == 1 {
		messagesLabel = "1 message received:"
	}
	output := fmt.Sprintf("\nIn recorder#%v → assertion#%v, ", j.rec.index, j.index) + messagesLabel + "\n"
	for _, item := range j.writes {
		output = fmt.Sprintf("%v\t%#v\n", output, item)
	}
	// actual error
	errorLabel := "Error occured on write:\n\t"
	if end {
		errorLabel = "Error occured on end:\n\t"
	}
	output = output + errorLabel + err + "\n"
	j.rec.addError(output)
}

func (j *assertionJob) assertOnEnd() {
	latest, _ := last(j.writes)
	// on end, done is considered true anyway
	_, currentPassed, currentErr := j.currentCondition().Try(true, latest, j.writes)
	j.done = true

	if currentPassed {
		j.incPassed()
	} else {
		j.addError(currentErr, true)
	}
	if !j.allPassed() {
		j.addError(fmt.Sprintf("only %v/%v condition(s) passed", j.currentIndex, len(j.a.conditions)), true)
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
		case w := <-j.writeCh:
			j.writes = append(j.writes, w)

			currentDone, currentPassed, currentError := j.currentCondition().Try(false, w, j.writes)
			if currentDone {
				if currentPassed { // current passed
					j.incPassed()
					if j.allPassed() { // all passed
						j.done = true
						return
					}
				} else {
					j.done = true
					j.addError(currentError, false)
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
