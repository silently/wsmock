package recorder_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestNoneToBe(t *testing.T) {
	t.Run("succeeds when message is not received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.Send("ping")
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong2")
			conn.WriteJSON("pong3")
		}()

		// assert
		rec.Assert().NoneToBe("pongpong")
		rec.RunAssertions(20 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("OneNotToBe should succeed, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("fails fast when message is received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.Send("ping")
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong2")
			conn.WriteJSON("pong3")
		}()

		// assert
		rec.Assert().NoneToBe("pong3")
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("NoneToBe should fail (message is received)")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 30*time.Millisecond {
				t.Errorf("NoneToBe should fail faster")
			}
		}
	})
}
