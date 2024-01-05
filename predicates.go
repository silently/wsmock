package wsmock

import (
	"encoding/json"
	"regexp"
	"strings"
)

// A Predicate function maps its input to true or false.
type Predicate func(msg any) (passed bool)

type predicateAndError struct {
	f   Predicate
	err string
}

func eq(target any) Predicate {
	return func(msg any) bool {
		return msg == target
	}
}

func contain(sub string) Predicate {
	return func(msg any) bool {
		if str, ok := msg.(string); ok {
			return strings.Contains(str, sub)
		} else if b, err := json.Marshal(msg); err == nil {
			return strings.Contains(string(b), sub)
		}
		return false
	}
}

func match(re *regexp.Regexp) Predicate {
	return func(msg any) bool {
		if str, ok := msg.(string); ok {
			return re.MatchString(str)
		} else if b, err := json.Marshal(msg); err == nil {
			return re.Match(b)
		}
		return false
	}
}

func not(f Predicate) Predicate {
	return func(msg any) bool {
		return !f(msg)
	}
}
