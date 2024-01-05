package wsmock

import "fmt"

// Generic interface that is chained to form Assertions. The Try method of a Condition is called in two cases:
//
// Case 1: every time a message is sent (on the underlying Conn, and thus on the associated Recorder),
// Try is called with (false, latest, all) where *all* is "all messages including the *latest*"
//
// Possible outcomes of Case 1 are:
// - if a decision can't be made about the assertion being true or not (e.g. if more data or reaching the timeout is needed)
// *done* is false and the other return values don't matter
// - if the test succeeds, *done* and *passed* are true, *err* does not matter
// - if test fails, *done* is true, *passed* is false and *err* is used to print an error
//
// Case 2: when end is reached (on timeout or when connection is closed),
// Try is called one last time with (true, latest, all). Contrary to Case 1,
// we don't know if there was indeed a *latest* (it could be nil, like *all*).
//
// For case 2, the return value `done` is considered true whatever is returned, while `passed` and `err`
// do give the test outcome.
type Condition interface {
	Try(end bool, latest any, all []any) (done, passed bool, err string)
}

// The ConditionFunc type is an adapter to allow the use of a function as an Condition.
//
// Its signature and behaviour has to match Condition' Try method.
type ConditionFunc func(end bool, latest any, all []any) (done, passed bool, err string)

func (f ConditionFunc) Try(end bool, latest any, all []any) (done, passed bool, err string) {
	return f(end, latest, all)
}

// The oneTo struct implements Condition. Its Predicate function is called on each message and on end.
//
// If the Predicate returns true, asserting is done and succeeds,
// If the Predicate returns false, asserting is not done,
// If the end is reached, asserting is done and fails.
type oneTo predicateAndError

func newOneTo(f Predicate, err string) *oneTo {
	return &oneTo{f, err}
}

func (c oneTo) Try(end bool, latest any, _ []any) (done, passed bool, err string) {
	// fails on end
	if end {
		return true, false, c.err
	}
	if c.f(latest) { // succeeds
		return true, true, ""
	}
	// unfinished
	return false, false, ""
}

// The nextTo struct implements Condition. Its Predicate function is called once, either on the (next) message
// or on timeout.
//
// The Predicate return value gives the test outcome (success/failure).
type nextTo predicateAndError

func newNextTo(f Predicate, err string) *nextTo {
	return &nextTo{f, err}
}

func (c nextTo) Try(end bool, latest any, _ []any) (done, passed bool, err string) {
	// fails on end
	if end {
		return true, false, c.err
	} else if c.f(latest) {
		return true, true, ""
	} else {
		return true, c.f(latest), c.err + fmt.Sprintf("\n\tFailing message (of type %T): %+v", latest, latest)
	}
}

// The lastTo struct implements Condition. Its Predicate function is called once, on end.
//
// The Predicate return value gives the test outcome (success/failure).
type lastTo predicateAndError

func newLastTo(f Predicate, err string) *lastTo {
	return &lastTo{f, err}
}

func (c lastTo) Try(end bool, latest any, _ []any) (done, passed bool, err string) {
	// fails on end
	if end {
		if latest != nil {
			if c.f(latest) {
				return true, true, ""
			} else {
				return true, false, c.err + fmt.Sprintf("\nFailing message (of type %T): %+v", latest, latest)
			}
		} else {
			return true, false, c.err + "\nReason: last message missing" // no "last" -> fails
		}
	}
	// unfinished
	return false, false, ""
}

// The allTo struct implements Condition. Its Predicate function is called on each message and on end.
//
// If the Predicate returns true, asserting is not done,
// If the Predicate returns false, asserting is done and fails,
// If the end is reached, asserting is done and succeeds.
type allTo predicateAndError

func newAllTo(p Predicate, err string) *allTo {
	return &allTo{p, err}
}

func (c allTo) Try(end bool, latest any, _ []any) (done, passed bool, err string) {
	if end {
		return true, true, ""
	} else {
		if c.f(latest) {
			return false, false, "" // ongoing
		} else {
			return true, false, c.err + fmt.Sprintf("\nFailing message (of type %T): %+v", latest, latest) // failed
		}
	}
}
