package main

import (
	"strings"
	"testing"
	"time"

	"wsmock"
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

// if type Asserter
func HasReceivedAutoSplit(expected string) wsmock.Finder {
	return func(messages []any) bool {
		for _, m := range messages {
			if str, ok := m.(string); ok {
				for _, w := range strings.Split(str, "\n") {
					if w == expected {
						return true
					}
				}
			}
		}
		return false
	}
}

func TestRunClient(t *testing.T) {
	t.Run("two clients on same hub receive own and other messages (AssertReceivedContains)", func(t *testing.T) {
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
		// - alternatively: write a custom Finder by splitting messages around newlines (check client.go line 108)
		rec1.AssertReceivedContains("one")
		rec1.AssertReceivedContains("two")
		rec2.AssertReceivedContains("one")
		rec2.AssertReceivedContains("two")

		// run all previously declared assertions with a timeout
		wsmock.RunAssertions(t, 100*time.Millisecond)
	})

	t.Run("two clients on same hub receive own and other messages (custom finder)", func(t *testing.T) {
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
		// - alternatively: write a custom Finder by splitting messages around newlines (check client.go line 108)
		rec1.AssertOnWrite("", HasReceivedAutoSplit("one"))
		rec2.AssertOnWrite("", HasReceivedAutoSplit("one"))

		// run all previously declared assertions with a timeout
		wsmock.RunAssertions(t, 100*time.Millisecond)
	})

}
