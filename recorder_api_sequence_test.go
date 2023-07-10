package wsmock

import (
	"testing"
	"time"
)

func toAnySlice[T any](expecteds []T) []any {
	expectedsAny := make([]any, len(expecteds))
	for i, v := range expecteds {
		expectedsAny[i] = v
	}
	return expectedsAny
}

func TestAssertReceivedSparseSequence(t *testing.T) {
	t.Run("succeeds when sparse sequence is received before timeout", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})
		seq := toAnySlice([]Message{{"chat", "sentence1"}, {"chat", "sentence3"}})
		rec.AssertReceivedSparseSequence(seq)

		before := time.Now()
		rec.RunAssertions(300 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("AssertReceivedSparseSequence should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 150*time.Millisecond {
				t.Error("AssertReceivedSparseSequence should succeed faster")
			}
		}
	})

	t.Run("fails when timeout occurs before sparse sequence", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})
		seq := toAnySlice([]Message{{"chat", "sentence1"}, {"chat", "sentence3"}})
		rec.AssertReceivedSparseSequence(seq)

		rec.RunAssertions(35 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("AssertReceivedSparseSequence shoud fail because of timeout")
		}
	})

	t.Run("fails when sparse sequence order differ", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})

		seq := toAnySlice([]Message{{"chat", "sentence2"}, {"chat", "sentence1"}})
		rec.AssertReceivedSparseSequence(seq)
		rec.RunAssertions(100 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("AssertReceivedSparseSequence should fail")
		}
	})

	t.Run("fails when sparse sequence is incomplete", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})

		seq := toAnySlice([]Message{{"chat", "sentence1"}, {"chat", "sentence2"}, {"chat", "sentence3"}, {"chat", "sentence4"}, {"chat", "sentence5"}})
		rec.AssertReceivedSparseSequence(seq)
		rec.RunAssertions(100 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("AssertReceivedSparseSequence should fail")
		}
	})
}

func TestAssertReceivedAdjacentSequence(t *testing.T) {
	t.Run("succeeds when adjacent sequence is received before timeout", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})

		seq := toAnySlice([]Message{{"chat", "sentence2"}, {"chat", "sentence3"}})
		rec.AssertReceivedAdjacentSequence(seq)

		before := time.Now()
		rec.RunAssertions(300 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("AssertReceivedAdjacentSequence should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 150*time.Millisecond {
				t.Error("AssertReceivedAdjacentSequence should succeed faster")
			}
		}
	})

	t.Run("fails when adjacent sequence is incomplete", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})

		seq := toAnySlice([]Message{{"chat", "sentence2"}, {"chat", "sentence4"}})
		rec.AssertReceivedAdjacentSequence(seq)
		rec.RunAssertions(100 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("AssertReceivedAdjacentSequence should fail")
		}
	})
}

func TestAssertReceivedExactSequence(t *testing.T) {
	t.Run("succeeds when exact sequence is received before timeout", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})
		seq := toAnySlice([]Message{{"chat", "sentence1"}, {"chat", "sentence2"}, {"chat", "sentence3"}, {"chat", "sentence4"}})
		rec.AssertReceivedExactSequence(seq)

		rec.RunAssertions(300 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("AssertReceivedExactSequence should succeed, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("fails when sequence differs", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})

		seq := toAnySlice([]Message{{"chat", "sentence1"}, {"chat", "sentence2"}, {"chat", "sentence3"}, {"chat", "sentence5"}})
		rec.AssertReceivedExactSequence(seq)
		rec.RunAssertions(100 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("AssertReceivedExactSequence should fail")
		}
	})

	t.Run("fails when sequence misses last message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})

		seq := toAnySlice([]Message{{"chat", "sentence1"}, {"chat", "sentence2"}, {"chat", "sentence3"}})
		rec.AssertReceivedExactSequence(seq)
		rec.RunAssertions(100 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("AssertReceivedExactSequence should fail")
		}
	})
}
