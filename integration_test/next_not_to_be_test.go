package integration_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestNextNotToBe_Success(t *testing.T) {
	t.Run("succeeds fast when first not equal message is received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong2")
			conn.WriteJSON("pong3")
			conn.WriteJSON("pong4")
		}()

		// assert
		rec.NewAssertion().NextNotToBe("pong2")
		before := time.Now()
		rec.RunAssertions(10 * durationUnit)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("NextNotToBe should succeed, mockT output is:\n", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 50*time.Millisecond {
				t.Error("NextToBe should succeed faster")
			}
		}
	})

	t.Run("succeeds when message arrives in order", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong2")
			conn.WriteJSON("pong3")
			conn.WriteJSON("pong4")
		}()

		// assert
		rec.NewAssertion().OneToBe("pong2").NextNotToBe("pong4")
		rec.RunAssertions(5 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("NextNotToBe should succeed, mockT output is:\n", getTestOutput(mockT))
		}
	})
}

func TestNextNotToBe_Failure(t *testing.T) {
	t.Run("fails when first message is equal", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("pong1")
		}()

		// assert
		rec.NewAssertion().NextNotToBe("pong1")
		rec.RunAssertions(5 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("NextNotToBe should fail")
		}
	})

	t.Run("fails when message order is not expected", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong3")
			conn.WriteJSON("pong2")
			conn.WriteJSON("pong4")
		}()

		// assert
		rec.NewAssertion().OneToBe("pong1").NextNotToBe("pong3")
		rec.RunAssertions(5 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("NextNotToBe should fail")
		}
	})

	t.Run("fails when timeout occurs before first message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(3 * durationUnit)
			conn.WriteJSON("pong")
		}()

		// assert
		rec.NewAssertion().NextNotToBe("pong")
		rec.RunAssertions(1 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("NextToBe should fail because of timeout")
		}
	})
}
