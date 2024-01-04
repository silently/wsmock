package recorder_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestAllToCheck_Success(t *testing.T) {
	t.Run("succeeds on end when all messages are equal", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("long")
			conn.WriteJSON("long")
			conn.WriteJSON("long")
		}()

		// assert
		rec.NewAssertion().AllToCheck(stringLongerThan3)
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("AllToCheck should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed < 30*time.Millisecond {
				t.Error("AllToCheck should not succeed before timeout")
			}
		}
	})

	t.Run("succeeds when not checking message is received too late", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("long")
			conn.WriteJSON("long")
			time.Sleep(60 * time.Millisecond)
			conn.WriteJSON("no")
		}()

		// assert
		rec.NewAssertion().AllToCheck(stringLongerThan3)
		rec.RunAssertions(50 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("AllToCheck should succeed, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when conn is closed before message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("long")
			conn.WriteJSON("long")
			time.Sleep(50 * time.Millisecond)
			conn.WriteJSON("no")
		}()
		go func() {
			time.Sleep(20 * time.Millisecond)
			conn.Close()
		}()

		// assert
		rec.NewAssertion().AllToCheck(stringLongerThan3)
		rec.RunAssertions(100 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("AllToCheck should succeed because of Close")
		}
	})
}

func TestAllToCheck_Failure(t *testing.T) {
	t.Run("fails fast when one message does not check", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("long")
			conn.WriteJSON("long")
			conn.WriteJSON("no")
		}()

		// assert
		rec.NewAssertion().AllToCheck(stringLongerThan3)
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("AllToCheck should fail (message is received)")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 50*time.Millisecond {
				t.Errorf("AllToCheck should fail faster")
			}
		}
	})
}
