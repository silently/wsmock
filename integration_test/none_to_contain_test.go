package integration_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestNoneToContain_Success(t *testing.T) {
	t.Run("succeeds when no message contains", func(t *testing.T) {
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
		rec.NewAssertion().NoneToContain("spec")
		rec.RunAssertions(5 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("NoneToContain should succeed, mockT output is:\n", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when containing message is received too late", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("no")
			time.Sleep(6 * durationUnit)
			conn.WriteJSON(Message{"nothing", "special"})
		}()

		// assert
		rec.NewAssertion().NoneToContain("spec")
		rec.RunAssertions(4 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("NoneToContain should succeed, mockT output is:\n", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when conn is closed before containing message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("no")
			time.Sleep(5 * durationUnit)
			conn.WriteJSON(Message{"nothing", "special"})
		}()
		go func() {
			time.Sleep(2 * durationUnit)
			conn.Close()
		}()

		// assert
		rec.NewAssertion().NoneToContain("spec")
		rec.RunAssertions(10 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("NoneToContain should succeed because of Close")
		}
	})
}

func TestNoneToContain_Failure(t *testing.T) {
	t.Run("fails fast when containg message is received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("no")
			conn.WriteJSON("no")
			conn.WriteJSON(Message{"nothing", "special"})
			conn.WriteJSON("no")
		}()

		// assert
		rec.NewAssertion().NoneToContain("spec")
		before := time.Now()
		rec.RunAssertions(10 * durationUnit)
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("NoneToContain should fail (message is received)")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 5*durationUnit {
				t.Errorf("NoneToContain should fail faster")
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
			conn.WriteJSON(Message{"nothing", "special"})
		}()

		// short assert does not catch pong2
		rec.NewAssertion().NoneToContain("spec")
		rec.RunAssertions(5 * durationUnit)

		rec.NewAssertion().NoneToContain("spec")
		rec.RunAssertions(5 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("NoneToContain should fail because of message history")
		}
	})
}
