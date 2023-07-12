package wsmock

import (
	"testing"
	"time"
)

func TestAssertReceivedSparseSequence(t *testing.T) {
	t.Run("succeeds when sparse sequence is received before timeout", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		seq := []any{Message{"chat", "sentence1"}, Message{"chat", "sentence3"}}
		rec.AssertReceivedSparseSequence(seq)
		before := time.Now()
		rec.Run(300 * time.Millisecond)
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
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		seq := []any{Message{"chat", "sentence1"}, Message{"chat", "sentence3"}}
		rec.AssertReceivedSparseSequence(seq)
		rec.Run(35 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("AssertReceivedSparseSequence shoud fail because of timeout")
		}
	})

	t.Run("fails when sparse sequence order differ", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		seq := []any{Message{"chat", "sentence2"}, Message{"chat", "sentence1"}}

		rec.AssertReceivedSparseSequence(seq)
		rec.Run(100 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("AssertReceivedSparseSequence should fail")
		}
	})

	t.Run("fails when sparse sequence is incomplete", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		seq := []any{
			Message{"chat", "sentence1"},
			Message{"chat", "sentence2"},
			Message{"chat", "sentence3"},
			Message{"chat", "sentence4"},
			Message{"chat", "sentence5"},
		}
		rec.AssertReceivedSparseSequence(seq)
		rec.Run(100 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("AssertReceivedSparseSequence should fail")
		}
	})

	t.Run("succeeds twice in same Run", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		seq := []any{Message{"chat", "sentence1"}, Message{"chat", "sentence3"}}

		rec.AssertReceivedSparseSequence(seq)
		rec.AssertReceivedSparseSequence(seq) // twice
		rec.Run(150 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("AssertReceivedSparseSequence should succeed, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("succeeds once on two Runs", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		seq := []any{Message{"chat", "sentence1"}, Message{"chat", "sentence3"}}

		rec.AssertReceivedSparseSequence(seq)
		rec.Run(150 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("AssertReceivedSparseSequence should succeed, mockT output is:", getTestOutput(mockT))
		}

		rec.AssertReceivedSparseSequence(seq)
		rec.Run(150 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("AssertReceivedSparseSequence should fail for second Run")
		}
	})
}

func TestAssertReceivedAdjacentSequence(t *testing.T) {
	t.Run("succeeds when adjacent sequence is received before timeout", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		seq := []any{Message{"chat", "sentence2"}, Message{"chat", "sentence3"}}
		rec.AssertReceivedAdjacentSequence(seq)

		before := time.Now()
		rec.Run(300 * time.Millisecond)
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
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		seq := []any{Message{"chat", "sentence2"}, Message{"chat", "sentence4"}}

		rec.AssertReceivedAdjacentSequence(seq)
		rec.Run(100 * time.Millisecond)

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
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		seq := []any{
			Message{"chat", "sentence1"},
			Message{"chat", "sentence2"},
			Message{"chat", "sentence3"},
			Message{"chat", "sentence4"},
		}
		rec.AssertReceivedExactSequence(seq)
		rec.Run(300 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("AssertReceivedExactSequence should succeed, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("fails when sequence differs", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		seq := []any{
			Message{"chat", "sentence1"},
			Message{"chat", "sentence2"},
			Message{"chat", "sentence3"},
			Message{"chat", "sentence5"},
		}

		rec.AssertReceivedExactSequence(seq)
		rec.Run(100 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("AssertReceivedExactSequence should fail")
		}
	})

	t.Run("fails when sequence misses last message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		seq := []any{Message{"chat", "sentence1"}, Message{"chat", "sentence2"}, Message{"chat", "sentence3"}}
		rec.AssertReceivedExactSequence(seq)
		rec.Run(100 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("AssertReceivedExactSequence should fail")
		}
	})
}
