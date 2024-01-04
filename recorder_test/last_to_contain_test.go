package recorder_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestLastToContain_Success(t *testing.T) {
	t.Run("succeeds on end when last message is containing", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("nothing")
			conn.WriteJSON("nothing")
			conn.WriteJSON("nothing")
			conn.WriteJSON(Message{"nothing", "special"})
		}()

		// assert
		rec.NewAssertion().LastToContain("spec")
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("LastToContain should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed < 30*time.Millisecond {
				t.Error("LastToContain should not succeed before timeout")
			}
		}
	})
}

func TestLastToCheck_Failure(t *testing.T) {
	t.Run("fails when timeout occurs before last message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("nothing")
			conn.WriteJSON("nothing")
			conn.WriteJSON("nothing")
			time.Sleep(90 * time.Millisecond)
			conn.WriteJSON(Message{"nothing", "special"})
		}()

		// assert
		rec.NewAssertion().LastToContain("spec")
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("LastToContain should fail because of timeout")
		}
	})

	t.Run("fails when last message does not contain", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("nothing")
			conn.WriteJSON("nothing")
			conn.WriteJSON("nothing")
			conn.WriteJSON("nothing")
		}()

		// assert
		rec.NewAssertion().LastToContain("spec")
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("LastToContain should fail")
		}
	})

	t.Run("fails when no message received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// dumb script
		go conn.Send("ping")

		// assert
		rec.NewAssertion().LastToContain("spec")
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("LastToContain should fail")
		}
	})
}
