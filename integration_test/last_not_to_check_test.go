package integration_test

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
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("no")
			conn.WriteJSON("no")
			conn.WriteJSON("long")
			conn.WriteJSON("no")
		}()

		// assert
		rec.NewAssertion().LastNotToCheck(stringLongerThan3)
		before := time.Now()
		rec.RunAssertions(10 * durationUnit)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("LastNotToCheck should succeed, mockT output is:\n", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed < 3*durationUnit {
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
		rec.NewAssertion().LastNotToCheck(stringLongerThan3)
		rec.RunAssertions(5 * durationUnit)

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
			time.Sleep(3 * durationUnit)
			conn.WriteJSON("no")
		}()

		// assert
		rec.NewAssertion().LastNotToCheck(stringLongerThan3)
		rec.RunAssertions(1 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("LastNotToCheck should fail because of timeout")
		}
	})
}
