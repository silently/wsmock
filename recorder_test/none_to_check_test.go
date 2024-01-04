package recorder_test

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
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("no")
			conn.WriteJSON("no")
			conn.WriteJSON("no")
		}()

		// assert
		rec.NewAssertion().NoneToCheck(stringLongerThan3)
		rec.RunAssertions(50 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("NoneToCheck should succeed, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when matching message is received too late", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("no")
			time.Sleep(60 * time.Millisecond)
			conn.WriteJSON("long")
		}()

		// assert
		rec.NewAssertion().NoneToCheck(stringLongerThan3)
		rec.RunAssertions(50 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("NoneToCheck should succeed, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when conn is closed before matching message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("no")
			time.Sleep(50 * time.Millisecond)
			conn.WriteJSON("long")
		}()
		go func() {
			time.Sleep(20 * time.Millisecond)
			conn.Close()
		}()

		// assert
		rec.NewAssertion().NoneToCheck(stringLongerThan3)
		rec.RunAssertions(100 * time.Millisecond)

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
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("NoneToCheck should fail (message is received)")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 50*time.Millisecond {
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
			time.Sleep(60 * time.Millisecond)
			conn.WriteJSON("long")
		}()

		// short assert does not catch pong2
		rec.NewAssertion().NoneToCheck(stringLongerThan3)
		rec.RunAssertions(50 * time.Millisecond)

		rec.NewAssertion().NoneToCheck(stringLongerThan3)
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("NoneToCheck should fail because of message history")
		}
	})
}
