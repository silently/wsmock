package wsmock

import (
	"encoding/json"
	"regexp"
	"strings"
)

func not(f Predicate) Predicate {
	return func(msg any) bool {
		return !f(msg)
	}
}

func eq(target any) Predicate {
	return func(msg any) bool {
		return msg == target
	}
}

func contains(sub string) Predicate {
	return func(msg any) bool {
		if str, ok := msg.(string); ok {
			return strings.Contains(str, sub)
		} else {
			b, err := json.Marshal(msg)
			if err == nil {
				return strings.Contains(string(b), sub)
			}
		}
		return false
	}
}

func matches(re *regexp.Regexp) Predicate {
	return func(msg any) bool {
		if str, ok := msg.(string); ok {
			return re.MatchString(str)
		} else {
			b, err := json.Marshal(msg)
			if err == nil {
				return re.Match(b)
			}
		}
		return false
	}
}
