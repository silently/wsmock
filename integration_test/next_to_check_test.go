package integration_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestNextToCheck_Success(t *testing.T) {
	t.Run("succeeds fast when valid message is received first", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("long")
			conn.WriteJSON("no")
		}()

		// assert
		rec.NewAssertion().NextToCheck(stringLongerThan3)
		before := time.Now()
		rec.RunAssertions(10 * durationUnit)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("NextToCheck should succeed, mockT output is:\n", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 5*durationUnit {
				t.Error("NextToCheck should succeed faster")
			}
		}
	})

	t.Run("succeeds when valid message arrives in order", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("no")
			conn.WriteJSON("long")
		}()

		// assert
		rec.NewAssertion().OneToBe("no").NextToCheck(stringLongerThan3)
		rec.RunAssertions(5 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("NextToCheck should succeed, mockT output is:\n", getTestOutput(mockT))
		}
	})
}

func TestNextToCheck_Failure(t *testing.T) {
	t.Run("fails when first message is invalid", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("no")
		}()

		// assert
		rec.NewAssertion().NextToCheck(stringLongerThan3)
		rec.RunAssertions(5 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("NextToCheck should fail")
		}
	})

	t.Run("fails when message arrives in the wrong order", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("no")
			conn.WriteJSON("no")
			conn.WriteJSON("long")
		}()

		// assert
		rec.NewAssertion().OneToBe("no").NextToCheck(stringLongerThan3)
		rec.RunAssertions(10 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("NextToCheck should fail")
		}
	})
}
