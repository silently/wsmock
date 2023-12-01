package wsmock

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func removeSpaces(s string) (out string) {
	out = strings.Replace(s, "\n", "", -1)
	out = strings.Replace(out, "\t", "", -1)
	out = strings.Replace(out, " ", "", -1)
	return
}

func TestToReceive(t *testing.T) {
	t.Run("succeeds when message is received before timeout", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"join", "room:1"})

		// assert
		rec.ToReceive(Message{"joined", "room:1"})
		before := time.Now()
		rec.RunAssertions(300 * time.Millisecond) // it's a max
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("ToReceive should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 150*time.Millisecond {
				t.Errorf("ToReceive should succeed faster")
			}
		}
	})

	t.Run("fails when timeout occurs before message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"join", "room:1"})

		// assert
		rec.ToReceive(Message{"joined", "room:1"})
		rec.RunAssertions(75 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("ToReceive should fail because of timeout")
		}
	})

	t.Run("fails when conn is closed before message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"join", "room:1"})
		go func() {
			time.Sleep(75 * time.Millisecond)
			conn.Close()
		}()

		// assert
		rec.ToReceive(Message{"joined", "room:1"})
		before := time.Now()
		rec.RunAssertions(200 * time.Millisecond)
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("ToReceive should fail because of Close")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 100*time.Millisecond {
				t.Error("ToReceive should fail faster because of Close")
			}
		}
	})

	t.Run("fails with correct message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"join", "room:1"})

		// assert
		target := Message{"not", "received"}
		rec.ToReceive(target)
		rec.RunAssertions(75 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("ToReceive should fail because of unexpected message")
		}
		// assert error message
		expectedErrorMessage := fmt.Sprintf("message not received\nexpected: %+v", target)
		processedErrorMessage := removeSpaces(expectedErrorMessage)
		processedActualErrorMessage := removeSpaces(getTestOutput(mockT))
		if !strings.Contains(processedActualErrorMessage, processedErrorMessage) {
			t.Errorf("ToReceive wrong error message, expected:\n\"%v\"", expectedErrorMessage)
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
		rec.ToReceive(Message{"chat", "sentence1"})
		rec.ToReceive(Message{"chat", "sentence1"}) // twice

		rec.RunAssertions(50 * time.Millisecond) // it's a max

		if mockT.Failed() { // fail not expected
			t.Error("ToReceive should succeed, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("succeeds once on two RunS", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		rec.ToReceive(Message{"chat", "sentence1"})
		rec.RunAssertions(50 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("ToReceive should succeed, mockT output is:", getTestOutput(mockT))
		}

		rec.ToReceive(Message{"chat", "sentence1"})
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("ToReceive should fail for second Run", getTestOutput(mockT))
		}
	})
}

func TestToReceiveNotReceived(t *testing.T) {
	t.Run("should fail fast", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		rec.ToReceive(Message{"chat", "sentence4"})
		rec.NotToReceive(Message{"chat", "sentence1"}) // failing assertion
		before := time.Now()
		rec.RunAssertions(300 * time.Millisecond) // it's a max
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("NotToReceive should fail")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 100*time.Millisecond {
				t.Errorf("NotToReceive should fail faster")
			}
		}
	})
}

func TestToReceiveFirst(t *testing.T) {
	t.Run("succeeds when first message is received before timeout", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		rec.ToReceiveFirst(Message{"chat", "sentence1"})
		before := time.Now()
		rec.RunAssertions(300 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("ToReceiveFirst should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 40*time.Millisecond {
				t.Error("ToReceiveFirst should succeed faster")
			}
		}
	})

	t.Run("fails when timeout occurs before first message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		rec.ToReceiveFirst(Message{"chat", "sentence1"})
		rec.RunAssertions(5 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("ToReceiveFirst should fail because of timeout")
		}
	})

	t.Run("fails when first message differs", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		rec.ToReceiveFirst(Message{"chat", "sentence2"})
		rec.RunAssertions(20 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("ToReceiveFirst should fail")
			// } else {
			// 	output := getTestOutput(mockT)
			// 	if !strings.Contains(output, "should be: {chat sentence2}, received: {chat sentence1}") {
			// 		t.Errorf("ToReceiveFirst unexpected error message: \"%v\"", output)
			// 	}
		}
	})
}

func TestToReceiveLast(t *testing.T) {
	t.Run("succeeds when last message is received before timeout", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		rec.ToReceiveLast(Message{"chat", "sentence4"})
		before := time.Now()
		rec.RunAssertions(300 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("ToReceiveLast should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed < 300*time.Millisecond {
				t.Error("ToReceiveLast should not succeed before timeout")
			}
		}
	})

	t.Run("fails when timeout occurs before last message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		rec.ToReceiveLast(Message{"chat", "sentence4"})
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("AssertLastReceived should fail because of timeout")
		}
	})

	t.Run("fails when last message differs", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		rec.ToReceiveLast(Message{"chat", "sentence5"})
		rec.RunAssertions(300 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("ToReceiveLast should fail")
			// } else {
			// 	output := getTestOutput(mockT)
			// 	if !strings.Contains(output, "should be: {chat sentence5}, received: {chat sentence4}") {
			// 		t.Errorf("ToReceiveLast unexpected error message: \"%v\"", output)
			// 	}
		}
	})

	t.Run("fails when no message received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script: nothing

		// assert
		rec.ToReceiveLast(Message{"chat", "sentence1"})
		rec.RunAssertions(100 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("ToReceiveLast should fail")
			// } else {
			// 	output := getTestOutput(mockT)
			// 	if !strings.Contains(output, "should be: {chat sentence1}, received none") {
			// 		t.Errorf("ToReceiveLast unexpected error message: \"%v\"", output)
			// 	}
		}
	})
}

