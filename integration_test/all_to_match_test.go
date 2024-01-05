package integration_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestAllToMatch_Success(t *testing.T) {
	t.Run("succeeds on end when all messages match", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("goal")
			conn.WriteJSON("goaal")
			conn.WriteJSON("goaaal")
		}()

		// assert
		rec.NewAssertion().AllToMatch(goalRE)
		before := time.Now()
		rec.RunAssertions(10 * durationUnit)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("AllToMatch should succeed, mockT output is:\n", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed < 3*durationUnit {
				t.Error("AllToMatch should not succeed before timeout")
			}
		}
	})

	t.Run("succeeds when not matching message is received too late", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("goal")
			conn.WriteJSON("goal")
			time.Sleep(6 * durationUnit)
			conn.WriteJSON("missed")
		}()

		// assert
		rec.NewAssertion().AllToMatch(goalRE)
		rec.RunAssertions(5 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("AllToMatch should succeed, mockT output is:\n", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when conn is closed before not matching message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("goal")
			conn.WriteJSON("goal")
			time.Sleep(5 * durationUnit)
			conn.WriteJSON("missed")
		}()
		go func() {
			time.Sleep(2 * durationUnit)
			conn.Close()
		}()

		// assert
		rec.NewAssertion().AllToMatch(goalRE)
		rec.RunAssertions(10 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("AllToMatch should succeed because of Close")
		}
	})
}

func TestAllToMatch_Failure(t *testing.T) {
	t.Run("fails fast when one message does not match", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("goal")
			conn.WriteJSON("goal")
			conn.WriteJSON("missed")
		}()

		// assert
		rec.NewAssertion().AllToMatch(goalRE)
		before := time.Now()
		rec.RunAssertions(10 * durationUnit)
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("AllToMatch should fail (message is received)")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 5*durationUnit {
				t.Errorf("AllToMatch should fail faster")
			}
		}
	})
}
