package integration_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestAllToBe_Success(t *testing.T) {
	t.Run("succeeds on end when all messages are equal", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong1")
		}()

		// assert
		rec.NewAssertion().AllToBe("pong1")
		before := time.Now()
		rec.RunAssertions(10 * durationUnit)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("AllToBe should succeed, mockT output is:\n", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed < 3*durationUnit {
				t.Error("AllToBe should not succeed before timeout")
			}
		}
	})

	t.Run("succeeds when not equal message is received too late", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong1")
			time.Sleep(6 * durationUnit)
			conn.WriteJSON("pong2")
		}()

		// assert
		rec.NewAssertion().AllToBe("pong1")
		rec.RunAssertions(4 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("AllToBe should succeed, mockT output is:\n", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when conn is closed before message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong1")
			time.Sleep(5 * durationUnit)
			conn.WriteJSON("pong2")
		}()
		go func() {
			time.Sleep(2 * durationUnit)
			conn.Close()
		}()

		// assert
		rec.NewAssertion().AllToBe("pong1")
		rec.RunAssertions(10 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("AllToBe should succeed because of Close")
		}
	})
}

func TestAllToBe_Failure(t *testing.T) {
	t.Run("fails fast when one message differs", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.Send("ping")
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong2")
		}()

		// assert
		rec.NewAssertion().AllToBe("pong1")
		before := time.Now()
		rec.RunAssertions(10 * durationUnit)
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("AllToBe should fail (message is received)")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 5*durationUnit {
				t.Errorf("AllToBe should fail faster")
			}
		}
	})
}
