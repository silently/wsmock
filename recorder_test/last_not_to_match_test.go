package recorder_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestLastNotToMatch_Success(t *testing.T) {
	t.Run("succeeds on end when last message does not match", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("shoot")
			conn.WriteJSON("shoot")
			conn.WriteJSON("shoot")
			conn.WriteJSON("missed")
		}()

		// assert
		rec.NewAssertion().LastNotToMatch(goalRE)
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("LastNotToMatch should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed < 30*time.Millisecond {
				t.Error("LastNotToMatch should not succeed before timeout")
			}
		}
	})
}

func TestLastNotToMatch_Failure(t *testing.T) {
	t.Run("fails when last message matches", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("shoot")
			conn.WriteJSON("shoot")
			conn.WriteJSON("shoot")
			conn.WriteJSON("goaal")
		}()

		// assert
		rec.NewAssertion().LastNotToMatch(goalRE)
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("LastNotToMatch should fail")
		}
	})

	t.Run("fails when timeout occurs before last message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(30 * time.Millisecond)
			conn.WriteJSON("missed")
		}()

		// assert
		rec.NewAssertion().LastNotToMatch(goalRE)
		rec.RunAssertions(10 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("LastNotToMatch should fail because of timeout")
		}
	})
}
