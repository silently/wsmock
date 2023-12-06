package wsmock

import (
	"fmt"
	"regexp"
)

type AssertionBuilder struct {
	rec  *Recorder
	list []Asserter
}

// more general signature than *With*
func (ab *AssertionBuilder) append(a Asserter) *AssertionBuilder {
	ab.list = append(ab.list, a)
	return ab
}

// Generic API

// Adds custom AsserterFunc
func (ab *AssertionBuilder) With(a AsserterFunc) *AssertionBuilder {
	return ab.append(a)
}

// OneTo*

// Asserts if a message has been received by recorder
func (ab *AssertionBuilder) OneToBe(target any) *AssertionBuilder {
	return ab.append(newOneTo(eq(target), fmt.Sprintf("[OneToBe] no message is equal to: %+v", target)))
}

// Adds asserter that may succeed on receiving message, and fails if it dit not happen on end
func (ab *AssertionBuilder) OneToCheck(f Predicate) *AssertionBuilder {
	return ab.append(newOneTo(f, fmt.Sprintf("[OneToCheck] no message checks predicate: %+v", f)))
}

// Asserts if a message received by recorder contains a given string.
// Messages that can't be converted to strings are JSON-marshalled
func (ab *AssertionBuilder) OneToContain(sub string) *AssertionBuilder {
	return ab.append(newOneTo(contains(sub), fmt.Sprintf("[OneToContain] no message contains string: %v", sub)))
}

func (ab *AssertionBuilder) OneToMatch(re *regexp.Regexp) *AssertionBuilder {
	return ab.append(newOneTo(matches(re), fmt.Sprintf("[OneToMatch] no message matches regexp: %v", re)))
}

// OneNot*

// Asserts if...
func (ab *AssertionBuilder) OneNotToBe(target any) *AssertionBuilder {
	return ab.append(newOneTo(not(eq(target)), fmt.Sprintf("[OneNotToBe] message unexpectedly equal to: %+v", target)))
}

func (ab *AssertionBuilder) OneNotToCheck(f Predicate) *AssertionBuilder {
	return ab.append(newOneTo(not(f), fmt.Sprintf("[OneNotToCheck] message unexpectedly checks predicate: %+v", f)))
}

func (ab *AssertionBuilder) OneNotToContain(sub string) *AssertionBuilder {
	return ab.append(newOneTo(not(contains(sub)), fmt.Sprintf("[OneNotToContain] message unexpectedly contains string: %v", sub)))
}

func (ab *AssertionBuilder) OneNotToMatch(re *regexp.Regexp) *AssertionBuilder {
	return ab.append(newOneTo(not(matches(re)), fmt.Sprintf("[OneNotToMatch] message unexpectedly matches regexp: %v", re)))
}

// NextTo*

// Asserts first message (times out only if no message is received)
func (ab *AssertionBuilder) NextToBe(target any) *AssertionBuilder {
	return ab.append(newNextTo(eq(target), fmt.Sprintf("[NextToBe] next message is not equal to: %+v", target)))
}

// Adds asserter that may succeed on receiving message, and fails if it dit not happen on end
func (ab *AssertionBuilder) NextToCheck(f Predicate) *AssertionBuilder {
	return ab.append(newNextTo(f, fmt.Sprintf("[NextToCheck] next message does not check predicate: %+v", f)))
}

// Asserts if a message received by recorder contains a given string.
// Messages that can't be converted to strings are JSON-marshalled
func (ab *AssertionBuilder) NextToContain(sub string) *AssertionBuilder {
	return ab.append(newNextTo(contains(sub), fmt.Sprintf("[NextToContain] next message does not contain string: %v", sub)))
}

func (ab *AssertionBuilder) NextToMatch(re *regexp.Regexp) *AssertionBuilder {
	return ab.append(newNextTo(matches(re), fmt.Sprintf("[NextToMatch] next message does not match regexp: %v", re)))
}

// NextNot*

// Asserts if...
func (ab *AssertionBuilder) NextNotToBe(target any) *AssertionBuilder {
	return ab.append(newNextTo(not(eq(target)), fmt.Sprintf("[NextNotToBe] next message unexpectedly equal to: %+v", target)))
}

func (ab *AssertionBuilder) NextNotToCheck(f Predicate) *AssertionBuilder {
	return ab.append(newNextTo(not(f), fmt.Sprintf("[NextNotToCheck] next message unexpectedly checks predicate: %+v", f)))
}

