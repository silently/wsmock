package wsmock

import (
	"testing"
	"time"
)

func TestAssertReceived(t *testing.T) {
	t.Run("message written to conn", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"join", "room:1"})
		rec.AssertReceived(Message{"joined", "room:1"})
		before := time.Now()
		rec.RunAssertions(300 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("message should be received")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 150*time.Millisecond {
				t.Error("message should be received faster")
			}
		}
	})

	t.Run("message written to conn too late", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		conn.Send(Message{"join", "room:1"})
		rec.AssertReceived(Message{"joined", "room:1"})
		rec.RunAssertions(75 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("message should not be already received")
		}
	})

	t.Run("message written to closed conn", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		conn.Send(Message{"join", "room:1"})
		go func() {
			<-time.After(75 * time.Millisecond)
			conn.Close()
		}()
		rec.AssertReceived(Message{"joined", "room:1"})
		before := time.Now()
		rec.RunAssertions(200 * time.Millisecond)
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("message should not be received after Close")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 100*time.Millisecond {
				t.Error("fail should be received faster")
			}
		}
	})
}

func TestAssertNotReceived(t *testing.T) {
	t.Run("message not written to conn", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		conn.Send(Message{"join", "room:1"})
		rec.AssertNotReceived(Message{"not", "planned"})
		rec.RunAssertions(110 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("message should not be received")
		}
	})

	t.Run("message written to conn", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		conn.Send(Message{"join", "room:1"})
		rec.AssertNotReceived(Message{"joined", "room:1"})
		rec.RunAssertions(110 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("message should be received")
		}
	})
}

func TestAssertClosed(t *testing.T) {
	t.Run("no event implying close", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		rec.AssertClosed()
		rec.RunAssertions(time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("closed status should not be reported")
		}
	})

	t.Run("quit event implying close", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"quit", ""})
		rec.AssertClosed()
		rec.RunAssertions(200 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("closed status should be reported")
		}
	})
}
