package recorder_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestLastNotToCheck_Success(t *testing.T) {
	t.Run("succeeds on end when last message does not check", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("no")
			conn.WriteJSON("no")
			conn.WriteJSON("long")
			conn.WriteJSON("no")
		}()

		// assert
		rec.Assert().LastNotToCheck(stringLongerThan3)
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("LastNotToCheck should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed < 30*time.Millisecond {
				t.Error("LastNotToCheck should not succeed before timeout")
			}
		}
	})
}

func TestLastNotToCheck_Failure(t *testing.T) {
	t.Run("fails when last message checks", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("no")
			conn.WriteJSON("no")
			conn.WriteJSON("no")
			conn.WriteJSON("long")
		}()

		// assert
		rec.Assert().LastNotToCheck(stringLongerThan3)
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("LastNotToCheck should fail")
		}
	})

	t.Run("fails when timeout occurs before last message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(30 * time.Millisecond)
			conn.WriteJSON("no")
		}()

		// assert
		rec.Assert().LastNotToCheck(stringLongerThan3)
		rec.RunAssertions(10 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("LastNotToCheck should fail because of timeout")
		}
	})
}
