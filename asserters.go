package wsmock

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

// A AllPredicate function maps its input to true or false.
type AllPredicate func(all []any) (passed bool)

// The oneTo struct implements Asserter. Its Predicate function is called each time a message is received.
//
// If the Predicate returns true, asserting succeeds,
// If the Predicate returns false, asserting is considered not done/solved and waits for a new message or until end,
// If the end is reached, the asserting fails and the err is displayed.
type oneTo struct {
	f   Predicate
	err string
}

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

// The nextTo struct implements Asserter. Its Predicate function is called once, when the next message
// is received or on timeout.
//
// The predicate return value gives the test outcome.
type nextTo struct {
	f   Predicate
	err string
}

func newNextTo(f Predicate, err string) *nextTo {
	return &nextTo{f, err}
}

func (a nextTo) Try(end bool, latest any, _ []any) (done, passed bool, err string) {
	// fails on end
	if end {
		return true, false, a.err
	}
	return true, a.f(latest), a.err
}

// The lastTo struct implements Asserter. Its Predicate function is called once, on timeout.
//
// The predicate return value gives the test outcome.
type lastTo struct {
	f   Predicate
	err string
}

func newLastTo(f Predicate, err string) *lastTo {
	return &lastTo{f, err}
}

func (a lastTo) Try(end bool, _ any, all []any) (done, passed bool, err string) {
	// fails on end
	if end {
		if length := len(all); length > 0 {
			last := all[length-1]
			return true, a.f(last), a.err // there's a last to be tested
		}
		return true, false, a.err // no "last" -> fails
	}
	// unfinished
	return false, false, ""
}

// An AssertOnEnd function is only called when Assert ends (timeout or connection close).
//
// It gets in order all the messages received during the run, and returns the assertion outcome.
type allTo struct {
	f   AllPredicate
	err string
}

func newAllTo(p AllPredicate, err string) *allTo {
	return &allTo{p, err}
}

func (a allTo) Try(end bool, _ any, all []any) (done, passed bool, err string) {
	// bypassed until end
	if !end {
		return false, false, ""
	}
	return true, a.f(all), a.err
}
