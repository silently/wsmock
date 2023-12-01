package wsmock

import (
	"testing"
	"time"
)

func TestMultiConn_Chat(t *testing.T) {
	t.Run("succeeds when testing messages written before and after other users join", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		server := newChatWsStub()
		conn1, rec1 := NewGorillaMockAndRecorder(mockT)
		conn2, rec2 := NewGorillaMockAndRecorder(mockT)
		conn3, rec3 := NewGorillaMockAndRecorder(mockT)
		server.handle(conn1)
		server.handle(conn2)
		server.handle(conn3)

		// script
		conn1.Send(Message{"join", "user1"})
		conn1.Send(Message{"message", "hello"})

		// assert
		rec1.ToReceive(Message{"joined", "user1"})
		rec2.NotToReceive(Message{"user1", "hello"}) // user2 has not joined
		RunAssertions(mockT, 110*time.Millisecond)

		// script
		conn2.Send(Message{"join", "user2"})
		time.Sleep(10 * time.Millisecond) // ensure user2 join is processed before user3's
		conn3.Send(Message{"join", "user3"})
		conn3.Send(Message{"message", "hi"})

		// assert
		rec1.ToReceive(Message{"user3", "hi"})
		rec2.ToReceive(Message{"user3", "hi"})
		rec3.NotToReceive(Message{"user3", "hi"})
		RunAssertions(mockT, 110*time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("unexpected messages in chat room, mockT output is:", getTestOutput(mockT))
		}
	})
}

func TestMultiConn_RPSt(t *testing.T) {
	t.Run("sends won/lost/draw to players", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		server := newRpsWsStub()
		conn1, rec1 := NewGorillaMockAndRecorder(mockT)
		conn2, rec2 := NewGorillaMockAndRecorder(mockT)
		server.handle(conn1)
		server.handle(conn2)

		// script
		conn1.Send("rock")
		conn2.Send("paper")

		// assert
		rec1.ToReceive("lost")
		rec1.NotToReceive("won")
		rec2.ToReceive("won")
		RunAssertions(mockT, 50*time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("unexpected messages in RPS game, mockT output is:", getTestOutput(mockT))
		}

		// script
		conn1.Send("scissors")
		conn2.Send("paper")

		// assert
		rec1.ToReceive("won")
		rec1.NotToReceive("lost")
		rec2.ToReceive("lost")
		RunAssertions(mockT, 50*time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("unexpected messages in RPS game, mockT output is:", getTestOutput(mockT))
		}

		// script
		conn1.Send("paper")
		conn2.Send("paper")

		// assert
		rec1.ToReceive("draw")
		rec1.NotToReceive("won")
		rec1.NotToReceive("lost")
		rec2.ToReceive("draw")
		RunAssertions(mockT, 50*time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("unexpected messages in RPS game, mockT output is:", getTestOutput(mockT))
		}
	})
}
