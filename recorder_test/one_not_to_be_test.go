package recorder_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestOneNotToBe_Success(t *testing.T) {
	t.Run("succeeds fast when a different message is received first", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong2")
		}()

		// assert
		rec.Assert().OneNotToBe("pong2")
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("OneNotToBe should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 50*time.Millisecond {
				t.Errorf("OneNotToBe should succeed faster")
			}
		}
	})

	t.Run("succeeds fast when a different message is received second", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong2")
		}()

		// assert
		rec.Assert().OneNotToBe("pong1")
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("OneNotToBe should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 50*time.Millisecond {
				t.Errorf("OneNotToBe should succeed faster")
			}
		}
	})
}

func TestOneNotToBe_Failure(t *testing.T) {
	t.Run("fails when no message is received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// dumb script
		go conn.Send("ping")

		// assert
		rec.Assert().OneNotToBe("pong")
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("OneNotToBe should fail because no message is received")
		}
	})

	t.Run("fails when only same message is received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("pong")
			conn.WriteJSON("pong")
		}()

		// assert
		rec.Assert().OneNotToBe("pong")
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("OneNotToBe should fail (message is received)")
		}
	})

	t.Run("fails when timeout occurs before message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("pong1")
			time.Sleep(60 * time.Millisecond)
			conn.WriteJSON("pong2")
		}()

		// assert
		rec.Assert().OneNotToBe("pong1")
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("OneNotToBe should fail because of timeout")
		}
	})

	t.Run("fails when conn closed occurs before different message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("pong1")
			time.Sleep(60 * time.Millisecond)
			conn.WriteJSON("pong2")
		}()

		go func() {
			time.Sleep(20 * time.Millisecond)
			conn.Close()
		}()

		// assert
		rec.Assert().OneNotToBe("pong1")
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("OneNotToBe should fail because of conn closed")
		}
	})
}
