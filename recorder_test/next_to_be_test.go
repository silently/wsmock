package recorder_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestNextToBe(t *testing.T) {
	t.Run("succeeds when first message is received before timeout", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		rec.Assert().NextToBe(Message{"chat", "sentence1"})
		before := time.Now()
		rec.RunAssertions(300 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("NextToBe should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 40*time.Millisecond {
				t.Error("NextToBe should succeed faster")
			}
		}
	})

	t.Run("fails when timeout occurs before first message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		rec.Assert().NextToBe(Message{"chat", "sentence1"})
		rec.RunAssertions(5 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("NextToBe should fail because of timeout")
		}
	})

	t.Run("fails when first message differs", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		rec.Assert().NextToBe(Message{"chat", "sentence2"})
		rec.RunAssertions(20 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("NextToBe should fail")
			// } else {
			// 	output := getTestOutput(mockT)
			// 	if !strings.Contains(output, "should be: {chat sentence2}, received: {chat sentence1}") {
			// 		t.Errorf("NextToBe unexpected error message: \"%v\"", output)
			// 	}
		}
	})
}