func (ab *AssertionBuilder) NextNotToContain(sub string) *AssertionBuilder {
	return ab.append(newNextTo(not(contains(sub)), fmt.Sprintf("[NextNotToContain] next message unexpectedly contains string: %v", sub)))
}

func (ab *AssertionBuilder) NextNotToMatch(re *regexp.Regexp) *AssertionBuilder {
	return ab.append(newNextTo(not(matches(re)), fmt.Sprintf("[NextNotToMatch] next message unexpectedly matches regexp: %v", re)))
}

// Last*

func (ab *AssertionBuilder) LastToBe(target any) {
	ab.append(newLastTo(eq(target), fmt.Sprintf("[LastToBe] last message is not equal to: %+v", target)))
}

func (ab *AssertionBuilder) LastToCheck(f Predicate) {
	ab.append(newLastTo(f, fmt.Sprintf("[LastToCheck] last message deos not check predicate: %+v", f)))
}

func (ab *AssertionBuilder) LastToContain(sub string) {
	ab.append(newLastTo(contains(sub), fmt.Sprintf("[LastToContain] last message does not contain string: %v", sub)))
}

func (ab *AssertionBuilder) LastToMatch(re *regexp.Regexp) {
	ab.append(newLastTo(matches(re), fmt.Sprintf("[LastToMatch] last message does not match regexp: %v", re)))
}

// LastNot*

func (ab *AssertionBuilder) LastNotToBe(target any) {
	ab.append(newLastTo(not(eq(target)), fmt.Sprintf("[LastNotToBe] last message unexpectedly equal to: %+v", target)))
}

func (ab *AssertionBuilder) LastNotToCheck(f Predicate) {
	ab.append(newLastTo(not(f), fmt.Sprintf("[LastNotToCheck] last message unexpectedly checks predicate: %+v", f)))
}

func (ab *AssertionBuilder) LastNotToContain(sub string) {
	ab.append(newLastTo(not(contains(sub)), fmt.Sprintf("[LastNotToContain] last message unexpectedly contains string: %v", sub)))
}

func (ab *AssertionBuilder) LastNotToMatch(re *regexp.Regexp) {
	ab.append(newLastTo(not(matches(re)), fmt.Sprintf("[LastNotToMatch] last message unexpectedly matches regexp: %v", re)))
}

// All*

func (ab *AssertionBuilder) AllToBe(target any) *AssertionBuilder {
	return ab.append(newAllTo(eq(target), fmt.Sprintf("[AllToBe] message is not equal to: %+v", target)))
}

func (ab *AssertionBuilder) AllToCheck(f Predicate) *AssertionBuilder {
	return ab.append(newAllTo(f, fmt.Sprintf("[AllToCheck] message does not check predicate: %+v", f)))
}

func (ab *AssertionBuilder) AllToContain(sub string) *AssertionBuilder {
	return ab.append(newAllTo(contains(sub), fmt.Sprintf("[AllToContain] message does not contain string: %v", sub)))
}

func (ab *AssertionBuilder) AllToMatch(re *regexp.Regexp) *AssertionBuilder {
	return ab.append(newAllTo(matches(re), fmt.Sprintf("[AllToMatch] message does not match regexp: %v", re)))
}

// None*

// Asserts if a message has not been received by recorder (can fail before time out)
func (ab *AssertionBuilder) NoneToBe(target any) {
	ab.append(newAllTo(not(eq(target)), fmt.Sprintf("[NoneToBe] message unexpectedly equal to: %+v", target)))
}

func (ab *AssertionBuilder) NoneToCheck(f Predicate) {
	ab.append(newAllTo(not(f), fmt.Sprintf("[NoneToCheck] message unexpectedly checks predicate: %+v", f)))
}

func (ab *AssertionBuilder) NoneToContain(sub string) {
	ab.append(newAllTo(not(contains(sub)), fmt.Sprintf("[NoneToContain] message unexpectedly contains string: %v", sub)))
}

func (ab *AssertionBuilder) NoneToMatch(re *regexp.Regexp) {
	ab.append(newAllTo(not(matches(re)), fmt.Sprintf("[NoneToMatch] message unexpectedly matches regexp: %v", re)))
}

// Other

// Asserts if conn has been closed
func (ab *AssertionBuilder) ConnClosed() {
	ab.With(func(end bool, latest any, all []any) (done, passed bool, err string) {
		if end {
			passed = ab.rec.done // conn closed => recorder done
			err = "conn should be closed"
		}
		return
	})
}
