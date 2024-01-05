package integration_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestNoneToMatch_Success(t *testing.T) {
	t.Run("succeeds when no message matches", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("missed")
			conn.WriteJSON("missed")
			conn.WriteJSON("missed")
		}()

		// assert
		rec.NewAssertion().NoneToMatch(goalRE)
		rec.RunAssertions(5 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("NoneToMatch should succeed, mockT output is:\n", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when matching message is received too late", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("missed")
			time.Sleep(6 * durationUnit)
			conn.WriteJSON("goal")
		}()

		// assert
		rec.NewAssertion().NoneToMatch(goalRE)
		rec.RunAssertions(5 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("NoneToMatch should succeed, mockT output is:\n", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when conn is closed before matching message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("missed")
			time.Sleep(5 * durationUnit)
			conn.WriteJSON("goal")
		}()
		go func() {
			time.Sleep(2 * durationUnit)
			conn.Close()
		}()

		// assert
		rec.NewAssertion().NoneToMatch(goalRE)
		rec.RunAssertions(10 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("NoneToMatch should succeed because of Close")
		}
	})
}

func TestNoneToMatch_Failure(t *testing.T) {
	t.Run("fails fast when matching message is received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("missed")
			conn.WriteJSON("missed")
			conn.WriteJSON("goal")
			conn.WriteJSON("missed")
		}()

		// assert
		rec.NewAssertion().NoneToMatch(goalRE)
		before := time.Now()
		rec.RunAssertions(10 * durationUnit)
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("NoneToMatch should fail (message is received)")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 5*durationUnit {
				t.Errorf("NoneToMatch should fail faster")
			}
		}
	})

	t.Run("fails in second run", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("missed")
			time.Sleep(6 * durationUnit)
			conn.WriteJSON("goal")
		}()

		// short assert does not catch pong2
		rec.NewAssertion().NoneToMatch(goalRE)
		rec.RunAssertions(5 * durationUnit)

		rec.NewAssertion().NoneToMatch(goalRE)
		rec.RunAssertions(5 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("NoneToMatch should fail because of message history")
		}
	})
}
