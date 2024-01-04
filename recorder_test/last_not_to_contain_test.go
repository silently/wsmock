package recorder_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestLastNotToContain_Success(t *testing.T) {
	t.Run("succeeds on end when last message does not contain", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON(Message{"nothing", "special"})
			conn.WriteJSON(Message{"nothing", "special"})
			conn.WriteJSON(Message{"nothing", "special"})
			conn.WriteJSON("other")
		}()

		// assert
		rec.NewAssertion().LastNotToContain("spec")
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("LastNotToContain should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed < 30*time.Millisecond {
				t.Error("LastNotToContain should not succeed before timeout")
			}
		}
	})
}

func TestLastNotToContain_Failure(t *testing.T) {
	t.Run("fails when last message contains", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("other")
			conn.WriteJSON("other")
			conn.WriteJSON("other")
			conn.WriteJSON(Message{"nothing", "special"})
		}()

		// assert
		rec.NewAssertion().LastNotToContain("spec")
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("LastNotToContain should fail")
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
		rec.NewAssertion().LastNotToContain("spec")
		rec.RunAssertions(10 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("LastNotToContain should fail because of timeout")
		}
	})
}
