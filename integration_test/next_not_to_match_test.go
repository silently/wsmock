package integration_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestNextNotToMatch_Success(t *testing.T) {
	t.Run("succeeds fast when not matching message is received first", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.Send("shoot")
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("missed")
		}()

		// assert
		rec.NewAssertion().NextNotToMatch(goalRE)
		before := time.Now()
		rec.RunAssertions(10 * durationUnit)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("NextNotToMatch should succeed, mockT output is:\n", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 50*time.Millisecond {
				t.Error("NextNotToMatch should succeed faster")
			}
		}
	})

	t.Run("succeeds when not matching message arrives in order", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.Send("shoot")
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("post")
			conn.WriteJSON("missed")
		}()

		// assert
		rec.NewAssertion().OneToBe("post").NextNotToMatch(goalRE)
		rec.RunAssertions(5 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("NextNotToMatch should succeed, mockT output is:\n", getTestOutput(mockT))
		}
	})
}

func TestNextNotToMatch_Failure(t *testing.T) {
	t.Run("fails when first message does match", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("goal")
		}()

		// assert
		rec.NewAssertion().NextNotToMatch(goalRE)
		rec.RunAssertions(5 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("NextNotToMatch should fail")
		}
	})

	t.Run("fails when second message does match", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("defender")
			conn.WriteJSON("goooooal")
		}()

		// assert
		rec.NewAssertion().OneToBe("defender").NextNotToMatch(goalRE)
		rec.RunAssertions(10 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("NextNotToMatch should fail")
		}
	})
}
