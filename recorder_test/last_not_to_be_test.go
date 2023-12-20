package recorder_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestLastNotToBe_Success(t *testing.T) {
	t.Run("succeeds on end when last message is not equal", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong2")
			conn.WriteJSON("pong3")
			conn.WriteJSON("pong4")
		}()

		// assert
		rec.Assert().LastNotToBe("pong3")
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("LastNotToBe should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed < 30*time.Millisecond {
				t.Error("LastNotToBe should not succeed before timeout")
			}
		}
	})
}

func TestLastNotToBe_Failure(t *testing.T) {
	t.Run("fails when last message is equal", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong2")
			conn.WriteJSON("pong3")
			conn.WriteJSON("pong4")
		}()

		// assert
		rec.Assert().LastNotToBe("pong4")
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("LastNotToBe should fail")
		}
	})

	t.Run("fails when timeout occurs before last message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(30 * time.Millisecond)
			conn.WriteJSON("ping")
		}()

		// assert
		rec.Assert().LastNotToBe("pong")
		rec.RunAssertions(10 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("LastNotToBe should fail because of timeout")
		}
	})
}
