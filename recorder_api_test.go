package wsmock

import (
	"strings"
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

		// assert
		rec.AssertReceived(Message{"joined", "room:1"})
		before := time.Now()
		rec.Run(300 * time.Millisecond) // it's a max
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("AssertReceived should succeed, mockT output is:", getTestOutput(mockT))
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

		// script
		conn.Send(Message{"join", "room:1"})

		// assert
		rec.AssertReceived(Message{"joined", "room:1"})
		rec.Run(75 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("AssertReceived should fail because of timeout")
		}
	})

	t.Run("fails when conn is closed before message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"join", "room:1"})
		go func() {
			time.Sleep(75 * time.Millisecond)
			conn.Close()
		}()

		// assert
		rec.AssertReceived(Message{"joined", "room:1"})
		before := time.Now()
		rec.Run(200 * time.Millisecond)
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

func TestAssertFirstReceived(t *testing.T) {
	t.Run("succeeds when first message is received before timeout", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		rec.AssertFirstReceived(Message{"chat", "sentence1"})
		before := time.Now()
		rec.Run(300 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("AssertFirstReceived should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 40*time.Millisecond {
				t.Error("AssertFirstReceived should succeed faster")
			}
		}
	})

	t.Run("fails when timeout occurs before first message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		rec.AssertFirstReceived(Message{"chat", "sentence1"})
		rec.Run(5 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("AssertFirstReceived should fail because of timeout")
		}
	})

	t.Run("fails when first message differs", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		rec.AssertFirstReceived(Message{"chat", "sentence2"})
		rec.Run(20 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("AssertFirstReceived should fail")
		} else {
			output := getTestOutput(mockT)
			if !strings.Contains(output, "should be: {chat sentence2}, received: {chat sentence1}") {
				t.Errorf("AssertFirstReceived unexpected error message: \"%v\"", output)
			}
		}
	})
}

func TestAssertLastReceivedOnTimeout(t *testing.T) {
	t.Run("succeeds when last message is received before timeout", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		rec.AssertLastReceivedOnTimeout(Message{"chat", "sentence4"})
		before := time.Now()
		rec.Run(300 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("AssertLastReceivedOnTimeout should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed < 300*time.Millisecond {
				t.Error("AssertLastReceivedOnTimeout should not succeed before timeout")
			}
		}
	})

	t.Run("fails when timeout occurs before last message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		rec.AssertLastReceivedOnTimeout(Message{"chat", "sentence4"})
		rec.Run(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("AssertLastReceived should fail because of timeout")
		}
	})

	t.Run("fails when last message differs", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		rec.AssertLastReceivedOnTimeout(Message{"chat", "sentence5"})
		rec.Run(300 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("AssertLastReceivedOnTimeout should fail")
		} else {
			output := getTestOutput(mockT)
			if !strings.Contains(output, "should be: {chat sentence5}, received: {chat sentence4}") {
				t.Errorf("AssertLastReceivedOnTimeout unexpected error message: \"%v\"", output)
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

		// script
		conn.Send(Message{"join", "room:1"})

		// assert
		rec.AssertNotReceived(Message{"not", "planned"})
		rec.Run(110 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("AssertNotReceived should succeed, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("fails when message is received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"join", "room:1"})

		// assert
		rec.AssertNotReceived(Message{"joined", "room:1"})
		rec.Run(110 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("AssertNotReceived should fail (message is received)")
		}
	})
}

func TestAssertReceivedContains(t *testing.T) {
	t.Run("succeeds when containing message is received before timeout", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"join", "room:1"})

		// assert
		rec.AssertReceivedContains("room:")
		before := time.Now()
		rec.Run(300 * time.Millisecond) // it's a max
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("AssertReceivedContains should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 150*time.Millisecond {
				t.Errorf("AssertReceivedContains should succeed faster")
			}
		}
	})

	t.Run("fails when timeout occurs before containing message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"join", "room:1"})

		// assert
		rec.AssertReceivedContains("room:")
		rec.Run(75 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("AssertReceivedContains should fail because of timeout")
		}
	})

	t.Run("fails when no containing message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"join", "room:1"})

		// assert
		rec.AssertReceivedContains("notfound")
		rec.Run(300 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("AssertReceivedContains should fail because substr is not found")
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

		// assert
		rec.Run(time.Millisecond)

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

		// assert
		rec.AssertClosed()
		rec.Run(200 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("AssertClosed should succeed because of serveWsStub logic, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when conn is closed client-side", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"join", "room:1"})
		conn.Close()

		// assert
		rec.AssertClosed()
		rec.Run(50 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("AssertClosed should succeed because of explicit conn Close, mockT output is:", getTestOutput(mockT))
		}
	})
}

func TestNoAssertion(t *testing.T) {
	t.Run("should succeed", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"join", "room:1"})

		// no assertion!
		rec.Run(time.Millisecond)

		if mockT.Failed() { // fail expected
			t.Error("NoAssertion can't fail")
		}
	})
}
