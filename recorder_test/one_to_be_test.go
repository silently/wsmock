package recorder_test

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

func TestOneToBe(t *testing.T) {
	t.Run("succeeds when message is received before timeout", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.Send("ping")
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("pong")
		}()

		// assert
		rec.Assert().OneToBe("pong")
		before := time.Now()
		rec.RunAssertions(20 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("OneToBe should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 150*time.Millisecond {
				t.Errorf("OneToBe should succeed faster")
			}
		}
	})

	t.Run("fails when timeout occurs before message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.Send("ping")
			time.Sleep(50 * time.Millisecond)
			conn.WriteJSON("pong")
		}()

		// assert
		rec.Assert().OneToBe("pong")
		rec.RunAssertions(20 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("OneToBe should fail because of timeout")
		}
	})

	t.Run("fails when conn is closed before message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.Send("ping")
			time.Sleep(50 * time.Millisecond)
			conn.WriteJSON("pong")
		}()
		go func() {
			time.Sleep(20 * time.Millisecond)
			conn.Close()
		}()

		// assert
		rec.Assert().OneToBe("pong")
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("OneToBe should fail because of Close")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 30*time.Millisecond {
				t.Error("OneToBe should fail faster because of Close")
			}
		}
	})

	t.Run("fails with correct message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.Send("ping")
			time.Sleep(20 * time.Millisecond)
			conn.WriteJSON("pong")
		}()

		rec.Assert().OneToBe("pongpong")
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("OneToBe should fail because of unexpected message")
		}
		// assert error message
		expectedErrorMessage := "message not received\nexpected: pongpong"
		processedErrorMessage := removeSpaces(expectedErrorMessage)
		processedActualErrorMessage := removeSpaces(getTestOutput(mockT))
		if !strings.Contains(processedActualErrorMessage, processedErrorMessage) {
			t.Errorf("OneToBe wrong error message, expected:\n\"%v\"", expectedErrorMessage)
		}
	})

	t.Run("succeeds twice in same Run", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.Send("ping")
			time.Sleep(20 * time.Millisecond)
			conn.WriteJSON("pong")
		}()

		// script
		conn.Send(Message{"history", ""})

		// assert
		rec.Assert().OneToBe("pong")
		rec.Assert().OneToBe("pong") // twice is ok

		rec.RunAssertions(50 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("OneToBe should succeed, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("succeeds once on two RunS", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.Send("ping")
			time.Sleep(20 * time.Millisecond)
			conn.WriteJSON("pong")
		}()

		// assert
		rec.Assert().OneToBe("pong")
		rec.RunAssertions(50 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("OneToBe should succeed, mockT output is:", getTestOutput(mockT))
		}

		rec.Assert().OneToBe("pong")
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("OneToBe should fail for second Run", getTestOutput(mockT))
		}
	})
}
