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

		conn2.Send(Message{"join", "user2"})
		time.Sleep(10 * time.Millisecond) // ensure user2 join is processed before user3's
		conn3.Send(Message{"join", "user3"})

		conn3.Send(Message{"message", "hi"})
		rec1.AssertReceived(Message{"user3", "hi"})
		rec2.AssertReceived(Message{"user3", "hi"})
		rec3.AssertNotReceived(Message{"user3", "hi"})
		RunAssertions(mockT, 110*time.Millisecond)

		log.Println("rec1", rec1.serverWrites)
		log.Println("rec2", rec2.serverWrites)
		log.Println("rec3", rec3.serverWrites)

		if mockT.Failed() { // fail not expected
			t.Error("unexpected messages in chat room, mockT output is:", getTestOutput(mockT))
		}
	})
}
