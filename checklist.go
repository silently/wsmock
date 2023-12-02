package wsmock

import (
	"encoding/json"
	"fmt"
	"strings"
)

type AssertNode interface {
	OneToBe(target any) AssertNode
	OneToCheck(Predicate) AssertNode
	OneToContain(sub string) AssertNode
	// OneToMatch(re regexp.Regexp) AssertNode
	// OneNotToBe(target any) AssertNode
	// OneNotToCheck(Predicate) AssertNode
	// OneNotToContain(sub string) AssertNode
	// OneNotToMatch(re regexp.Regexp) AssertNode
	FirstToBe(target any) AssertNode
	LastToBe(target any) AssertNode
	ConnClosed() AssertNode
}

type Checklist struct {
	r    *Recorder
	list []Asserter
}

// Generic API

// Adds custom Asserter
func (c *Checklist) AddAsserter(a Asserter) *Checklist {
	c.list = append(c.list, a)
	return c
}

// Adds custom AsserterFunc
func (c *Checklist) Check(a AsserterFunc) *Checklist {
	c.list = append(c.list, a)
	return c
}

func (c *Checklist) AllToCheck(a AssertOnEnd) *Checklist {
	c.list = append(c.list, a)
	return c
}

// One*

func (c *Checklist) OneToBe(target any) *Checklist {
	c.list = append(c.list, NewFailOnEnd(func(msg any) bool {
		return msg == target
	}, fmt.Sprintf("message not received\nexpected: %+v", target)))
	return c
}

func (c *Checklist) OneToCheck(f FailOnEnd) *Checklist {
	c.list = append(c.list, f)
	return c
}

func (c *Checklist) OneToContain(sub string) *Checklist {
	c.list = append(c.list, NewFailOnEnd(
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
	return c
}

// First*

func (c *Checklist) FirstToBe(target any) *Checklist {
	return c.Check(func(_ bool, _ any, all []any) (done, passed bool, errorMessage string) {
		done = true
		hasReceivedOne := len(all) > 0
		passed = hasReceivedOne && all[0] == target
		if passed {
			return
		}
		if hasReceivedOne {
			errorMessage = fmt.Sprintf("incorrect first message\nexpected: %+v\nreceived: %+v", target, all[0])
		} else {
			errorMessage = fmt.Sprintf("incorrect first message\nexpected: %+v\nreceived none", target)
		}
		return
	})
}

// Last*

// Asserts last message (always times out)
func (c *Checklist) LastToBe(target any) *Checklist {
	return c.AllToCheck(*NewAssertOnEnd(func(all []any) bool {
		length := len(all)
		return length > 0 && all[length-1] == target
	}, fmt.Sprintf("incorrect last message on timeout\nexpected: %+v", target)))
}

// OneNot*
// Asserts if a message has not been received by recorder (can fail before time out)
func (c *Checklist) OneNotToBe(target any) *Checklist {
	return c.Check(func(end bool, latest any, _ []any) (done, passed bool, errorMessage string) {
		if end {
			done = true
			passed = true
		} else if latest == target {
			done = true
			passed = false
			errorMessage = fmt.Sprintf("message should not be received\nunexpected: %+v", target)
		}
		return
	})
}

// Other

// Asserts if conn has been closed
func (c *Checklist) ConnClosed() *Checklist {
	return c.Check(func(end bool, latest any, all []any) (done, passed bool, errorMessage string) {
		if end {
			passed = c.r.done // conn closed => recorder done
			errorMessage = "conn should be closed"
		}
		return
	})
}
