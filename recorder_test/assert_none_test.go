package recorder_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestNoRunAssertionsion(t *testing.T) {
	t.Run("no assertion should succeed", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// dumb script
		go conn.Send("ping")

		// no assertion!
		rec.RunAssertions(10 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("NoAssertion can't fail")
		}
	})
}

// this test should be skipped, it's only there to inspect wsmock failing output
func TestFailing(t *testing.T) {
	t.Run("should fail", func(t *testing.T) {
		t.Skip()
		conn, rec := ws.NewGorillaMockAndRecorder(t)

		go conn.WriteJSON("no")

		// assert
		rec.Assert().OneNotToBe(Message{"chat", "sentence1"})
		rec.Assert().NextToBe(Message{"chat", "notfound"})
		rec.Assert().OneToCheck(stringLongerThan3)
		rec.RunAssertions(100 * time.Millisecond)
	})
}
