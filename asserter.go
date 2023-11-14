package wsmock

// This type could be made public and we could enable other type of asserters (than AsserterFunc)
// to be added on Recorders
type asserter interface {
	assert(end bool, latestWrite any, allWrites []any) (done, passed bool, errorMessage string)
}

// AsserterFunc functions can be added to Recorders and are called possibly several times
// during the same Run to determine the outcome of the test.
//
// Case 1: when a write occurs (on the underlying Conn, and thus on the associated Recorder),
// AsserterFunc is called with (false, latestWrite, allWritesIncludingLatest)
//
// Possible outcomes of Case 1 are:
// - if a decision can't be made about the assertion being true or not (e.g. if more data or timeout needed)
// *done* is false and the other return values don't matter
// - if the test succeeds, *done* and *passed* are true, *errorMessage* does not matter
// - if test fails, *done* is true, *passed* is false and *errorMessage* will be used to print an error
//
// Case 2: when timeout is reached (the underlying Conn is closed or the Run times out),
// AsserterFunc is called one last time with (true, latestWrite, allWritesIncludingLatest). Contrary to Case 1,
// we don't know if there was indeed a *latestWrite* (it could be nil, like *allWritesIncludingLatest*).
//
// For case 2, the return value `done` is considered true whatever is returned, while `passed` and `errorMessage`
// do give the test outcome.
type AsserterFunc func(end bool, latestWrite any, allWrites []any) (done, passed bool, errorMessage string)

func (f AsserterFunc) assert(end bool, latestWrite any, allWrites []any) (done, passed bool, errorMessage string) {
	return f(end, latestWrite, allWrites)
}
