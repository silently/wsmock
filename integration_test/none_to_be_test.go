package integration_test

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
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong2")
			conn.WriteJSON("pong3")
		}()

		// assert
		rec.NewAssertion().NoneToBe("pong4")
		rec.RunAssertions(5 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("NoneToBe should succeed, mockT output is:\n", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when message is received too late", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("pong1")
			time.Sleep(6 * durationUnit)
			conn.WriteJSON("pong2")
			conn.WriteJSON("pong3")
		}()

		// assert
		rec.NewAssertion().NoneToBe("pong2")
		rec.RunAssertions(5 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("NoneToBe should succeed, mockT output is:\n", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when conn is closed before message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("pong1")
			time.Sleep(5 * durationUnit)
			conn.WriteJSON("pong2")
		}()
		go func() {
			time.Sleep(2 * durationUnit)
			conn.Close()
		}()

		// assert
		rec.NewAssertion().NoneToBe("pong2")
		rec.RunAssertions(10 * durationUnit)

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
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong2")
			conn.WriteJSON("pong3")
		}()

		// assert
		rec.NewAssertion().NoneToBe("pong3")
		before := time.Now()
		rec.RunAssertions(10 * durationUnit)
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
			time.Sleep(6 * durationUnit)
			conn.WriteJSON("pong2")
		}()

		// short assert does not catch pong2
		rec.NewAssertion().NoneToBe("pong2")
		rec.RunAssertions(5 * durationUnit)

		rec.NewAssertion().NoneToBe("pong2")
		rec.RunAssertions(5 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("NoneToBe should fail because of message history")
		}
	})
}
