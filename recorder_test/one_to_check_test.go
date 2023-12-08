package recorder_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func stringLongerThan3(msg any) bool {
	time.Sleep(10 * time.Millisecond)
	if str, ok := msg.(string); ok {
		return len(str) > 3
	}
	return false
}

func TestOneToCheck_Success(t *testing.T) {
	t.Run("succeeds fast when valid message is received before timeout", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("sentence")
		}()

		// assert
		rec.Assert().OneToCheck(stringLongerThan3)
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("OneToCheck should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 20*time.Millisecond {
				t.Errorf("OneToCheck should succeed faster")
			}
		}
	})

	t.Run("succeeds when valid message is received among others", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("no")
			conn.WriteJSON("0")
			conn.WriteJSON("sentence")
			conn.WriteJSON("ko")
		}()

		// assert
		rec.Assert().OneToCheck(stringLongerThan3)
		rec.RunAssertions(50 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("OneToCheck should succeed, mockT output is:", getTestOutput(mockT))
		}
	})
}

func TestOneToCheck_Failure(t *testing.T) {

	t.Run("fails when no message received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.Send("ping")
		}()

		// assert
		rec.Assert().OneToCheck(stringLongerThan3)
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("OneToCheck should fail because no message is received")
		}
	})

	t.Run("fails when only invalid messages are received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("p1")
			conn.WriteJSON("p2")
		}()

		rec.Assert().OneToCheck(stringLongerThan3)
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("OneToCheck should fail because of unexpected message")
		}
	})

	t.Run("fails when timeout occurs before valid message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(60 * time.Millisecond)
			conn.WriteJSON("sentence")
		}()

		// assert
		rec.Assert().OneToCheck(stringLongerThan3)
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("OneToCheck should fail because of timeout")
		}
	})

	t.Run("fails fast when conn is closed before valid message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(50 * time.Millisecond)
			conn.WriteJSON("pong")
		}()
		go func() {
			time.Sleep(20 * time.Millisecond)
			conn.Close()
		}()

		// assert
		rec.Assert().OneToCheck(stringLongerThan3)
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("OneToCheck should fail because of Close")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 30*time.Millisecond {
				t.Error("OneToCheck should fail faster because of Close")
			}
		}
	})
}
