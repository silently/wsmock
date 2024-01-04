package recorder_test

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
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("missed")
			conn.WriteJSON("missed")
			conn.WriteJSON("missed")
		}()

		// assert
		rec.NewAssertion().NoneToMatch(goalRE)
		rec.RunAssertions(50 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("NoneToMatch should succeed, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when matching message is received too late", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("missed")
			time.Sleep(60 * time.Millisecond)
			conn.WriteJSON("goal")
		}()

		// assert
		rec.NewAssertion().NoneToMatch(goalRE)
		rec.RunAssertions(50 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("NoneToMatch should succeed, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when conn is closed before matching message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("missed")
			time.Sleep(50 * time.Millisecond)
			conn.WriteJSON("goal")
		}()
		go func() {
			time.Sleep(20 * time.Millisecond)
			conn.Close()
		}()

		// assert
		rec.NewAssertion().NoneToMatch(goalRE)
		rec.RunAssertions(100 * time.Millisecond)

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
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("NoneToMatch should fail (message is received)")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 50*time.Millisecond {
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
			time.Sleep(60 * time.Millisecond)
			conn.WriteJSON("goal")
		}()

		// short assert does not catch pong2
		rec.NewAssertion().NoneToMatch(goalRE)
		rec.RunAssertions(50 * time.Millisecond)

		rec.NewAssertion().NoneToMatch(goalRE)
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("NoneToMatch should fail because of message history")
		}
	})
}
