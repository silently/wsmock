package recorder_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestNextToMatch_Success(t *testing.T) {
	t.Run("succeeds fast when matching message is received first", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.Send("shoot")
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("goooooal")
		}()

		// assert
		rec.Assert().NextToMatch(goalRE)
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("NextToMatch should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 50*time.Millisecond {
				t.Error("NextToMatch should succeed faster")
			}
		}
	})

	t.Run("succeeds when matching message arrives in order", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.Send("shoot")
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("post")
			conn.WriteJSON("goooooal")
		}()

		// assert
		rec.Assert().OneToBe("post").NextToMatch(goalRE)
		rec.RunAssertions(50 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("NextToMatch should succeed, mockT output is:", getTestOutput(mockT))
		}
	})
}

func TestNextToMatch_Failure(t *testing.T) {
	t.Run("fails when first message does not match", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("no")
		}()

		// assert
		rec.Assert().NextToMatch(goalRE)
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("NextToMatch should fail")
		}
	})

	t.Run("fails when matching message arrives in the wrong order", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("defender")
			conn.WriteJSON("post")
			conn.WriteJSON("goooooal")
		}()

		// assert
		rec.Assert().OneToBe("defender").NextToMatch(goalRE)
		rec.RunAssertions(100 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("NextToMatch should fail")
		}
	})
}
