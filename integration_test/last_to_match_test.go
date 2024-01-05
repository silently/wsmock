package integration_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestLastToMatch_Success(t *testing.T) {
	t.Run("succeeds on end when last message is matching", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.Send("shoot")
			conn.Send("shoot")
			conn.Send("shoot")
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("goooooal")
		}()

		// assert
		rec.NewAssertion().LastToMatch(goalRE)
		before := time.Now()
		rec.RunAssertions(10 * durationUnit)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("LastToMatch should succeed, mockT output is:\n", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed < 30*time.Millisecond {
				t.Error("LastToMatch should not succeed before timeout")
			}
		}
	})
}

func TestLastToMatch_Failure(t *testing.T) {
	t.Run("fails when timeout occurs before last message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.Send("shoot")
			conn.Send("shoot")
			conn.Send("shoot")
			time.Sleep(9 * durationUnit)
			conn.WriteJSON("goooooal")
		}()

		// assert
		rec.NewAssertion().LastToMatch(goalRE)
		rec.RunAssertions(5 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("LastToMatch should fail because of timeout")
		}
	})

	t.Run("fails when last message does not match", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.Send("shoot")
			conn.Send("shoot")
			conn.Send("shoot")
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("missed")
		}()

		// assert
		rec.NewAssertion().LastToMatch(goalRE)
		rec.RunAssertions(5 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("LastToMatch should fail")
		}
	})

	t.Run("fails when no message received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// dumb script
		go conn.Send("ping")

		// assert
		rec.NewAssertion().LastToMatch(goalRE)
		rec.RunAssertions(5 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("LastToMatch should fail")
		}
	})
}
