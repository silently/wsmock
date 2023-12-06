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
	return ab.append(newOneTo(eq(target), fmt.Sprintf("message not received: %+v", target)))
}

// Adds asserter that may succeed on receiving message, and fails if it dit not happen on end
func (ab *AssertionBuilder) OneToCheck(f Predicate) *AssertionBuilder {
	return ab.append(newOneTo(f, fmt.Sprintf("no message checking predicate: %+v", f)))
}

// Asserts if a message received by recorder contains a given string.
// Messages that can't be converted to strings are JSON-marshalled
func (ab *AssertionBuilder) OneToContain(sub string) *AssertionBuilder {
	return ab.append(newOneTo(contains(sub), fmt.Sprintf("no message containing string: %v", sub)))
}

func (ab *AssertionBuilder) OneToMatch(re *regexp.Regexp) *AssertionBuilder {
	return ab.append(newOneTo(matches(re), fmt.Sprintf("no message matching regexp: %v", re)))
}

// OneNot*

// Asserts if...
func (ab *AssertionBuilder) OneNotToBe(target any) *AssertionBuilder {
	return ab.append(newOneTo(not(eq(target)), fmt.Sprintf("unexpected message received: %+v", target)))
}

func (ab *AssertionBuilder) OneNotToCheck(f Predicate) *AssertionBuilder {
	return ab.append(newOneTo(not(f), fmt.Sprintf("message unexpectedly checking predicate: %+v", f)))
}

func (ab *AssertionBuilder) OneNotToContain(sub string) *AssertionBuilder {
	return ab.append(newOneTo(not(contains(sub)), fmt.Sprintf("message unexpectedly containing string: %v", sub)))
}

func (ab *AssertionBuilder) OneNotToMatch(re *regexp.Regexp) *AssertionBuilder {
	return ab.append(newOneTo(not(matches(re)), fmt.Sprintf("message unexpectedly matching regexp: %v", re)))
}

// NextTo*

// Asserts first message (times out only if no message is received)
func (ab *AssertionBuilder) NextToBe(target any) *AssertionBuilder {
	return ab.append(newNextTo(eq(target), fmt.Sprintf("next message not received: %+v", target)))
}

// Adds asserter that may succeed on receiving message, and fails if it dit not happen on end
func (ab *AssertionBuilder) NextToCheck(f Predicate) *AssertionBuilder {
	return ab.append(newNextTo(f, fmt.Sprintf("next message not checking predicate: %+v", f)))
}

// Asserts if a message received by recorder contains a given string.
// Messages that can't be converted to strings are JSON-marshalled
func (ab *AssertionBuilder) NextToContain(sub string) *AssertionBuilder {
	return ab.append(newNextTo(contains(sub), fmt.Sprintf("next message not containing string: %v", sub)))
}

func (ab *AssertionBuilder) NextToMatch(re *regexp.Regexp) *AssertionBuilder {
	return ab.append(newNextTo(matches(re), fmt.Sprintf("next message not matching regexp: %v", re)))
}

// NextNot*

// Asserts if...
func (ab *AssertionBuilder) NextNotToBe(target any) *AssertionBuilder {
	return ab.append(newNextTo(not(eq(target)), fmt.Sprintf("unexpected next message received: %+v", target)))
}

func (ab *AssertionBuilder) NextNotToCheck(f Predicate) *AssertionBuilder {
	return ab.append(newNextTo(not(f), fmt.Sprintf("next message unexpectedly checking predicate: %+v", f)))
}

func (ab *AssertionBuilder) NextNotToContain(sub string) *AssertionBuilder {
	return ab.append(newNextTo(not(contains(sub)), fmt.Sprintf("next message unexpectedly containing string: %v", sub)))
}

func (ab *AssertionBuilder) NextNotToMatch(re *regexp.Regexp) *AssertionBuilder {
	return ab.append(newNextTo(not(matches(re)), fmt.Sprintf("next message unexpectedly matching regexp: %v", re)))
}

// Last*

func (ab *AssertionBuilder) LastToBe(target any) {
	ab.append(newLastTo(eq(target), fmt.Sprintf("unexpected last message, wanted: %+v", target)))
}

func (ab *AssertionBuilder) LastToCheck(f Predicate) {
	ab.append(newLastTo(f, fmt.Sprintf("last message not checking predicate: %+v", f)))
}

func (ab *AssertionBuilder) LastToContain(sub string) {
	ab.append(newLastTo(contains(sub), fmt.Sprintf("last message not containing string: %v", sub)))
}

func (ab *AssertionBuilder) LastToMatch(re *regexp.Regexp) {
	ab.append(newLastTo(matches(re), fmt.Sprintf("last message not matching regexp: %v", re)))
}

// LastNot*

func (ab *AssertionBuilder) LastNotToBe(target any) {
	ab.append(newLastTo(not(eq(target)), fmt.Sprintf("unexpected message received: %+v", target)))
}

func (ab *AssertionBuilder) LastNotToCheck(f Predicate) {
	ab.append(newLastTo(not(f), fmt.Sprintf("last message unexpectedly checking predicate: %+v", f)))
}

func (ab *AssertionBuilder) LastNotToContain(sub string) {
	ab.append(newLastTo(not(contains(sub)), fmt.Sprintf("last message unexpectedly containing string: %v", sub)))
}

func (ab *AssertionBuilder) LastNotToMatch(re *regexp.Regexp) {
	ab.append(newLastTo(not(matches(re)), fmt.Sprintf("last message unexpectedly matching regexp: %v", re)))
}

// None*

// Asserts if a message has not been received by recorder (can fail before time out)
func (ab *AssertionBuilder) NoneToBe(target any) {
	ab.With(func(end bool, latest any, _ []any) (done, passed bool, err string) {
		if end {
			done = true
			passed = true
		} else if latest == target {
			done = true
			passed = false
			err = fmt.Sprintf("message should not be received\nunexpected: %+v", target)
		}
		return
	})
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
