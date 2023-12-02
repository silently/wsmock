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

func TestOneToBe(t *testing.T) {
	t.Run("succeeds when message is received before timeout", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"join", "room:1"})

		// assert
		rec.OneToBe(Message{"joined", "room:1"})
		before := time.Now()
		rec.RunAssertions(300 * time.Millisecond) // it's a max
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("OneToBe should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 150*time.Millisecond {
				t.Errorf("OneToBe should succeed faster")
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
		rec.OneToBe(Message{"joined", "room:1"})
		rec.RunAssertions(75 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("OneToBe should fail because of timeout")
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
		rec.OneToBe(Message{"joined", "room:1"})
		before := time.Now()
		rec.RunAssertions(200 * time.Millisecond)
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("OneToBe should fail because of Close")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 100*time.Millisecond {
				t.Error("OneToBe should fail faster because of Close")
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
		rec.OneToBe(target)
		rec.RunAssertions(75 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("OneToBe should fail because of unexpected message")
		}
		// assert error message
		expectedErrorMessage := fmt.Sprintf("message not received\nexpected: %+v", target)
		processedErrorMessage := removeSpaces(expectedErrorMessage)
		processedActualErrorMessage := removeSpaces(getTestOutput(mockT))
		if !strings.Contains(processedActualErrorMessage, processedErrorMessage) {
			t.Errorf("OneToBe wrong error message, expected:\n\"%v\"", expectedErrorMessage)
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
		rec.OneToBe(Message{"chat", "sentence1"})
		rec.OneToBe(Message{"chat", "sentence1"}) // twice

		rec.RunAssertions(50 * time.Millisecond) // it's a max

		if mockT.Failed() { // fail not expected
			t.Error("OneToBe should succeed, mockT output is:", getTestOutput(mockT))
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
		rec.OneToBe(Message{"chat", "sentence1"})
		rec.RunAssertions(50 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("OneToBe should succeed, mockT output is:", getTestOutput(mockT))
		}

		rec.OneToBe(Message{"chat", "sentence1"})
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("OneToBe should fail for second Run", getTestOutput(mockT))
		}
	})
}

func TestOneToBeOneNotTobe(t *testing.T) {
	t.Run("should fail fast", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		rec.OneToBe(Message{"chat", "sentence4"})
		rec.OneNotToBe(Message{"chat", "sentence1"}) // failing assertion
		before := time.Now()
		rec.RunAssertions(300 * time.Millisecond) // it's a max
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("OneNotToBe should fail")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 100*time.Millisecond {
				t.Errorf("OneNotToBe should fail faster")
			}
		}
	})
}

