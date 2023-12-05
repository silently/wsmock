package recorder_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestOneNotToBe(t *testing.T) {
	t.Run("succeeds when message is not received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)
		go serveWsHistory(conn)

		// script
		conn.Send(Message{"join", "room:1"})

		// assert
		rec.Assert().OneNotToBe(Message{"not", "planned"})
		rec.RunAssertions(110 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("OneNotToBe should succeed, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("fails when message is received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)
		go serveWsHistory(conn)

		// script
		conn.Send(Message{"join", "room:1"})

		// assert
		rec.Assert().OneNotToBe(Message{"joined", "room:1"})
		rec.RunAssertions(110 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("OneNotToBe should fail (message is received)")
		}
	})
}
