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

// An AsserterOnReceive function is called each time a message is received.
//
// If it returns true, the assertion succeeds,
// If it returns false, the assertion is considered not done/solved and waits for a new message or until end,
// If the end is reached, the assertion fails.
type AsserterOnReceive func(received any) (passed bool)

func (f AsserterOnReceive) Try(end bool, received any, allReceived []any) (done, passed bool, errorMessage string) {
	// fails on end
	if end {
		return true, false, ""
	}
	if f(received) { // succeeds
		return true, true, ""
	}
	// unfinished
	return false, false, ""
}

// An AsserterOnEnd function is only called when RunAssertions ends (timeout or connection close).
//
// It gets in order all the messages received during the run, and returns the assertion outcome. 
type AsserterOnEnd func(all []any) (passed bool)

func (f AsserterOnEnd) Try(end bool, _ any, all []any) (done, passed bool, errorMessage string) {
	// bypassed until end
	if !end {
		return false, false, ""
	}
	return true, f(all), ""
}