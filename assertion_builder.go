package wsmock

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime"
)

// Type used to represent assertions as an ordered chain of conditions.
//
// When doing `rec.Assert().OneToBe(1).OneToBe(2).LastToBe(3)` the following happens:
// - an AssertionBuilder struct has been created with `rec.Assert()`
// - `OneToBe` is chainable and is used twice to add a condition (Asserter) to the assertion
// - `LastToBe` is not chainable and adds a final condition to the assertion
// - since assertion are ordered, the previous will succeed with `1 2 3` but will fail with `2 1 3`
//
// When several AssertionBuilder structs are created on the same recorder, they are run independently from each other.
type AssertionBuilder struct {
	conditions []Asserter
}

func getFunctionName(f Predicate) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

// more general signature than *With*
func (ab *AssertionBuilder) append(a Asserter) *AssertionBuilder {
	ab.conditions = append(ab.conditions, a)
	return ab
}

// Generic API

// Asserts with a custom AsserterFunc
func (ab *AssertionBuilder) With(a AsserterFunc) *AssertionBuilder {
	return ab.append(a)
}

// OneTo*

// Succeeds if a new message is equal to the given interface (according to the equality operator `==`)
func (ab *AssertionBuilder) OneToBe(target any) *AssertionBuilder {
	return ab.append(newOneTo(eq(target), fmt.Sprintf("[OneToBe] no message is equal to: %+v", target)))
}

// Succeeds if a new message checks the Predicate
func (ab *AssertionBuilder) OneToCheck(f Predicate) *AssertionBuilder {
	return ab.append(newOneTo(f, fmt.Sprintf("[OneToCheck] no message checks predicate: %v", getFunctionName(f))))
}

// Succeeds if a new message contains the given string (messages that can't be converted to strings are JSON-marshalled first)
func (ab *AssertionBuilder) OneToContain(sub string) *AssertionBuilder {
	return ab.append(newOneTo(contains(sub), fmt.Sprintf("[OneToContain] no message contains string: %v", sub)))
}

// Succeeds if a new message matches the regular expression
func (ab *AssertionBuilder) OneToMatch(re *regexp.Regexp) *AssertionBuilder {
	return ab.append(newOneTo(matches(re), fmt.Sprintf("[OneToMatch] no message matches regexp: %v", re)))
}

// OneNot*

// Succeeds if a new message is not equal to the given interface (according to the equality operator `==`)
func (ab *AssertionBuilder) OneNotToBe(target any) *AssertionBuilder {
	return ab.append(newOneTo(not(eq(target)), fmt.Sprintf("[OneNotToBe] message unexpectedly equal to: %+v", target)))
}

// Succeeds if a new message does not check the Predicate
func (ab *AssertionBuilder) OneNotToCheck(f Predicate) *AssertionBuilder {
	return ab.append(newOneTo(not(f), fmt.Sprintf("[OneNotToCheck] message unexpectedly checks predicate: %v", getFunctionName(f))))
}

// Succeeds if a new message does not contain the given string (messages that can't be converted to strings are JSON-marshalled first)
func (ab *AssertionBuilder) OneNotToContain(sub string) *AssertionBuilder {
	return ab.append(newOneTo(not(contains(sub)), fmt.Sprintf("[OneNotToContain] message unexpectedly contains string: %v", sub)))
}

// Succeeds if a new message does not match the regular expression
func (ab *AssertionBuilder) OneNotToMatch(re *regexp.Regexp) *AssertionBuilder {
	return ab.append(newOneTo(not(matches(re)), fmt.Sprintf("[OneNotToMatch] message unexpectedly matches regexp: %v", re)))
}

// NextTo*

// Succeeds if the next message is equal to the given interface (according to the equality operator `==`)
func (ab *AssertionBuilder) NextToBe(target any) *AssertionBuilder {
	return ab.append(newNextTo(eq(target), fmt.Sprintf("[NextToBe] next message is not equal to: %+v", target)))
}

// Succeeds if the next message checks the Predicate
func (ab *AssertionBuilder) NextToCheck(f Predicate) *AssertionBuilder {
	return ab.append(newNextTo(f, fmt.Sprintf("[NextToCheck] next message does not check predicate: %v", getFunctionName(f))))
}

// Succeeds if the next message contains the given string (messages that can't be converted to strings are JSON-marshalled first)
func (ab *AssertionBuilder) NextToContain(sub string) *AssertionBuilder {
	return ab.append(newNextTo(contains(sub), fmt.Sprintf("[NextToContain] next message does not contain string: %v", sub)))
}

// Succeeds if the next message matches the regular expression
func (ab *AssertionBuilder) NextToMatch(re *regexp.Regexp) *AssertionBuilder {
	return ab.append(newNextTo(matches(re), fmt.Sprintf("[NextToMatch] next message does not match regexp: %v", re)))
}

// NextNot*

// Succeeds if the next message is not equal to the given interface (according to the equality operator `==`)
func (ab *AssertionBuilder) NextNotToBe(target any) *AssertionBuilder {
	return ab.append(newNextTo(not(eq(target)), fmt.Sprintf("[NextNotToBe] next message unexpectedly equal to: %+v", target)))
}

// Succeeds if the next message does not check the Predicate
func (ab *AssertionBuilder) NextNotToCheck(f Predicate) *AssertionBuilder {
	return ab.append(newNextTo(not(f), fmt.Sprintf("[NextNotToCheck] next message unexpectedly checks predicate: %v", getFunctionName(f))))
}

