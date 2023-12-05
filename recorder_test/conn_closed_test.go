package recorder_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestConnClosed(t *testing.T) {
	t.Run("fails when conn is not closed", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)
		go serveWsHistory(conn)

		// script
		rec.Assert().ConnClosed()

		// assert
		rec.RunAssertions(10 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("ConnClosed should fail")
		}
	})

	t.Run("succeeds when conn is closed server-side", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)
		go serveWsHistory(conn)

		// script
		conn.Send(Message{"quit", ""})

		// assert
		rec.Assert().ConnClosed()
		rec.RunAssertions(200 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("ConnClosed should succeed because of serveWsHistory logic, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when conn is closed client-side", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)
		go serveWsHistory(conn)

		// script
		conn.Send(Message{"join", "room:1"})
		conn.Close()

		// assert
		rec.Assert().ConnClosed()
		rec.RunAssertions(50 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("ConnClosed should succeed because of explicit conn Close, mockT output is:", getTestOutput(mockT))
		}
	})
}
