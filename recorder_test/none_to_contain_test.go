package recorder_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestNoneToContain_Success(t *testing.T) {
	t.Run("succeeds when no message contains", func(t *testing.T) {
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
		rec.Assert().NoneToContain("spec")
		rec.RunAssertions(50 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("NoneToContain should succeed, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when containing message is received too late", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("no")
			time.Sleep(60 * time.Millisecond)
			conn.WriteJSON(Message{"nothing", "special"})
		}()

		// assert
		rec.Assert().NoneToContain("spec")
		rec.RunAssertions(50 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("NoneToContain should succeed, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when conn is closed before containing message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("no")
			time.Sleep(50 * time.Millisecond)
			conn.WriteJSON(Message{"nothing", "special"})
		}()
		go func() {
			time.Sleep(20 * time.Millisecond)
			conn.Close()
		}()

		// assert
		rec.Assert().NoneToContain("spec")
		rec.RunAssertions(100 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("NoneToContain should succeed because of Close")
		}
	})
}

func TestNoneToContain_Failure(t *testing.T) {
	t.Run("fails fast when containg message is received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("no")
			conn.WriteJSON("no")
			conn.WriteJSON(Message{"nothing", "special"})
			conn.WriteJSON("no")
		}()

		// assert
		rec.Assert().NoneToContain("spec")
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("NoneToContain should fail (message is received)")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 50*time.Millisecond {
				t.Errorf("NoneToContain should fail faster")
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
			conn.WriteJSON(Message{"nothing", "special"})
		}()

		// short assert does not catch pong2
		rec.Assert().NoneToContain("spec")
		rec.RunAssertions(50 * time.Millisecond)

		rec.Assert().NoneToContain("spec")
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("NoneToContain should fail because of message history")
		}
	})
}
