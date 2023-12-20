package recorder_test

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
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON(Message{"nothing", "special"})
		}()

		// assert
		rec.Assert().NextNotToContain("else")
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("NextNotToContain should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 50*time.Millisecond {
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
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("no")
			conn.WriteJSON(Message{"nothing", "special"})
		}()

		// assert
		rec.Assert().OneToBe("no").NextNotToContain("else")
		rec.RunAssertions(50 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("NextNotToContain should succeed, mockT output is:", getTestOutput(mockT))
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
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON(Message{"nothing", "special"})
		}()

		// assert
		rec.Assert().NextNotToContain("spec")
		rec.RunAssertions(50 * time.Millisecond)

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
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("no")
			conn.WriteJSON(Message{"nothing", "special"})
			conn.WriteJSON("no")
		}()

		// assert
		rec.Assert().OneToBe("no").NextNotToContain("spec")
		rec.RunAssertions(100 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("NextNotToContain should fail")
		}
	})
}
