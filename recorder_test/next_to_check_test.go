package recorder_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestNextToCheck_Success(t *testing.T) {
	t.Run("succeeds fast when valid message is received first", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("long")
			conn.WriteJSON("no")
		}()

		// assert
		rec.Assert().NextToCheck(stringLongerThan3)
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("NextToCheck should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 50*time.Millisecond {
				t.Error("NextToCheck should succeed faster")
			}
		}
	})

	t.Run("succeeds when valid message arrives in order", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("no")
			conn.WriteJSON("long")
		}()

		// assert
		rec.Assert().OneToBe("no").NextToCheck(stringLongerThan3)
		rec.RunAssertions(50 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("NextToCheck should succeed, mockT output is:", getTestOutput(mockT))
		}
	})
}

func TestNextToCheck_Failure(t *testing.T) {
	t.Run("fails when first message is invalid", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("no")
		}()

		// assert
		rec.Assert().NextToCheck(stringLongerThan3)
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("NextToCheck should fail")
		}
	})

	t.Run("fails when message arrives in the wrong order", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("no")
			conn.WriteJSON("no")
			conn.WriteJSON("long")
		}()

		// assert
		rec.Assert().OneToBe("no").NextToCheck(stringLongerThan3)
		rec.RunAssertions(100 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("NextToCheck should fail")
		}
	})
}
