package recorder_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestOneToContain(t *testing.T) {
	t.Run("succeeds when containing message is received before timeout", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)
		go serveWsHistory(conn)

		// script
		conn.Send(Message{"join", "room:1"})

		// assert
		rec.Assert().OneToContain("room:")
		before := time.Now()
		rec.RunAssertions(300 * time.Millisecond) // it's a max
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("OneToContain should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 150*time.Millisecond {
				t.Errorf("OneToContain should succeed faster")
			}
		}
	})

	t.Run("succeeds when containing JSON message is written", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)
		go serveWsHistory(conn)

		// script
		conn.Send(Message{"join", "room:1"})

		// assert
		rec.Assert().OneToContain("kind") // json field is lowercased
		rec.Assert().OneToContain("joined")
		rec.RunAssertions(200 * time.Millisecond) // it's a max

		if mockT.Failed() { // fail not expected
			t.Error("OneToContain should succeed, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when containing JSON message pointer is written", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)
		go serveWsHistory(conn)

		// script
		conn.Send(Message{"pointer", ""})

		// assert
		rec.Assert().OneToContain("kind") // json field is lowercased
		rec.Assert().OneToContain("pointer")
		rec.Assert().OneToContain("payload")
		rec.Assert().OneToContain("sent")
		rec.RunAssertions(200 * time.Millisecond) // it's a max

		if mockT.Failed() { // fail not expected
			t.Error("OneToContain should succeed, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when containing string is received before timeout", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)
		go serveWsLogStrings(conn)

		// script
		conn.Send("logs")

		// assert
		rec.Assert().OneToContain("log")
		rec.Assert().OneToContain("log1")
		before := time.Now()
		rec.RunAssertions(300 * time.Millisecond) // it's a max
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("OneToContain should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 150*time.Millisecond {
				t.Errorf("OneToContain should succeed faster")
			}
		}
	})

	t.Run("succeeds when containing bytes is received before timeout", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)
		go serveWsLogBytes(conn)

		// script
		conn.Send("logs")

		// assert
		rec.Assert().OneToContain("log")
		rec.Assert().OneToContain("log1")
		before := time.Now()
		rec.RunAssertions(300 * time.Millisecond) // it's a max
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("OneToContain should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 150*time.Millisecond {
				t.Errorf("OneToContain should succeed faster")
			}
		}
	})

	t.Run("fails when timeout occurs before containing message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)
		go serveWsHistory(conn)

		// script
		conn.Send(Message{"join", "room:1"})

		// assert
		rec.Assert().OneToContain("room:")
		rec.RunAssertions(75 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("OneToContain should fail because of timeout")
		}
	})

	t.Run("fails when no containing message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)
		go serveWsHistory(conn)

		// script
		conn.Send(Message{"join", "room:1"})

		// assert
		rec.Assert().OneToContain("notfound")
		rec.RunAssertions(300 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("OneToContain should fail because substr is not found")
		}
	})

	t.Run("fails when no containing string", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)
		go serveWsLogStrings(conn)

		// script
		conn.Send("logs")

		// assert
		rec.Assert().OneToContain("notfound")
		rec.RunAssertions(300 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("OneToContain should fail because substr is not found")
		}
	})
}
