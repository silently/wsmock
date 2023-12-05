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

		// script
		conn.Send("ping")

		// assert
		rec.Assert().ConnClosed()
		rec.RunAssertions(10 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("ConnClosed should fail")
		}
	})

	t.Run("succeeds when conn is closed", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.Send("ping")
			time.Sleep(10 * time.Millisecond)
			conn.Close()
		}()

		// assert
		rec.Assert().ConnClosed()
		rec.RunAssertions(100 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("ConnClosed should succeed because of explicit close, mockT output is:", getTestOutput(mockT))
		}
	})
}
