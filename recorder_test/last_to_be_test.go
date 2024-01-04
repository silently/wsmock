package recorder_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestLastToBe_Success(t *testing.T) {
	t.Run("succeeds on end when last message is received", func(t *testing.T) {
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
		rec.NewAssertion().LastToBe("pong4")
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("LastToBe should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed < 30*time.Millisecond {
				t.Error("LastToBe should not succeed before timeout")
			}
		}
	})
}

func TestLastToBe_Failure(t *testing.T) {
	t.Run("fails when timeout occurs before last message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong2")
			conn.WriteJSON("pong3")
			time.Sleep(90 * time.Millisecond)
			conn.WriteJSON("pong4")
		}()

		// assert
		rec.NewAssertion().LastToBe("pong4")
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("LastToBe should fail because of timeout")
		}
	})

	t.Run("fails when last message differs", func(t *testing.T) {
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
		rec.NewAssertion().LastToBe("pong")
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("LastToBe should fail")
		}
	})

	t.Run("fails when no message received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// dumb script
		go conn.Send("ping")

		// assert
		rec.NewAssertion().LastToBe("pong")
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("LastToBe should fail")
		}
	})
}
