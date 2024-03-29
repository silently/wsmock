package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/silently/wsmock"
)

type Message struct {
	Kind    string `json:"kind"`
	Payload any    `json:"payload"`
}

func runNewHub() *Hub {
	hub := newHub()
	go hub.run()
	return hub
}

// Custom Condition that splits received writes into several messages if separated by "\n"
// and then test if one of them is target
func hasReceivedAutoSplit(target string) wsmock.ConditionFunc {
	return func(end bool, latest any, _ []any) (done, passed bool, err string) {
		if end {
			passed = false
			err = fmt.Sprintf("[wsmock] no message received containing: %v", target)
		} else if str, ok := latest.(string); ok {
			for _, w := range strings.Split(str, "\n") {
				if w == target {
					done = true
					passed = true
					break
				}
			}
		}
		return
	}
}

func TestRunClient(t *testing.T) {
	t.Run("two clients on same hub receive own and other messages (OneToContain)", func(t *testing.T) {
		hub := runNewHub()
		conn1, rec1 := wsmock.NewGorillaMockAndRecorder(t)
		conn2, rec2 := wsmock.NewGorillaMockAndRecorder(t)
		runClient(hub, conn1)
		runClient(hub, conn2)

		// script sends
		conn1.Send("one")
		conn2.Send("two")
		// - assertions: client may queue and bundle messages together, that's why we check
		//   if message is contained in a wider bundled message
		// - alternatively: check next test with a custom Condition (see next test)
		rec1.NewAssertion().OneToContain("one")
		rec1.NewAssertion().OneToContain("two")
		rec2.NewAssertion().OneToContain("one")
		rec2.NewAssertion().OneToContain("two")

		// run all previously declared assertions with a timeout
		wsmock.RunAssertions(t, 250*time.Millisecond)
	})

	t.Run("two clients on same hub receive own and other messages (with custom Condition)", func(t *testing.T) {
		hub := runNewHub()
		conn1, rec1 := wsmock.NewGorillaMockAndRecorder(t)
		conn2, rec2 := wsmock.NewGorillaMockAndRecorder(t)
		runClient(hub, conn1)
		runClient(hub, conn2)

		// script sends
		conn1.Send("one")
		conn2.Send("two")
		// use a custom Condition that splits messages around newlines (check client.go line 108)
		rec1.NewAssertion().With(hasReceivedAutoSplit("one"))
		rec2.NewAssertion().With(hasReceivedAutoSplit("one"))

		// run all previously declared assertions with a timeout
		wsmock.RunAssertions(t, 100*time.Millisecond)
	})
}
