package wsmock

import (
	"testing"
	"time"
)

func TestAssertReceived(t *testing.T) {
	t.Run("succeeds when message is received before timeout", func(t *testing.T) {
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

	t.Run("fails when timeout occurs before message", func(t *testing.T) {
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

	t.Run("fails when conn is closed before message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		conn.Send(Message{"join", "room:1"})
		go func() {
			time.Sleep(75 * time.Millisecond)
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
	t.Run("succeeds when message is not received", func(t *testing.T) {
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

	t.Run("fails when message is received", func(t *testing.T) {
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
	t.Run("fails when conn is not closed", func(t *testing.T) {
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

	t.Run("succeeds when conn is closed server-side", func(t *testing.T) {
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

	t.Run("succeeds when conn is closed client-side", func(t *testing.T) {
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
