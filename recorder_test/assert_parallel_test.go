package recorder_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestMulti_FailFast(t *testing.T) {
	t.Run("should fail fast", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		go func() {
			conn.Send("ping")
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong2")
			conn.WriteJSON("pong3")
			conn.WriteJSON("pong4")
		}()

		// assert
		rec.Assert().OneToBe("pong4")
		rec.Assert().OneNotToBe("pong1") // failing assertion
		before := time.Now()
		rec.RunAssertions(300 * time.Millisecond)
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("OneNotToBe should fail")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 50*time.Millisecond {
				t.Errorf("OneNotToBe should fail faster")
			}
		}
	})
}

func TestMulti_Assert(t *testing.T) {
	t.Run("should succeed without blocking", func(t *testing.T) {
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
			conn.WriteJSON("pong4")
		}()

		// script
		rec.Assert().OneToBe("pong1")
		rec.Assert().OneToBe("pong2")
		rec.Assert().OneToBe("pong3")
		rec.Assert().OneToBe("pong4")

		// no assertion!
		rec.RunAssertions(100 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("several OneToBe should not fail")
		}
	})
}
