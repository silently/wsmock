package recorder_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestNextToBe(t *testing.T) {
	t.Run("succeeds fast when first message is received before timeout", func(t *testing.T) {
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
			conn.WriteJSON("pong4")
		}()

		// assert
		rec.Assert().NextToBe("pong1")
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("NextToBe should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 30*time.Millisecond {
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
			conn.Send("ping")
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong2")
			conn.WriteJSON("pong3")
			conn.WriteJSON("pong4")
		}()

		// assert
		rec.Assert().OneToBe("pong2").NextToBe("pong3")
		rec.RunAssertions(100 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("NextToBe should succeed, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("fails when timeout occurs before first message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.Send("ping")
			time.Sleep(30 * time.Millisecond)
			conn.WriteJSON("pong")
		}()

		// assert
		rec.Assert().NextToBe("pong")
		rec.RunAssertions(10 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("NextToBe should fail because of timeout")
		}
	})

	t.Run("fails when first message differs", func(t *testing.T) {
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
		rec.Assert().NextToBe("pongpong")
		rec.RunAssertions(20 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("NextToBe should fail")
			// } else { // TODO?
			// 	output := getTestOutput(mockT)
			// 	if !strings.Contains(output, "should be: {chat sentence2}, received: {chat sentence1}") {
			// 		t.Errorf("NextToBe unexpected error message: \"%v\"", output)
			// 	}
		}
	})

	t.Run("fails when message arrives in the wrong order", func(t *testing.T) {
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
			conn.WriteJSON("pong4")
		}()

		// assert
		rec.Assert().OneToBe("pong1").NextToBe("pong3")
		rec.RunAssertions(100 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("NextToBe should fail")
		}
	})
}
