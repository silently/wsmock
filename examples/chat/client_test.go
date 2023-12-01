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

// Custom Asserter that splits received writes into several messages if separated by "\n"
// and then test if one of them is target
func hasReceivedAutoSplit(target string) wsmock.AsserterFunc {
	return func(end bool, latest any, _ []any) (done, passed bool, errorMessage string) {
		if end {
			passed = false
			errorMessage = fmt.Sprintf("[wsmock] no message received containing: %v", target)
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
	t.Run("two clients on same hub receive own and other messages (ToReceiveContaining)", func(t *testing.T) {
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
		// - alternatively: check next test with a custom Asserter (see next test)
		rec1.ToReceiveContaining("one")
		rec1.ToReceiveContaining("two")
		rec2.ToReceiveContaining("one")
		rec2.ToReceiveContaining("two")

		// run all previously declared assertions with a timeout
		wsmock.RunAssertions(t, 250*time.Millisecond)
	})

	t.Run("two clients on same hub receive own and other messages (with custom Asserter)", func(t *testing.T) {
		hub := runNewHub()
		conn1, rec1 := wsmock.NewGorillaMockAndRecorder(t)
		conn2, rec2 := wsmock.NewGorillaMockAndRecorder(t)
		runClient(hub, conn1)
		runClient(hub, conn2)

		// script sends
		conn1.Send("one")
		conn2.Send("two")
		// use a custom Asserter that splits messages around newlines (check client.go line 108)
		rec1.AddAsserter(hasReceivedAutoSplit("one"))
		rec2.AddAsserter(hasReceivedAutoSplit("one"))

		// run all previously declared assertions with a timeout
		wsmock.RunAssertions(t, 100*time.Millisecond)
	})
}
