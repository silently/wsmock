package integration_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestNextNotToContain_Success(t *testing.T) {
	t.Run("succeeds fast when not containing message is received first", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON(Message{"nothing", "special"})
		}()

		// assert
		rec.NewAssertion().NextNotToContain("else")
		before := time.Now()
		rec.RunAssertions(10 * durationUnit)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("NextNotToContain should succeed, mockT output is:\n", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 5*durationUnit {
				t.Error("NextNotToContain should succeed faster")
			}
		}
	})

	t.Run("succeeds when not containing message arrives in order", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("no")
			conn.WriteJSON(Message{"nothing", "special"})
		}()

		// assert
		rec.NewAssertion().OneToBe("no").NextNotToContain("else")
		rec.RunAssertions(5 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("NextNotToContain should succeed, mockT output is:\n", getTestOutput(mockT))
		}
	})
}

func TestNextNotToContain_Failure(t *testing.T) {
	t.Run("fails when first message is containing", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON(Message{"nothing", "special"})
		}()

		// assert
		rec.NewAssertion().NextNotToContain("spec")
		rec.RunAssertions(5 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("NextNotToContain should fail")
		}
	})

	t.Run("fails when second message is containing", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("no")
			conn.WriteJSON(Message{"nothing", "special"})
			conn.WriteJSON("no")
		}()

		// assert
		rec.NewAssertion().OneToBe("no").NextNotToContain("spec")
		rec.RunAssertions(10 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("NextNotToContain should fail")
		}
	})
}
