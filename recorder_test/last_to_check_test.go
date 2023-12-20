package recorder_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestLastToCheck_Success(t *testing.T) {
	t.Run("succeeds on end when last message is checking", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("no")
			conn.WriteJSON("no")
			conn.WriteJSON("no")
			conn.WriteJSON("long")
		}()

		// assert
		rec.Assert().LastToCheck(stringLongerThan3)
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("LastToCheck should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed < 30*time.Millisecond {
				t.Error("LastToCheck should not succeed before timeout")
			}
		}
	})
}

func TestLastToContain_Failure(t *testing.T) {
	t.Run("fails when timeout occurs before last message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("no")
			conn.WriteJSON("no")
			conn.WriteJSON("no")
			time.Sleep(90 * time.Millisecond)
			conn.WriteJSON("long")
		}()

		// assert
		rec.Assert().LastToCheck(stringLongerThan3)
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("LastToCheck should fail because of timeout")
		}
	})

	t.Run("fails when last message does not check", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("no")
			conn.WriteJSON("no")
			conn.WriteJSON("no")
			conn.WriteJSON("no")
		}()

		// assert
		rec.Assert().LastToCheck(stringLongerThan3)
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("LastToCheck should fail")
		}
	})

	t.Run("fails when no message received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// dumb script
		go conn.Send("ping")

		// assert
		rec.Assert().LastToCheck(stringLongerThan3)
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("LastToCheck should fail")
		}
	})
}
