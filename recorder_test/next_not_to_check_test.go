package recorder_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestNextNotToCheck_Success(t *testing.T) {
	t.Run("succeeds fast when invalid message is received first", func(t *testing.T) {
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
		rec.NewAssertion().NextNotToCheck(stringLongerThan3)
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("NextNotToCheck should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 50*time.Millisecond {
				t.Error("NextNotToCheck should succeed faster")
			}
		}
	})

	t.Run("succeeds when messages arrives in order", func(t *testing.T) {
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
		rec.NewAssertion().OneToBe("no").NextNotToCheck(stringLongerThan3)
		rec.RunAssertions(50 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("NextNotToCheck should succeed, mockT output is:", getTestOutput(mockT))
		}
	})
}

func TestNextNotToCheck_Failure(t *testing.T) {
	t.Run("fails when first message checks", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("long")
		}()

		// assert
		rec.NewAssertion().NextNotToCheck(stringLongerThan3)
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("NextNotToCheck should fail")
		}
	})

	t.Run("fails when second message does not check", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("no")
			conn.WriteJSON("long")
			conn.WriteJSON("no")
		}()

		// assert
		rec.NewAssertion().OneToBe("no").NextNotToCheck(stringLongerThan3)
		rec.RunAssertions(100 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("NextNotToCheck should fail")
		}
	})
}
