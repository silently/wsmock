package integration_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestNoneToCheck_Success(t *testing.T) {
	t.Run("succeeds when no message checks", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("no")
			conn.WriteJSON("no")
			conn.WriteJSON("no")
		}()

		// assert
		rec.NewAssertion().NoneToCheck(stringLongerThan3)
		rec.RunAssertions(5 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("NoneToCheck should succeed, mockT output is:\n", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when matching message is received too late", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("no")
			time.Sleep(6 * durationUnit)
			conn.WriteJSON("long")
		}()

		// assert
		rec.NewAssertion().NoneToCheck(stringLongerThan3)
		rec.RunAssertions(5 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("NoneToCheck should succeed, mockT output is:\n", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when conn is closed before matching message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("no")
			time.Sleep(5 * durationUnit)
			conn.WriteJSON("long")
		}()
		go func() {
			time.Sleep(2 * durationUnit)
			conn.Close()
		}()

		// assert
		rec.NewAssertion().NoneToCheck(stringLongerThan3)
		rec.RunAssertions(10 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("NoneToCheck should succeed because of Close")
		}
	})
}

func TestNoneToCheck_Failure(t *testing.T) {
	t.Run("fails fast when matching message is received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("no")
			conn.WriteJSON("no")
			conn.WriteJSON("long")
			conn.WriteJSON("no")
		}()

		// assert
		rec.NewAssertion().NoneToCheck(stringLongerThan3)
		before := time.Now()
		rec.RunAssertions(10 * durationUnit)
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("NoneToCheck should fail (message is received)")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 5*durationUnit {
				t.Errorf("NoneToCheck should fail faster than %v", elapsed)
			}
		}
	})

	t.Run("fails in second run", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("no")
			time.Sleep(6 * durationUnit)
			conn.WriteJSON("long")
		}()

		// short assert does not catch pong2
		rec.NewAssertion().NoneToCheck(stringLongerThan3)
		rec.RunAssertions(5 * durationUnit)

		rec.NewAssertion().NoneToCheck(stringLongerThan3)
		rec.RunAssertions(5 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("NoneToCheck should fail because of message history")
		}
	})
}
