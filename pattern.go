package wsmock

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type Pattern struct {
	r    *Recorder
	list []Asserter
}

// package private

func (p *Pattern) assertAllOnEnd(a AssertOnEnd) *Pattern {
	p.list = append(p.list, a)
	return p
}

// Generic API

// Adds custom AsserterFunc
func (p *Pattern) AddAssert(a AsserterFunc) *Pattern {
	p.list = append(p.list, a)
	return p
}

// One*

func (p *Pattern) OneToBe(target any) *Pattern {
	p.list = append(p.list, NewFailOnEnd(func(msg any) bool {
		return msg == target
	}, fmt.Sprintf("message not received\nexpected: %+v", target)))
	return p
}

func (p *Pattern) OneToCheck(f FailOnEnd) *Pattern {
	p.list = append(p.list, f)
	return p
}

func (p *Pattern) OneToContain(sub string) *Pattern {
	p.list = append(p.list, NewFailOnEnd(
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
	return p
}

func (p *Pattern) OneToMatch(re regexp.Regexp) *Pattern {
	p.list = append(p.list, NewFailOnEnd(
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
	return p
}

// First*

func (p *Pattern) FirstToBe(target any) *Pattern {
	return p.AddAssert(func(_ bool, _ any, all []any) (done, passed bool, err string) {
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
func (p *Pattern) LastToBe(target any) *Pattern {
	return p.assertAllOnEnd(*NewAssertOnEnd(func(all []any) bool {
		length := len(all)
		return length > 0 && all[length-1] == target
	}, fmt.Sprintf("incorrect last message on timeout\nexpected: %+v", target)))
}

// OneNot*
// Asserts if a message has not been received by recorder (can fail before time out)
func (p *Pattern) OneNotToBe(target any) *Pattern {
	return p.AddAssert(func(end bool, latest any, _ []any) (done, passed bool, err string) {
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
func (p *Pattern) ConnClosed() *Pattern {
	return p.AddAssert(func(end bool, latest any, all []any) (done, passed bool, err string) {
		if end {
			passed = p.r.done // conn closed => recorder done
			err = "conn should be closed"
		}
		return
	})
}
