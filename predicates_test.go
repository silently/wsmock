package wsmock

import (
	"regexp"
	"testing"
)

type Message struct {
	Kind    string `json:"kind"`
	Payload string `json:"payload"`
}

var digitsRE *regexp.Regexp

func init() {
	digitsRE, _ = regexp.Compile("[0-9]+")
}

func TestPredicateEq(t *testing.T) {
	t.Run("eq creates a Predicate that checks equality", func(t *testing.T) {
		equalTo3 := eq(3)

		if !equalTo3(3) {
			t.Error("equalTo3: expected true but got false")
		}
		if equalTo3(4) {
			t.Error("equalTo3: expected false but got true")
		}
	})

	t.Run("not(eq) creates a Predicate that checks non equality", func(t *testing.T) {
		notEqualTo3 := not(eq(3))

		if !notEqualTo3(4) {
			t.Error("notEqualTo3: expected true but got false")
		}
		if notEqualTo3(3) {
			t.Error("notEqualTo3: expected false but got true")
		}
	})
}

func TestPredicateContains(t *testing.T) {
	t.Run("contain creates a Predicate that checks string containing substring", func(t *testing.T) {
		containWord := contain("word")

		if !containWord("say the word") {
			t.Error("containWord: expected true but got false")
		}
		if containWord("mot") {
			t.Error("containWord: expected false but got true")
		}
	})

	t.Run("not(contain) creates a Predicate that checks non containing", func(t *testing.T) {
		notContainWord := not(contain("word"))

		if !notContainWord("mot") {
			t.Error("notContainWord: expected true but got false")
		}
		if notContainWord("word") {
			t.Error("notContainWord: expected false but got true")
		}
	})

	t.Run("contain creates a Predicate that checks JSON message containing substring", func(t *testing.T) {
		containWord := contain("word")

		if !containWord(Message{"kind", "word"}) {
			t.Error("containWord: expected true but got false")
		}
		if containWord(Message{"kind", "letter"}) {
			t.Error("containWord: expected false but got true")
		}
	})
}

func TestPredicateMatches(t *testing.T) {
	t.Run("match creates a Predicate that checks string matching regex", func(t *testing.T) {
		matchDigits := match(digitsRE)

		if !matchDigits("0123") {
			t.Error("matchDigits: expected true but got false")
		}
		if matchDigits("abc") {
			t.Error("matchDigits: expected false but got true")
		}
	})

	t.Run("not(match) creates a Predicate that checks non matching", func(t *testing.T) {
		notMatchDigits := not(match(digitsRE))

		if !notMatchDigits("abc") {
			t.Error("notMatchDigits: expected true but got false")
		}
		if notMatchDigits("0123") {
			t.Error("notMatchDigits: expected false but got true")
		}
	})
}
