package wsmock

import (
	"testing"
	"time"
)

func TestMultiConn_Chat(t *testing.T) {
	t.Run("messages written before and after other users join", func(t *testing.T) {
		t.Skip() // TODO
		// init
		mockT := t //&testing.T{}
		server := newChatServerStub()
		conn1, rec1 := NewGorillaMockWithRecorder(mockT)
		conn2, rec2 := NewGorillaMockWithRecorder(mockT)
		conn3, rec3 := NewGorillaMockWithRecorder(mockT)
		server.handle(conn1)
		server.handle(conn2)
		server.handle(conn3)

		conn1.Send(Message{"join", "user1"})
		rec1.AssertReceived(Message{"joined", "user1"})
		conn1.Send(Message{"message", "hello"})
		rec2.AssertNotReceived(Message{"user1", "hello"}) // user2 has not joined
		RunAssertions(mockT, 100*time.Millisecond)

		conn2.Send(Message{"join", "user2"})
		conn3.Send(Message{"join", "user3"})
		conn3.Send(Message{"message", "hi"})
		rec1.AssertReceived(Message{"user3", "hi"})
		rec2.AssertReceived(Message{"user3", "hi"})
		rec3.AssertNotReceived(Message{"user3", "hi"})
		RunAssertions(mockT, 100*time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("unexpected messages in chat room")
		}
	})
}
