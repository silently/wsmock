package recorder_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestNextToContain_Success(t *testing.T) {
	t.Run("succeeds fast when containing message is received first", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON(Message{"nothing", "special"})
		}()

		// assert
		rec.Assert().NextToContain("spec")
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("NextToContain should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 50*time.Millisecond {
				t.Error("NextToContain should succeed faster")
			}
		}
	})

	t.Run("succeeds when containing message arrives in order", func(t *testing.T) {
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
		rec.Assert().OneToBe("no").NextToContain("spec")
		rec.RunAssertions(50 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("NextToContain should succeed, mockT output is:", getTestOutput(mockT))
		}
	})
}

func TestNextToContain_Failure(t *testing.T) {
	t.Run("fails when first message is not containing", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("no")
		}()

		// assert
		rec.Assert().NextToContain("spec")
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("NextToContain should fail")
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
			conn.WriteJSON(Message{"nothing", "special"})
		}()

		// assert
		rec.Assert().OneToBe("no").NextToContain("spec")
		rec.RunAssertions(100 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("NextToContain should fail")
		}
	})
}
