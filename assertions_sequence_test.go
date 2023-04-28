package wsmock

import (
	"testing"
	"time"
)

func TestAssertReceivedSparseSequence(t *testing.T) {
	t.Run("messages written includes sparse sequence", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})
		seq := ToAnySlice([]Message{{"chat", "sentence1"}, {"chat", "sentence3"}})
		rec.AssertReceivedSparseSequence(seq)

		before := time.Now()
		rec.RunAssertions(300 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("sparse sequence should be received")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 150*time.Millisecond {
				t.Error("sparse sequence should be received faster")
			}
		}
	})

	t.Run("messages written includes sparse sequence, but too late", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})
		seq := ToAnySlice([]Message{{"chat", "sentence1"}, {"chat", "sentence3"}})
		rec.AssertReceivedSparseSequence(seq)

		rec.RunAssertions(40 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("sparse sequence should not be already received")
		}
	})

	t.Run("messages written and sparse sequence order don't match", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})

		seq := ToAnySlice([]Message{{"chat", "sentence2"}, {"chat", "sentence1"}})
		rec.AssertReceivedSparseSequence(seq)
		rec.RunAssertions(100 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("sparse sequence should not be received")
		}
	})

	t.Run("messages written is shorter than sparse sequence", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})

		seq := ToAnySlice([]Message{{"chat", "sentence1"}, {"chat", "sentence2"}, {"chat", "sentence3"}, {"chat", "sentence4"}, {"chat", "sentence5"}})
		rec.AssertReceivedSparseSequence(seq)
		rec.RunAssertions(100 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("sparse sequence should not be received")
		}
	})
}

func TestAssertReceivedAdjacentSequence(t *testing.T) {
	t.Run("messages written include adjacent sequence", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})

		seq := ToAnySlice([]Message{{"chat", "sentence2"}, {"chat", "sentence3"}})
		rec.AssertReceivedAdjacentSequence(seq)

		before := time.Now()
		rec.RunAssertions(300 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("adjacent sequence should be received")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 150*time.Millisecond {
				t.Error("adjacent sequence should be received faster")
			}
		}
	})

	t.Run("messages written does not include adjacent sequence", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})

		seq := ToAnySlice([]Message{{"chat", "sentence2"}, {"chat", "sentence4"}})
		rec.AssertReceivedAdjacentSequence(seq)
		rec.RunAssertions(100 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("adjacent sequence should not be received")
		}
	})
}

func TestAssertReceivedExactSequence(t *testing.T) {
	t.Run("messages written match exact sequence", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})
		seq := ToAnySlice([]Message{{"chat", "sentence1"}, {"chat", "sentence2"}, {"chat", "sentence3"}, {"chat", "sentence4"}})
		rec.AssertReceivedExactSequence(seq)

		rec.RunAssertions(300 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("exact sequence should be received")
		}
	})

	t.Run("messages written and sequence differ", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})

		seq := ToAnySlice([]Message{{"chat", "sentence1"}, {"chat", "sentence2"}, {"chat", "sentence3"}, {"chat", "sentence5"}})
		rec.AssertReceivedExactSequence(seq)
		rec.RunAssertions(100 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("exact sequence should not be received")
		}
	})

	t.Run("messages written miss last message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})

		seq := ToAnySlice([]Message{{"chat", "sentence1"}, {"chat", "sentence2"}, {"chat", "sentence3"}})
		rec.AssertReceivedExactSequence(seq)
		rec.RunAssertions(100 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("exact sequence should not be received")
		}
	})
}
