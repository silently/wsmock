package wsmock

import (
	"testing"
	"time"
)

func TestAssertReceived(t *testing.T) {
	t.Run("joined message sent before timeout", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"join", "room:1"})
		rec.AssertReceived(Message{"joined", "room:1"})
		before := time.Now()
		rec.RunAssertions(300 * time.Millisecond) // it's a max
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("AssertReceived should succeed")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 150*time.Millisecond {
				t.Errorf("AssertReceived should succeed faster")
			}
		}
	})

	t.Run("timeout comes before joined message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		conn.Send(Message{"join", "room:1"})
		rec.AssertReceived(Message{"joined", "room:1"})
		rec.RunAssertions(75 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("AssertReceived should fail because of timeout")
		}
	})

	t.Run("closed conn comes before joined message", func(t *testing.T) {
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
			t.Error("AssertReceived should fail because of Close")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 100*time.Millisecond {
				t.Error("AssertReceived should fail faster because of Close")
			}
		}
	})
}

func TestAssertNotReceived(t *testing.T) {
	t.Run("unexpected message not received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		conn.Send(Message{"join", "room:1"})
		rec.AssertNotReceived(Message{"not", "planned"})
		rec.RunAssertions(110 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("AssertNotReceived should succeed")
		}
	})

	t.Run("joined message sent", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		conn.Send(Message{"join", "room:1"})
		rec.AssertNotReceived(Message{"joined", "room:1"})
		rec.RunAssertions(110 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("AssertNotReceived should fail (message is received)")
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
			t.Error("AssertClosed should fail")
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
			t.Error("AssertClosed succeed because of serveWsStub logic")
		}
	})

	t.Run("explicit close", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		conn.Send(Message{"join", "room:1"})
		conn.Close()
		rec.AssertClosed()
		rec.RunAssertions(50 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("AssertClosed succeed because of explicit conn Close")
		}
	})
}
