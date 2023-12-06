package wsmock

import "fmt"

// Generic interface to be added on recorders
// The Try method is called on *end* (timeout or connection closed)
// and each time a message is received (*latest* argument), with *all* being "all messages including the *latest*"
//
// Case 1: when a message is sent (on the underlying Conn, and thus on the associated Recorder),
// Try is called with (false, message, all).
//
// Possible outcomes of Case 1 are:
// - if a decision can't be made about the assertion being true or not (e.g. if more data or timeout needed)
// *done* is false and the other return values don't matter
// - if the test succeeds, *done* and *passed* are true, *err* does not matter
// - if test fails, *done* is true, *passed* is false and *err* is used to print an error
//
// Case 2: when end is reached (timeout or connection closed),
// Try is called one last time with (true, message, all). Contrary to Case 1,
// we don't know if there was indeed a *latest* (it could be nil, like *all*).
//
// For case 2, the return value `done` is considered true whatever is returned, while `passed` and `err`
// do give the test outcome.
type Asserter interface {
	Try(end bool, latest any, all []any) (done, passed bool, err string)
}

// The AsserterFunc type is an adapter to allow the use of an ordinary functions as an Asserter.
//
// Its signature and behaviour follows Asserters Try.
type AsserterFunc func(end bool, latest any, all []any) (done, passed bool, err string)

func (f AsserterFunc) Try(end bool, latest any, all []any) (done, passed bool, err string) {
	return f(end, latest, all)
}

// A Predicate function maps its input to true or false.
type Predicate func(msg any) (passed bool)

type predicateAndError struct {
	f   Predicate
	err string
}

// The oneTo struct implements Asserter. Its Predicate function is called each on message and on end.
//
// If the Predicate returns true, asserting is done and succeeds,
// If the Predicate returns false, asserting is not done,
// If the end is reached, asserting is done and fails.
type oneTo predicateAndError

func newOneTo(f Predicate, err string) *oneTo {
	return &oneTo{f, err}
}

func (a oneTo) Try(end bool, latest any, _ []any) (done, passed bool, err string) {
	// fails on end
	if end {
		return true, false, a.err
	}
	if a.f(latest) { // succeeds
		return true, true, ""
	}
	// unfinished
	return false, false, ""
}

// The nextTo struct implements Asserter. Its Predicate function is called once, either on the (next) message
// or on timeout.
//
// The Predicate return value gives the test outcome (success/failure).
type nextTo predicateAndError

func newNextTo(f Predicate, err string) *nextTo {
	return &nextTo{f, err}
}

func (a nextTo) Try(end bool, latest any, _ []any) (done, passed bool, err string) {
	// fails on end
	if end {
		return true, false, a.err
	} else if a.f(latest) {
		return true, true, ""
	} else {
		return true, a.f(latest), a.err + fmt.Sprintf("\nFailing message (of type %T): %+v", latest, latest)
	}
}

// The lastTo struct implements Asserter. Its Predicate function is called once, on end.
//
// The Predicate return value gives the test outcome (success/failure).
type lastTo predicateAndError

func newLastTo(f Predicate, err string) *lastTo {
	return &lastTo{f, err}
}

func (a lastTo) Try(end bool, latest any, _ []any) (done, passed bool, err string) {
	// fails on end
	if end {
		if latest != nil {
			if a.f(latest) {
				return true, true, ""
			} else {
				return true, false, a.err + fmt.Sprintf("\nFailing message (of type %T): %+v", latest, latest)
			}
		} else {
			return true, false, a.err + "\nReason: last message missing" // no "last" -> fails
		}
	}
	// unfinished
	return false, false, ""
}

// The allTo struct implements Asserter. Its Predicate function is called on message and on end.
//
// If the Predicate returns true, asserting is not done,
// If the Predicate returns false, asserting is done and fails,
// If the end is reached, asserting is done and succeeds.
type allTo predicateAndError

func newAllTo(p Predicate, err string) *allTo {
	return &allTo{p, err}
}

func (a allTo) Try(end bool, latest any, _ []any) (done, passed bool, err string) {
	if end {
		return true, true, ""
	} else {
		if a.f(latest) {
			return false, false, "" // ongoing
		} else {
			return true, false, a.err + fmt.Sprintf("\nFailing message (of type %T): %+v", latest, latest) // failed
		}
	}
}
