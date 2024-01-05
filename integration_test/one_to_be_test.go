package integration_test

import (
	"strings"
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func removeSpaces(s string) (out string) {
	out = strings.Replace(s, "\n", "", -1)
	out = strings.Replace(out, "\t", "", -1)
	out = strings.Replace(out, " ", "", -1)
	return
}

func TestOneToBe_Success(t *testing.T) {
	t.Run("succeeds fast when equal message is received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("pong")
		}()

		// assert
		rec.NewAssertion().OneToBe("pong")
		before := time.Now()
		rec.RunAssertions(10 * durationUnit)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("OneToBe should succeed, mockT output is:\n", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 50*time.Millisecond {
				t.Errorf("OneToBe should succeed faster")
			}
		}
	})

	t.Run("succeeds when equal message is received among others", func(t *testing.T) {
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
		rec.NewAssertion().OneToBe("pong2")
		rec.RunAssertions(10 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("OneToBe should succeed, mockT output is:\n", getTestOutput(mockT))
		}
	})
}

func TestOneToBe_Failure(t *testing.T) {
	t.Run("fails when no message is received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// dumb script
		go conn.Send("ping")

		// assert
		rec.NewAssertion().OneToBe(stringLongerThan3)
		rec.RunAssertions(5 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("OneToBe should fail because no message is received")
		}
	})

	t.Run("fails when only non equal message are received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong2")
		}()

		rec.NewAssertion().OneToBe("pong3")
		rec.RunAssertions(5 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("OneToBe should fail because of unexpected message")
		}
		// assert error message
		expectedErrorMessage := "no message is equal to: pong3"
		processedErrorMessage := removeSpaces(expectedErrorMessage)
		processedActualErrorMessage := removeSpaces(getTestOutput(mockT))
		if !strings.Contains(processedActualErrorMessage, processedErrorMessage) {
			t.Errorf("OneToBe wrong error message, expected:\n\"%v\"", expectedErrorMessage)
		}
	})

	t.Run("fails when timeout occurs before equal message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(6 * durationUnit)
			conn.WriteJSON("pong")
		}()

		// assert
		rec.NewAssertion().OneToBe("pong")
		rec.RunAssertions(5 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("OneToBe should fail because of timeout")
		}
	})

	t.Run("fails fast when conn is closed before message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(5 * durationUnit)
			conn.WriteJSON("pong")
		}()
		go func() {
			time.Sleep(2 * durationUnit)
			conn.Close()
		}()

		// assert
		rec.NewAssertion().OneToBe("pong")
		before := time.Now()
		rec.RunAssertions(10 * durationUnit)
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("OneToBe should fail because of Close")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 50*time.Millisecond {
				t.Error("OneToBe should fail faster because of Close")
			}
		}
	})
}