func TestNotToReceive(t *testing.T) {
	t.Run("succeeds when message is not received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"join", "room:1"})

		// assert
		rec.NotToReceive(Message{"not", "planned"})
		rec.RunAssertions(110 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("NotToReceive should succeed, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("fails when message is received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"join", "room:1"})

		// assert
		rec.NotToReceive(Message{"joined", "room:1"})
		rec.RunAssertions(110 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("NotToReceive should fail (message is received)")
		}
	})
}

func TestToReceiveContaining(t *testing.T) {
	t.Run("succeeds when containing message is received before timeout", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"join", "room:1"})

		// assert
		rec.ToReceiveContaining("room:")
		before := time.Now()
		rec.RunAssertions(300 * time.Millisecond) // it's a max
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("ToReceiveContaining should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 150*time.Millisecond {
				t.Errorf("ToReceiveContaining should succeed faster")
			}
		}
	})

	t.Run("succeeds when containing JSON message is written", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"join", "room:1"})

		// assert
		rec.ToReceiveContaining("kind") // json field is lowercased
		rec.ToReceiveContaining("joined")
		rec.RunAssertions(200 * time.Millisecond) // it's a max

		if mockT.Failed() { // fail not expected
			t.Error("ToReceiveContaining should succeed, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when containing JSON message pointer is written", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"pointer", ""})

		// assert
		rec.ToReceiveContaining("kind") // json field is lowercased
		rec.ToReceiveContaining("pointer")
		rec.ToReceiveContaining("payload")
		rec.ToReceiveContaining("sent")
		rec.RunAssertions(200 * time.Millisecond) // it's a max

		if mockT.Failed() { // fail not expected
			t.Error("ToReceiveContaining should succeed, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when containing string is received before timeout", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsLogStrings(conn)

		// script
		conn.Send("logs")

		// assert
		rec.ToReceiveContaining("log")
		rec.ToReceiveContaining("log1")
		before := time.Now()
		rec.RunAssertions(300 * time.Millisecond) // it's a max
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("ToReceiveContaining should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 150*time.Millisecond {
				t.Errorf("ToReceiveContaining should succeed faster")
			}
		}
	})

	t.Run("succeeds when containing bytes is received before timeout", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsLogBytes(conn)

		// script
		conn.Send("logs")

		// assert
		rec.ToReceiveContaining("log")
		rec.ToReceiveContaining("log1")
		before := time.Now()
		rec.RunAssertions(300 * time.Millisecond) // it's a max
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("ToReceiveContaining should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 150*time.Millisecond {
				t.Errorf("ToReceiveContaining should succeed faster")
			}
		}
	})

	t.Run("fails when timeout occurs before containing message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"join", "room:1"})

		// assert
		rec.ToReceiveContaining("room:")
		rec.RunAssertions(75 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("ToReceiveContaining should fail because of timeout")
		}
	})

	t.Run("fails when no containing message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"join", "room:1"})

		// assert
		rec.ToReceiveContaining("notfound")
		rec.RunAssertions(300 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("ToReceiveContaining should fail because substr is not found")
		}
	})

	t.Run("fails when no containing string", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsLogStrings(conn)

		// script
		conn.Send("logs")

		// assert
		rec.ToReceiveContaining("notfound")
		rec.RunAssertions(300 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("ToReceiveContaining should fail because substr is not found")
		}
	})
}

func TestToBeClosed(t *testing.T) {
	t.Run("fails when conn is not closed", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		rec.ToBeClosed()

		// assert
		rec.RunAssertions(10 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("ToBeClosed should fail")
		}
	})

	t.Run("succeeds when conn is closed server-side", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"quit", ""})

		// assert
		rec.ToBeClosed()
		rec.RunAssertions(200 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("ToBeClosed should succeed because of serveWsHistory logic, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when conn is closed client-side", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"join", "room:1"})
		conn.Close()

		// assert
		rec.ToBeClosed()
		rec.RunAssertions(50 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("ToBeClosed should succeed because of explicit conn Close, mockT output is:", getTestOutput(mockT))
		}
	})
}

func TestMultiAssertion(t *testing.T) {
	t.Run("should succeed without blocking", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})
		rec.ToReceive(Message{"chat", "sentence1"})
		rec.ToReceive(Message{"chat", "sentence2"})
		rec.ToReceive(Message{"chat", "sentence3"})
		rec.ToReceive(Message{"chat", "sentence4"})

		// no assertion!
		rec.RunAssertions(100 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("several ToReceive should not fail")
		}
	})
}

func TestNoAssertion(t *testing.T) {
	t.Run("no assertion should succeed", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"join", "room:1"})

		// no assertion!
		rec.RunAssertions(10 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("NoAssertion can't fail")
		}
	})
}

// this test should be skipped, it's only there to inspect wsmock failing output
func TestFailing(t *testing.T) {
	t.Run("should fail", func(t *testing.T) {
		t.Skip()
		conn, rec := NewGorillaMockAndRecorder(t)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		rec.ToReceive(Message{"chat", "notfound"})
		rec.ToReceiveFirst(Message{"chat", "notfound"})
		rec.ToReceiveLast(Message{"chat", "notfound"})
		rec.NotToReceive(Message{"chat", "sentence1"})
		rec.ToReceiveContaining("notfound")
		rec.ToBeClosed()
		rec.ToReceiveSparseSequence([]any{Message{"chat", "notfound1"}, Message{"chat", "notfound2"}})
		rec.ToReceiveSequence([]any{Message{"chat", "notfound1"}, Message{"chat", "notfound2"}})
		rec.ToReceiveOnlySequence([]any{Message{"chat", "notfound1"}, Message{"chat", "notfound2"}})

		rec.RunAssertions(100 * time.Millisecond)
	})
}
