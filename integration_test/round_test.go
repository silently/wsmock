package integration_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestRound(t *testing.T) {
	t.Run("same message is caught in one round (not 2)", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("pong")
		}()

		// assert
		rec.NewAssertion().OneToBe("pong")
		rec.RunAssertions(5 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("OneToBe should succeed, mockT output is:\n", getTestOutput(mockT))
		}

		rec.NewAssertion().OneToBe("pong")
		rec.RunAssertions(5 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("OneToBe should fail for second Run", getTestOutput(mockT))
		}
	})
}