// Succeeds if the next message does not contain the given string (messages that can't be converted to strings are JSON-marshalled first)
func (ab *AssertionBuilder) NextNotToContain(sub string) *AssertionBuilder {
	return ab.append(newNextTo(not(contains(sub)), fmt.Sprintf("[NextNotToContain] next message unexpectedly contains string: %v", sub)))
}

// Succeeds if the next message does not match the regular expression
func (ab *AssertionBuilder) NextNotToMatch(re *regexp.Regexp) *AssertionBuilder {
	return ab.append(newNextTo(not(matches(re)), fmt.Sprintf("[NextNotToMatch] next message unexpectedly matches regexp: %v", re)))
}

// Last*

// Succeeds if the last message is equal to the given interface (according to the equality operator `==`)
func (ab *AssertionBuilder) LastToBe(target any) {
	ab.append(newLastTo(eq(target), fmt.Sprintf("[LastToBe] last message is not equal to: %+v", target)))
}

// Succeeds if the last message checks the Predicate
func (ab *AssertionBuilder) LastToCheck(f Predicate) {
	ab.append(newLastTo(f, fmt.Sprintf("[LastToCheck] last message deos not check predicate: %v", getFunctionName(f))))
}

// Succeeds if the last message contains the given string (messages that can't be converted to strings are JSON-marshalled first)
func (ab *AssertionBuilder) LastToContain(sub string) {
	ab.append(newLastTo(contains(sub), fmt.Sprintf("[LastToContain] last message does not contain string: %v", sub)))
}

// Succeeds if the last message matches the regular expression
func (ab *AssertionBuilder) LastToMatch(re *regexp.Regexp) {
	ab.append(newLastTo(matches(re), fmt.Sprintf("[LastToMatch] last message does not match regexp: %v", re)))
}

// LastNot*

// Succeeds if the last message is not equal to the given interface (according to the equality operator `==`)
func (ab *AssertionBuilder) LastNotToBe(target any) {
	ab.append(newLastTo(not(eq(target)), fmt.Sprintf("[LastNotToBe] last message unexpectedly equal to: %+v", target)))
}

// Succeeds if the last message does not check the Predicate
func (ab *AssertionBuilder) LastNotToCheck(f Predicate) {
	ab.append(newLastTo(not(f), fmt.Sprintf("[LastNotToCheck] last message unexpectedly checks predicate: %v", getFunctionName(f))))
}

// Succeeds if the last message does not contain the given string (messages that can't be converted to strings are JSON-marshalled first)
func (ab *AssertionBuilder) LastNotToContain(sub string) {
	ab.append(newLastTo(not(contains(sub)), fmt.Sprintf("[LastNotToContain] last message unexpectedly contains string: %v", sub)))
}

// Succeeds if the last message does not match the regular expression
func (ab *AssertionBuilder) LastNotToMatch(re *regexp.Regexp) {
	ab.append(newLastTo(not(matches(re)), fmt.Sprintf("[LastNotToMatch] last message unexpectedly matches regexp: %v", re)))
}

// All*

// Succeeds if all remaining messages are equal to the given interface (according to the equality operator `==`)
func (ab *AssertionBuilder) AllToBe(target any) *AssertionBuilder {
	return ab.append(newAllTo(eq(target), fmt.Sprintf("[AllToBe] message is not equal to: %+v", target)))
}

// Succeeds if all remaining messages check the Predicate
func (ab *AssertionBuilder) AllToCheck(f Predicate) *AssertionBuilder {
	return ab.append(newAllTo(f, fmt.Sprintf("[AllToCheck] message does not check predicate: %v", getFunctionName(f))))
}

// Succeeds if all remaining messages contain the given string (messages that can't be converted to strings are JSON-marshalled first)
func (ab *AssertionBuilder) AllToContain(sub string) *AssertionBuilder {
	return ab.append(newAllTo(contains(sub), fmt.Sprintf("[AllToContain] message does not contain string: %v", sub)))
}

// Succeeds if all remaining messages match the regular expression
func (ab *AssertionBuilder) AllToMatch(re *regexp.Regexp) *AssertionBuilder {
	return ab.append(newAllTo(matches(re), fmt.Sprintf("[AllToMatch] message does not match regexp: %v", re)))
}

// None*

// Succeeds if no remaining message is equal to the given interface (according to the equality operator `==`)
func (ab *AssertionBuilder) NoneToBe(target any) {
	ab.append(newAllTo(not(eq(target)), fmt.Sprintf("[NoneToBe] message unexpectedly equal to: %+v", target)))
}

// Succeeds if no remaining message checks the Predicate
func (ab *AssertionBuilder) NoneToCheck(f Predicate) {
	ab.append(newAllTo(not(f), fmt.Sprintf("[NoneToCheck] message unexpectedly checks predicate: %v", getFunctionName(f))))
}

// Succeeds if no remaining message contains the given string (messages that can't be converted to strings are JSON-marshalled first)
func (ab *AssertionBuilder) NoneToContain(sub string) {
	ab.append(newAllTo(not(contains(sub)), fmt.Sprintf("[NoneToContain] message unexpectedly contains string: %v", sub)))
}

// Succeeds if no remaining message matches the regular expression
func (ab *AssertionBuilder) NoneToMatch(re *regexp.Regexp) {
	ab.append(newAllTo(not(matches(re)), fmt.Sprintf("[NoneToMatch] message unexpectedly matches regexp: %v", re)))
}
