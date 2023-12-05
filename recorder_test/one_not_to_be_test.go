package recorder_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestOneNotToBe(t *testing.T) {
	t.Run("succeeds when message is not received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.Send("ping")
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("pong")
		}()

		// assert
		rec.Assert().OneNotToBe("pongpong")
		rec.RunAssertions(20 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("OneNotToBe should succeed, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("fails when message is received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.Send("ping")
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("pong")
		}()

		// assert
		rec.Assert().OneNotToBe("pong")
		rec.RunAssertions(20 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("OneNotToBe should fail (message is received)")
		}
	})
}