func TestFirstToBe(t *testing.T) {
	t.Run("succeeds when first message is received before timeout", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		rec.FirstToBe(Message{"chat", "sentence1"})
		before := time.Now()
		rec.RunAssertions(300 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("FirstToBe should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 40*time.Millisecond {
				t.Error("FirstToBe should succeed faster")
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
		rec.FirstToBe(Message{"chat", "sentence1"})
		rec.RunAssertions(5 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("FirstToBe should fail because of timeout")
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
		rec.FirstToBe(Message{"chat", "sentence2"})
		rec.RunAssertions(20 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("FirstToBe should fail")
			// } else {
			// 	output := getTestOutput(mockT)
			// 	if !strings.Contains(output, "should be: {chat sentence2}, received: {chat sentence1}") {
			// 		t.Errorf("FirstToBe unexpected error message: \"%v\"", output)
			// 	}
		}
	})
}

func TestLastToBe(t *testing.T) {
	t.Run("succeeds when last message is received before timeout", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		rec.LastToBe(Message{"chat", "sentence4"})
		before := time.Now()
		rec.RunAssertions(300 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("LastToBe should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed < 300*time.Millisecond {
				t.Error("LastToBe should not succeed before timeout")
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
		rec.LastToBe(Message{"chat", "sentence4"})
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("LastToBe should fail because of timeout")
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
		rec.LastToBe(Message{"chat", "sentence5"})
		rec.RunAssertions(300 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("LastToBe should fail")
			// } else {
			// 	output := getTestOutput(mockT)
			// 	if !strings.Contains(output, "should be: {chat sentence5}, received: {chat sentence4}") {
			// 		t.Errorf("LastToBe unexpected error message: \"%v\"", output)
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
		rec.LastToBe(Message{"chat", "sentence1"})
		rec.RunAssertions(100 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("LastToBe should fail")
			// } else {
			// 	output := getTestOutput(mockT)
			// 	if !strings.Contains(output, "should be: {chat sentence1}, received none") {
			// 		t.Errorf("LastToBe unexpected error message: \"%v\"", output)
			// 	}
		}
	})
}

func TestOneNotToBe(t *testing.T) {
	t.Run("succeeds when message is not received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"join", "room:1"})

		// assert
		rec.OneNotToBe(Message{"not", "planned"})
		rec.RunAssertions(110 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("OneNotToBe should succeed, mockT output is:", getTestOutput(mockT))
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
		rec.OneNotToBe(Message{"joined", "room:1"})
		rec.RunAssertions(110 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("OneNotToBe should fail (message is received)")
		}
	})
}

func TestOneToContain(t *testing.T) {
	t.Run("succeeds when containing message is received before timeout", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"join", "room:1"})

		// assert
		rec.OneToContain("room:")
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
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"join", "room:1"})

		// assert
		rec.OneToContain("kind") // json field is lowercased
		rec.OneToContain("joined")
		rec.RunAssertions(200 * time.Millisecond) // it's a max

		if mockT.Failed() { // fail not expected
			t.Error("OneToContain should succeed, mockT output is:", getTestOutput(mockT))
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
		rec.OneToContain("kind") // json field is lowercased
		rec.OneToContain("pointer")
		rec.OneToContain("payload")
		rec.OneToContain("sent")
		rec.RunAssertions(200 * time.Millisecond) // it's a max

		if mockT.Failed() { // fail not expected
			t.Error("OneToContain should succeed, mockT output is:", getTestOutput(mockT))
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
		rec.OneToContain("log")
		rec.OneToContain("log1")
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
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsLogBytes(conn)

		// script
		conn.Send("logs")

		// assert
		rec.OneToContain("log")
		rec.OneToContain("log1")
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
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"join", "room:1"})

		// assert
		rec.OneToContain("room:")
		rec.RunAssertions(75 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("OneToContain should fail because of timeout")
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
		rec.OneToContain("notfound")
		rec.RunAssertions(300 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("OneToContain should fail because substr is not found")
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
		rec.OneToContain("notfound")
		rec.RunAssertions(300 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("OneToContain should fail because substr is not found")
		}
	})
}

func TestConnClosed(t *testing.T) {
	t.Run("fails when conn is not closed", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		rec.ConnClosed()

		// assert
		rec.RunAssertions(10 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("ConnClosed should fail")
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
		rec.ConnClosed()
		rec.RunAssertions(200 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("ConnClosed should succeed because of serveWsHistory logic, mockT output is:", getTestOutput(mockT))
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
		rec.ConnClosed()
		rec.RunAssertions(50 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("ConnClosed should succeed because of explicit conn Close, mockT output is:", getTestOutput(mockT))
		}
	})
}

func TestMultiRunAssertionsion(t *testing.T) {
	t.Run("should succeed without blocking", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})
		rec.OneToBe(Message{"chat", "sentence1"})
		rec.OneToBe(Message{"chat", "sentence2"})
		rec.OneToBe(Message{"chat", "sentence3"})
		rec.OneToBe(Message{"chat", "sentence4"})

		// no assertion!
		rec.RunAssertions(100 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("several OneToBe should not fail")
		}
	})
}

func TestNoRunAssertionsion(t *testing.T) {
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
		rec.OneToBe(Message{"chat", "notfound"})
		rec.FirstToBe(Message{"chat", "notfound"})
		rec.LastToBe(Message{"chat", "notfound"})
		rec.OneNotToBe(Message{"chat", "sentence1"})
		rec.OneToContain("notfound")
		rec.ConnClosed()

		rec.RunAssertions(100 * time.Millisecond)
	})
}
