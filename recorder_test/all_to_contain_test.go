package recorder_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestAllToContain_Success(t *testing.T) {
	t.Run("succeeds on end when all messages are containing", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON(Message{"nothing", "special"})
			conn.WriteJSON(Message{"nothing", "special"})
			conn.WriteJSON(Message{"nothing", "special"})
		}()

		// assert
		rec.Assert().AllToContain("spec")
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("AllToContain should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed < 30*time.Millisecond {
				t.Error("AllToContain should not succeed before timeout")
			}
		}
	})

	t.Run("succeeds when not containing message is received too late", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON(Message{"nothing", "special"})
			conn.WriteJSON(Message{"nothing", "special"})
			time.Sleep(60 * time.Millisecond)
			conn.WriteJSON("no")
		}()

		// assert
		rec.Assert().AllToContain("spec")
		rec.RunAssertions(50 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("AllToContain should succeed, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when conn is closed before not containing message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON(Message{"nothing", "special"})
			conn.WriteJSON(Message{"nothing", "special"})
			time.Sleep(50 * time.Millisecond)
			conn.WriteJSON("no")
		}()
		go func() {
			time.Sleep(20 * time.Millisecond)
			conn.Close()
		}()

		// assert
		rec.Assert().AllToContain("spec")
		rec.RunAssertions(100 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("AllToContain should succeed because of Close")
		}
	})
}

func TestAllToContain_Failure(t *testing.T) {
	t.Run("fails fast when one message does not contain", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON(Message{"nothing", "special"})
			conn.WriteJSON(Message{"nothing", "special"})
			conn.WriteJSON("no")
		}()

		// assert
		rec.Assert().AllToContain("spec")
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("AllToContain should fail (message is received)")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 50*time.Millisecond {
				t.Errorf("AllToContain should fail faster")
			}
		}
	})
}
