package recorder_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestOneToBeOneNotTobe(t *testing.T) {
	t.Run("should fail fast", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		rec.Assert().OneToBe(Message{"chat", "sentence4"})
		rec.Assert().OneNotToBe(Message{"chat", "sentence1"}) // failing assertion
		before := time.Now()
		rec.RunAssertions(300 * time.Millisecond) // it's a max
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("OneNotToBe should fail")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 100*time.Millisecond {
				t.Errorf("OneNotToBe should fail faster")
			}
		}
	})
}

func TestMultiRunAssertionsion(t *testing.T) {
	t.Run("should succeed without blocking", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})
		rec.Assert().OneToBe(Message{"chat", "sentence1"})
		rec.Assert().OneToBe(Message{"chat", "sentence2"})
		rec.Assert().OneToBe(Message{"chat", "sentence3"})
		rec.Assert().OneToBe(Message{"chat", "sentence4"})

		// no assertion!
		rec.RunAssertions(100 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("several OneToBe should not fail")
		}
	})
}
