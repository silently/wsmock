package wsmock

import (
	"log"
	"testing"
	"time"
)

func TestMultiConn_Chat(t *testing.T) {
	t.Run("succeeds when testing messages written before and after other users join", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		server := newChatServerStub()
		conn1, rec1 := NewGorillaMockAndRecorder(mockT)
		conn2, rec2 := NewGorillaMockAndRecorder(mockT)
		conn3, rec3 := NewGorillaMockAndRecorder(mockT)
		server.handle(conn1)
		server.handle(conn2)
		server.handle(conn3)

		conn1.Send(Message{"join", "user1"})
		rec1.AssertReceived(Message{"joined", "user1"})
		conn1.Send(Message{"message", "hello"})
		rec2.AssertNotReceived(Message{"user1", "hello"}) // user2 has not joined
		RunAssertions(mockT, 110*time.Millisecond)

		before := time.Now()
		conn2.Send(Message{"join", "user2"})
		conn3.Send(Message{"join", "user3"})
		conn3.Send(Message{"message", "hi"})
		rec1.AssertReceived(Message{"user3", "hi"})
		rec2.AssertReceived(Message{"user3", "hi"})
		log.Println(">>>>1", rec1.serverWrites)
		log.Println(">>>>2", rec2.serverWrites)
		rec3.AssertNotReceived(Message{"user3", "hi"})
		RunAssertions(mockT, 110*time.Millisecond)
		after := time.Now()
		log.Printf("> elapsed %+v", after.Sub(before))

		if mockT.Failed() { // fail not expected
			t.Error("unexpected messages in chat room")
		}
	})
}
