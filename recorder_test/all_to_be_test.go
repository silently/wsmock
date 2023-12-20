package recorder_test

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
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong1")
		}()

		// assert
		rec.Assert().AllToBe("pong1")
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("AllToBe should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed < 30*time.Millisecond {
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
			time.Sleep(60 * time.Millisecond)
			conn.WriteJSON("pong2")
		}()

		// assert
		rec.Assert().AllToBe("pong1")
		rec.RunAssertions(50 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("AllToBe should succeed, mockT output is:", getTestOutput(mockT))
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
			time.Sleep(50 * time.Millisecond)
			conn.WriteJSON("pong2")
		}()
		go func() {
			time.Sleep(20 * time.Millisecond)
			conn.Close()
		}()

		// assert
		rec.Assert().AllToBe("pong1")
		rec.RunAssertions(100 * time.Millisecond)

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
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong2")
		}()

		// assert
		rec.Assert().AllToBe("pong1")
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("AllToBe should fail (message is received)")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 50*time.Millisecond {
				t.Errorf("AllToBe should fail faster")
			}
		}
	})
}
