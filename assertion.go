package wsmock

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime"
)

// Assertions are ordered chains of conditions.
//
// When doing `rec.NewAssertion().OneToBe(1).OneToBe(2).LastToBe(3)` the following happens:
// - an Assertion struct has been created with `rec.NewAssertion()`
// - `OneToBe` is chainable and is used twice to add a Condition to the Assertion
// - `LastToBe` is not chainable and adds a final condition to the assertion
// - since assertion are ordered, the previous will succeed with `1 2 3` but will fail with `2 1 3`
//
// When several Assertion structs are created on the same recorder, they are run independently from each other.
type Assertion struct {
	conditions []Condition
}

func getFunctionName(f Predicate) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

// more general signature than *With*
func (a *Assertion) append(c Condition) *Assertion {
	a.conditions = append(a.conditions, c)
	return a
}

// Generic API

// Adds a ConditionFunc to the assertion
func (a *Assertion) With(f ConditionFunc) *Assertion {
	return a.append(f)
}

// OneTo*

// Adds a condition that succeeds if a new message is equal to the given interface (according to the equality operator `==`)
func (a *Assertion) OneToBe(target any) *Assertion {
	return a.append(newOneTo(eq(target), fmt.Sprintf("[OneToBe] no message is equal to: %#v", target)))
}

// Adds a condition that succeeds if a new message checks the Predicate
func (a *Assertion) OneToCheck(f Predicate) *Assertion {
	return a.append(newOneTo(f, fmt.Sprintf("[OneToCheck] no message checks predicate: %v", getFunctionName(f))))
}

// Adds a condition that succeeds if a new message contains the given string (messages that can't be converted to strings are JSON-marshalled first)
func (a *Assertion) OneToContain(sub string) *Assertion {
	return a.append(newOneTo(contain(sub), fmt.Sprintf("[OneToContain] no message contains string: %v", sub)))
}

// Adds a condition that succeeds if a new message matches the regular expression
func (a *Assertion) OneToMatch(re *regexp.Regexp) *Assertion {
	return a.append(newOneTo(match(re), fmt.Sprintf("[OneToMatch] no message matches regexp: %v", re)))
}

// OneNot*

// Adds a condition that succeeds if a new message is not equal to the given interface (according to the equality operator `==`)
func (a *Assertion) OneNotToBe(target any) *Assertion {
	return a.append(newOneTo(not(eq(target)), fmt.Sprintf("[OneNotToBe] message unexpectedly equal to: %#v", target)))
}

// Adds a condition that succeeds if a new message does not check the Predicate
func (a *Assertion) OneNotToCheck(f Predicate) *Assertion {
	return a.append(newOneTo(not(f), fmt.Sprintf("[OneNotToCheck] message unexpectedly checks predicate: %v", getFunctionName(f))))
}

// Adds a condition that succeeds if a new message does not contain the given string (messages that can't be converted to strings are JSON-marshalled first)
func (a *Assertion) OneNotToContain(sub string) *Assertion {
	return a.append(newOneTo(not(contain(sub)), fmt.Sprintf("[OneNotToContain] message unexpectedly contains string: %v", sub)))
}

// Adds a condition that succeeds if a new message does not match the regular expression
func (a *Assertion) OneNotToMatch(re *regexp.Regexp) *Assertion {
	return a.append(newOneTo(not(match(re)), fmt.Sprintf("[OneNotToMatch] message unexpectedly matches regexp: %v", re)))
}

// NextTo*

// Adds a condition that succeeds if the next message is equal to the given interface (according to the equality operator `==`)
func (a *Assertion) NextToBe(target any) *Assertion {
	return a.append(newNextTo(eq(target), fmt.Sprintf("[NextToBe] next message is not equal to: %#v", target)))
}

// Adds a condition that succeeds if the next message checks the Predicate
func (a *Assertion) NextToCheck(f Predicate) *Assertion {
	return a.append(newNextTo(f, fmt.Sprintf("[NextToCheck] next message does not check predicate: %v", getFunctionName(f))))
}

// Adds a condition that succeeds if the next message contains the given string (messages that can't be converted to strings are JSON-marshalled first)
func (a *Assertion) NextToContain(sub string) *Assertion {
	return a.append(newNextTo(contain(sub), fmt.Sprintf("[NextToContain] next message does not contain string: %v", sub)))
}

// Adds a condition that succeeds if the next message matches the regular expression
func (a *Assertion) NextToMatch(re *regexp.Regexp) *Assertion {
	return a.append(newNextTo(match(re), fmt.Sprintf("[NextToMatch] next message does not match regexp: %v", re)))
}

// NextNot*

// Adds a condition that succeeds if the next message is not equal to the given interface (according to the equality operator `==`)
func (a *Assertion) NextNotToBe(target any) *Assertion {
	return a.append(newNextTo(not(eq(target)), fmt.Sprintf("[NextNotToBe] next message unexpectedly equal to: %#v", target)))
}

// Adds a condition that succeeds if the next message does not check the Predicate
func (a *Assertion) NextNotToCheck(f Predicate) *Assertion {
	return a.append(newNextTo(not(f), fmt.Sprintf("[NextNotToCheck] next message unexpectedly checks predicate: %v", getFunctionName(f))))
}

