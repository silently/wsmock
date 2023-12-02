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
// - if the test succeeds, *done* and *passed* are true, *errorMessage* does not matter
// - if test fails, *done* is true, *passed* is false and *errorMessage* is used to print an error
//
// Case 2: when end is reached (timeout or connection closed),
// Try is called one last time with (true, message, all). Contrary to Case 1,
// we don't know if there was indeed a *latest* (it could be nil, like *all*).
//
// For case 2, the return value `done` is considered true whatever is returned, while `passed` and `errorMessage`
// do give the test outcome.
type Asserter interface {
	Try(end bool, latest any, all []any) (done, passed bool, errorMessage string)
}

// The AsserterFunc type is an adapter to allow the use of an ordinary functions as an Asserter.
//
// Its signature and behaviour follows Asserters Try.
type AsserterFunc func(end bool, latest any, all []any) (done, passed bool, errorMessage string)

func (f AsserterFunc) Try(end bool, latest any, all []any) (done, passed bool, errorMessage string) {
	return f(end, latest, all)
}

// A Predicate function maps its input to true or false.
type Predicate func(msg any) (passed bool)

// The FailOnEnd struct implements Asserter. Its Predicate function is called each time a message is received.
//
// If the Predicate returns true, assertint succeeds,
// If the Predicate returns false, asserting is considered not done/solved and waits for a new message or until end,
// If the end is reached, the asserting fails and the errorMessage is displayed.
type FailOnEnd struct {
	p            Predicate
	errorMessage string
}

func NewFailOnEnd(p Predicate, errorMessage string) *FailOnEnd {
	return &FailOnEnd{p, errorMessage}
}

func (a FailOnEnd) Try(end bool, latest any, all []any) (done, passed bool, errorMessage string) {
	// fails on end
	if end {
		return true, false, a.errorMessage
	}
	if a.p(latest) { // succeeds
		return true, true, ""
	}
	// unfinished
	return false, false, ""
}

// A AllPredicate function maps its input to true or false.
type AllPredicate func(all []any) (passed bool)

// An AssertOnEnd function is only called when Assert ends (timeout or connection close).
//
// It gets in order all the messages received during the run, and returns the assertion outcome.
type AssertOnEnd struct {
	p            AllPredicate
	errorMessage string
}

func NewAssertOnEnd(p AllPredicate, errorMessage string) *AssertOnEnd {
	return &AssertOnEnd{p, errorMessage}
}

func (a AssertOnEnd) Try(end bool, _ any, all []any) (done, passed bool, errorMessage string) {
	// bypassed until end
	if !end {
		return false, false, ""
	}
	return true, a.p(all), ""
}
