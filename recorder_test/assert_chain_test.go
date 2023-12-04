package recorder_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestChainOneToBe(t *testing.T) {
	t.Run("succeeds when chain is an ordered subpart of messages", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)
		conn.WriteJSON(Message{"chat", "sentence1"})
		conn.WriteJSON(Message{"chat", "sentence2"})
		conn.WriteJSON(Message{"chat", "sentence3"})

		// declare expected chains
		rec.Assert().
			OneToBe(Message{"chat", "sentence1"}).
			OneToBe(Message{"chat", "sentence2"})
		rec.Assert().
			OneToBe(Message{"chat", "sentence1"}).
			OneToBe(Message{"chat", "sentence3"})
		rec.Assert().
			OneToBe(Message{"chat", "sentence2"}).
			OneToBe(Message{"chat", "sentence3"})

		rec.RunAssertions(300 * time.Millisecond) // it's a max

		if mockT.Failed() { // fail not expected
			t.Error("OneToBe*s* should succeed, mockT output is:", getTestOutput(mockT))
		}
	})
}
