package wsmock

import (
	"testing"
	"time"
)

func TestToReceiveSparseSequence(t *testing.T) {
	t.Run("succeeds when sparse sequence is received before timeout", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		seq := []any{Message{"chat", "sentence1"}, Message{"chat", "sentence3"}}
		rec.ToReceiveSparseSequence(seq)
		before := time.Now()
		rec.RunAssertions(300 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("ToReceiveSparseSequence should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 150*time.Millisecond {
				t.Error("ToReceiveSparseSequence should succeed faster")
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
		rec.ToReceiveSparseSequence(seq)
		rec.RunAssertions(35 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("ToReceiveSparseSequence shoud fail because of timeout")
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

		rec.ToReceiveSparseSequence(seq)
		rec.RunAssertions(100 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("ToReceiveSparseSequence should fail")
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
		rec.ToReceiveSparseSequence(seq)
		rec.RunAssertions(100 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("ToReceiveSparseSequence should fail")
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

		rec.ToReceiveSparseSequence(seq)
		rec.ToReceiveSparseSequence(seq) // twice
		rec.RunAssertions(150 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("ToReceiveSparseSequence should succeed, mockT output is:", getTestOutput(mockT))
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

		rec.ToReceiveSparseSequence(seq)
		rec.RunAssertions(150 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("ToReceiveSparseSequence should succeed, mockT output is:", getTestOutput(mockT))
		}

		rec.ToReceiveSparseSequence(seq)
		rec.RunAssertions(150 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("ToReceiveSparseSequence should fail for second Run")
		}
	})
}

func TestToReceiveSequence(t *testing.T) {
	t.Run("succeeds when adjacent sequence is received before timeout", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		seq := []any{Message{"chat", "sentence2"}, Message{"chat", "sentence3"}}
		rec.ToReceiveSequence(seq)

		before := time.Now()
		rec.RunAssertions(300 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("ToReceiveSequence should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 150*time.Millisecond {
				t.Error("ToReceiveSequence should succeed faster")
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

		rec.ToReceiveSequence(seq)
		rec.RunAssertions(100 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("ToReceiveSequence should fail")
		}
	})
}

func TestToReceiveOnlySequence(t *testing.T) {
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
		rec.ToReceiveOnlySequence(seq)
		rec.RunAssertions(300 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("ToReceiveOnlySequence should succeed, mockT output is:", getTestOutput(mockT))
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

		rec.ToReceiveOnlySequence(seq)
		rec.RunAssertions(100 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("ToReceiveOnlySequence should fail")
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
		rec.ToReceiveOnlySequence(seq)
		rec.RunAssertions(100 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("ToReceiveOnlySequence should fail")
		}
	})
}
