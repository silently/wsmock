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

		// script
		conn.Send("ping")

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

		// script
		conn.Send("ping")

		// assert
		rec.Assert().OneToBe(Message{"chat", "notfound"})
		rec.Assert().NextToBe(Message{"chat", "notfound"})
		rec.Assert().LastToBe(Message{"chat", "notfound"})
		rec.Assert().OneToContain("notfound")
		rec.Assert().ConnClosed()

		rec.RunAssertions(100 * time.Millisecond)
	})
}
