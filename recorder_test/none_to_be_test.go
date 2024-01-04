package recorder_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestNoneToBe_Success(t *testing.T) {
	t.Run("succeeds when message is not received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong2")
			conn.WriteJSON("pong3")
		}()

		// assert
		rec.NewAssertion().NoneToBe("pong4")
		rec.RunAssertions(50 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("NoneToBe should succeed, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when message is received too late", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("pong1")
			time.Sleep(60 * time.Millisecond)
			conn.WriteJSON("pong2")
			conn.WriteJSON("pong3")
		}()

		// assert
		rec.NewAssertion().NoneToBe("pong2")
		rec.RunAssertions(50 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("NoneToBe should succeed, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when conn is closed before message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("pong1")
			time.Sleep(50 * time.Millisecond)
			conn.WriteJSON("pong2")
		}()
		go func() {
			time.Sleep(20 * time.Millisecond)
			conn.Close()
		}()

		// assert
		rec.NewAssertion().NoneToBe("pong2")
		rec.RunAssertions(100 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("NoneToBe should succeed because of Close")
		}
	})
}

func TestNoneToBe_Failure(t *testing.T) {
	t.Run("fails fast when message is received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.Send("ping")
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong2")
			conn.WriteJSON("pong3")
		}()

		// assert
		rec.NewAssertion().NoneToBe("pong3")
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("NoneToBe should fail (message is received)")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 50*time.Millisecond {
				t.Errorf("NoneToBe should fail faster")
			}
		}
	})

	t.Run("fails in second run", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("pong1")
			time.Sleep(60 * time.Millisecond)
			conn.WriteJSON("pong2")
		}()

		// short assert does not catch pong2
		rec.NewAssertion().NoneToBe("pong2")
		rec.RunAssertions(50 * time.Millisecond)

		rec.NewAssertion().NoneToBe("pong2")
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("NoneToBe should fail because of message history")
		}
	})
}
