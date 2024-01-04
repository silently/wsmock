package recorder_test

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
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("goal")
			conn.WriteJSON("goaal")
			conn.WriteJSON("goaaal")
		}()

		// assert
		rec.NewAssertion().AllToMatch(goalRE)
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("AllToMatch should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed < 30*time.Millisecond {
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
			time.Sleep(60 * time.Millisecond)
			conn.WriteJSON("missed")
		}()

		// assert
		rec.NewAssertion().AllToMatch(goalRE)
		rec.RunAssertions(50 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("AllToMatch should succeed, mockT output is:", getTestOutput(mockT))
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
			time.Sleep(50 * time.Millisecond)
			conn.WriteJSON("missed")
		}()
		go func() {
			time.Sleep(20 * time.Millisecond)
			conn.Close()
		}()

		// assert
		rec.NewAssertion().AllToMatch(goalRE)
		rec.RunAssertions(100 * time.Millisecond)

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
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("AllToMatch should fail (message is received)")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 50*time.Millisecond {
				t.Errorf("AllToMatch should fail faster")
			}
		}
	})
}
