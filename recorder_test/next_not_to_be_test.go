package recorder_test

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
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong2")
			conn.WriteJSON("pong3")
			conn.WriteJSON("pong4")
		}()

		// assert
		rec.Assert().NextNotToBe("pong2")
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("NextNotToBe should succeed, mockT output is:", getTestOutput(mockT))
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
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong2")
			conn.WriteJSON("pong3")
			conn.WriteJSON("pong4")
		}()

		// assert
		rec.Assert().OneToBe("pong2").NextNotToBe("pong4")
		rec.RunAssertions(50 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("NextNotToBe should succeed, mockT output is:", getTestOutput(mockT))
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
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("pong1")
		}()

		// assert
		rec.Assert().NextNotToBe("pong1")
		rec.RunAssertions(50 * time.Millisecond)

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
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong3")
			conn.WriteJSON("pong2")
			conn.WriteJSON("pong4")
		}()

		// assert
		rec.Assert().OneToBe("pong1").NextNotToBe("pong3")
		rec.RunAssertions(50 * time.Millisecond)

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
			time.Sleep(30 * time.Millisecond)
			conn.WriteJSON("pong")
		}()

		// assert
		rec.Assert().NextNotToBe("pong")
		rec.RunAssertions(10 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("NextToBe should fail because of timeout")
		}
	})
}