// Adds a condition that succeeds if the next message does not contain the given string (messages that can't be converted to strings are JSON-marshalled first)
func (a *Assertion) NextNotToContain(sub string) *Assertion {
	return a.append(newNextTo(not(contain(sub)), fmt.Sprintf("[NextNotToContain] next message unexpectedly contains string: %v", sub)))
}

// Adds a condition that succeeds if the next message does not match the regular expression
func (a *Assertion) NextNotToMatch(re *regexp.Regexp) *Assertion {
	return a.append(newNextTo(not(match(re)), fmt.Sprintf("[NextNotToMatch] next message unexpectedly matches regexp: %v", re)))
}

// Last*

// Adds a condition that succeeds if the last message is equal to the given interface (according to the equality operator `==`)
func (a *Assertion) LastToBe(target any) {
	a.append(newLastTo(eq(target), fmt.Sprintf("[LastToBe] last message is not equal to: %#v", target)))
}

// Adds a condition that succeeds if the last message checks the Predicate
func (a *Assertion) LastToCheck(f Predicate) {
	a.append(newLastTo(f, fmt.Sprintf("[LastToCheck] last message deos not check predicate: %v", getFunctionName(f))))
}

// Adds a condition that succeeds if the last message contains the given string (messages that can't be converted to strings are JSON-marshalled first)
func (a *Assertion) LastToContain(sub string) {
	a.append(newLastTo(contain(sub), fmt.Sprintf("[LastToContain] last message does not contain string: %v", sub)))
}

// Adds a condition that succeeds if the last message matches the regular expression
func (a *Assertion) LastToMatch(re *regexp.Regexp) {
	a.append(newLastTo(match(re), fmt.Sprintf("[LastToMatch] last message does not match regexp: %v", re)))
}

// LastNot*

// Adds a condition that succeeds if the last message is not equal to the given interface (according to the equality operator `==`)
func (a *Assertion) LastNotToBe(target any) {
	a.append(newLastTo(not(eq(target)), fmt.Sprintf("[LastNotToBe] last message unexpectedly equal to: %#v", target)))
}

// Adds a condition that succeeds if the last message does not check the Predicate
func (a *Assertion) LastNotToCheck(f Predicate) {
	a.append(newLastTo(not(f), fmt.Sprintf("[LastNotToCheck] last message unexpectedly checks predicate: %v", getFunctionName(f))))
}

// Adds a condition that succeeds if the last message does not contain the given string (messages that can't be converted to strings are JSON-marshalled first)
func (a *Assertion) LastNotToContain(sub string) {
	a.append(newLastTo(not(contain(sub)), fmt.Sprintf("[LastNotToContain] last message unexpectedly contains string: %v", sub)))
}

// Adds a condition that succeeds if the last message does not match the regular expression
func (a *Assertion) LastNotToMatch(re *regexp.Regexp) {
	a.append(newLastTo(not(match(re)), fmt.Sprintf("[LastNotToMatch] last message unexpectedly matches regexp: %v", re)))
}

// All*

// Adds a condition that succeeds if all remaining messages are equal to the given interface (according to the equality operator `==`)
func (a *Assertion) AllToBe(target any) *Assertion {
	return a.append(newAllTo(eq(target), fmt.Sprintf("[AllToBe] message is not equal to: %#v", target)))
}

// Adds a condition that succeeds if all remaining messages check the Predicate
func (a *Assertion) AllToCheck(f Predicate) *Assertion {
	return a.append(newAllTo(f, fmt.Sprintf("[AllToCheck] message does not check predicate: %v", getFunctionName(f))))
}

// Adds a condition that succeeds if all remaining messages contain the given string (messages that can't be converted to strings are JSON-marshalled first)
func (a *Assertion) AllToContain(sub string) *Assertion {
	return a.append(newAllTo(contain(sub), fmt.Sprintf("[AllToContain] message does not contain string: %v", sub)))
}

// Adds a condition that succeeds if all remaining messages match the regular expression
func (a *Assertion) AllToMatch(re *regexp.Regexp) *Assertion {
	return a.append(newAllTo(match(re), fmt.Sprintf("[AllToMatch] message does not match regexp: %v", re)))
}

// None*

// Adds a condition that succeeds if no remaining message is equal to the given interface (according to the equality operator `==`)
func (a *Assertion) NoneToBe(target any) {
	a.append(newAllTo(not(eq(target)), fmt.Sprintf("[NoneToBe] message unexpectedly equal to: %#v", target)))
}

// Adds a condition that succeeds if no remaining message checks the Predicate
func (a *Assertion) NoneToCheck(f Predicate) {
	a.append(newAllTo(not(f), fmt.Sprintf("[NoneToCheck] message unexpectedly checks predicate: %v", getFunctionName(f))))
}

// Adds a condition that succeeds if no remaining message contains the given string (messages that can't be converted to strings are JSON-marshalled first)
func (a *Assertion) NoneToContain(sub string) {
	a.append(newAllTo(not(contain(sub)), fmt.Sprintf("[NoneToContain] message unexpectedly contains string: %v", sub)))
}

// Adds a condition that succeeds if no remaining message matches the regular expression
func (a *Assertion) NoneToMatch(re *regexp.Regexp) {
	a.append(newAllTo(not(match(re)), fmt.Sprintf("[NoneToMatch] message unexpectedly matches regexp: %v", re)))
}
