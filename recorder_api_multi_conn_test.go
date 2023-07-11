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
		rec1.AssertReceived(Message{"joined", "user1"})
		rec2.AssertNotReceived(Message{"user1", "hello"}) // user2 has not joined
		Run(mockT, 110*time.Millisecond)

		// script
		conn2.Send(Message{"join", "user2"})
		time.Sleep(10 * time.Millisecond) // ensure user2 join is processed before user3's
		conn3.Send(Message{"join", "user3"})
		conn3.Send(Message{"message", "hi"})

		// assert
		rec1.AssertReceived(Message{"user3", "hi"})
		rec2.AssertReceived(Message{"user3", "hi"})
		rec3.AssertNotReceived(Message{"user3", "hi"})
		Run(mockT, 110*time.Millisecond)

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
		rec1.AssertReceived("lost")
		rec1.AssertNotReceived("won")
		rec2.AssertReceived("won")
		Run(mockT, 50*time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("unexpected messages in RPS game, mockT output is:", getTestOutput(mockT))
		}

		// script
		conn1.Send("scissors")
		conn2.Send("paper")

		// assert
		rec1.AssertReceived("won")
		rec1.AssertNotReceived("lost")
		rec2.AssertReceived("lost")
		Run(mockT, 50*time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("unexpected messages in RPS game, mockT output is:", getTestOutput(mockT))
		}

		// script
		conn1.Send("paper")
		conn2.Send("paper")

		// assert
		rec1.AssertReceived("draw")
		rec1.AssertNotReceived("won")
		rec1.AssertNotReceived("lost")
		rec2.AssertReceived("draw")
		Run(mockT, 50*time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("unexpected messages in RPS game, mockT output is:", getTestOutput(mockT))
		}
	})
}
