package integration_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestAllToContain_Success(t *testing.T) {
	t.Run("succeeds on end when all messages are containing", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON(Message{"nothing", "special"})
			conn.WriteJSON(Message{"nothing", "special"})
			conn.WriteJSON(Message{"nothing", "special"})
		}()

		// assert
		rec.NewAssertion().AllToContain("spec")
		before := time.Now()
		rec.RunAssertions(10 * durationUnit)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("AllToContain should succeed, mockT output is:\n", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed < 3*durationUnit {
				t.Error("AllToContain should not succeed before timeout")
			}
		}
	})

	t.Run("succeeds when not containing message is received too late", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON(Message{"nothing", "special"})
			conn.WriteJSON(Message{"nothing", "special"})
			time.Sleep(6 * durationUnit)
			conn.WriteJSON("no")
		}()

		// assert
		rec.NewAssertion().AllToContain("spec")
		rec.RunAssertions(5 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("AllToContain should succeed, mockT output is:\n", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when conn is closed before not containing message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON(Message{"nothing", "special"})
			conn.WriteJSON(Message{"nothing", "special"})
			time.Sleep(5 * durationUnit)
			conn.WriteJSON("no")
		}()
		go func() {
			time.Sleep(2 * durationUnit)
			conn.Close()
		}()

		// assert
		rec.NewAssertion().AllToContain("spec")
		rec.RunAssertions(10 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("AllToContain should succeed because of Close")
		}
	})
}

func TestAllToContain_Failure(t *testing.T) {
	t.Run("fails fast when one message does not contain", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON(Message{"nothing", "special"})
			conn.WriteJSON(Message{"nothing", "special"})
			conn.WriteJSON("no")
		}()

		// assert
		rec.NewAssertion().AllToContain("spec")
		before := time.Now()
		rec.RunAssertions(10 * durationUnit)
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("AllToContain should fail (message is received)")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 5*durationUnit {
				t.Errorf("AllToContain should fail faster")
			}
		}
	})
}
