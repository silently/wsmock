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
	t.Run("succeeds when checked message is received before timeout", func(t *testing.T) {
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

	t.Run("succeeds when at least one checked message is received", func(t *testing.T) {
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
		rec.RunAssertions(30 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("OneToCheck should succeed, mockT output is:", getTestOutput(mockT))
		}
	})
}

func TestOneToCheck_Failure(t *testing.T) {
	t.Run("fails when timeout occurs before checked message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(50 * time.Millisecond)
			conn.WriteJSON("sentence")
		}()

		// assert
		rec.Assert().OneToCheck(stringLongerThan3)
		rec.RunAssertions(30 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("OneToCheck should fail because of timeout")
		}
	})

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
		rec.RunAssertions(30 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("OneToCheck should fail because no message is received")
		}
	})

	t.Run("fails when no checked message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("no")
		}()

		// assert
		rec.Assert().OneToCheck(stringLongerThan3)
		rec.RunAssertions(30 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("OneToCheck should fail because no message is checked")
		}
	})
}
