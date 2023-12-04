package wsmock

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type AssertionBuilder struct {
	r    *Recorder
	list []Asserter
}

// private

func (ab *AssertionBuilder) withAllOnEnd(a AssertOnEnd) *AssertionBuilder {
	ab.list = append(ab.list, a)
	return ab
}

// Generic API

// Adds custom AsserterFunc
func (ab *AssertionBuilder) With(a AsserterFunc) *AssertionBuilder {
	ab.list = append(ab.list, a)
	return ab
}

// One*

// Asserts if a message has been received by recorder
func (ab *AssertionBuilder) OneToBe(target any) *AssertionBuilder {
	ab.list = append(ab.list, NewFailOnEnd(func(msg any) bool {
		return msg == target
	}, fmt.Sprintf("message not received\nexpected: %+v", target)))
	return ab
}

// Adds asserter that may succeed on receiving message, and fails if it dit not happen on end
func (ab *AssertionBuilder) OneToCheck(f FailOnEnd) *AssertionBuilder {
	ab.list = append(ab.list, f)
	return ab
}

// Asserts if a message received by recorder contains a given string.
// Messages that can't be converted to strings are JSON-marshalled
func (ab *AssertionBuilder) OneToContain(sub string) *AssertionBuilder {
	ab.list = append(ab.list, NewFailOnEnd(
		func(msg any) bool {
			if str, ok := msg.(string); ok {
				return strings.Contains(str, sub)
			} else {
				b, err := json.Marshal(msg)
				if err == nil {
					return strings.Contains(string(b), sub)
				}
			}
			return false
		}, fmt.Sprintf("no message containing string\nexpected: %v", sub)))
	return ab
}

func (ab *AssertionBuilder) OneToMatch(re regexp.Regexp) *AssertionBuilder {
	ab.list = append(ab.list, NewFailOnEnd(
		func(msg any) bool {
			if str, ok := msg.(string); ok {
				return re.Match([]byte(str))
			} else {
				b, err := json.Marshal(msg)
				if err == nil {
					return re.Match(b)
				}
			}
			return false
		}, fmt.Sprintf("no message matching regexp\nexpected: %v", re)))
	return ab
}

// First*

// Asserts first message (times out only if no message is received)
func (ab *AssertionBuilder) FirstToBe(target any) *AssertionBuilder {
	return ab.With(func(_ bool, _ any, all []any) (done, passed bool, err string) {
		done = true
		hasReceivedOne := len(all) > 0
		passed = hasReceivedOne && all[0] == target
		if passed {
			return
		}
		if hasReceivedOne {
			err = fmt.Sprintf("incorrect first message\nexpected: %+v\nreceived: %+v", target, all[0])
		} else {
			err = fmt.Sprintf("incorrect first message\nexpected: %+v\nreceived none", target)
		}
		return
	})
}

// Last*

// Asserts last message (always times out)
func (ab *AssertionBuilder) LastToBe(target any) *AssertionBuilder {
	return ab.withAllOnEnd(*NewAssertOnEnd(func(all []any) bool {
		length := len(all)
		return length > 0 && all[length-1] == target
	}, fmt.Sprintf("incorrect last message on timeout\nexpected: %+v", target)))
}

// OneNot*

// Asserts if a message has not been received by recorder (can fail before time out)
func (ab *AssertionBuilder) OneNotToBe(target any) *AssertionBuilder {
	return ab.With(func(end bool, latest any, _ []any) (done, passed bool, err string) {
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
func (ab *AssertionBuilder) ConnClosed() *AssertionBuilder {
	return ab.With(func(end bool, latest any, all []any) (done, passed bool, err string) {
		if end {
			passed = ab.r.done // conn closed => recorder done
			err = "conn should be closed"
		}
		return
	})
}
