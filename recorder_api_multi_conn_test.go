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
		rec1.OneToBe(Message{"joined", "user1"})
		rec2.OneNotToBe(Message{"user1", "hello"}) // user2 has not joined
		RunChecks(mockT, 110*time.Millisecond)

		// script
		conn2.Send(Message{"join", "user2"})
		time.Sleep(10 * time.Millisecond) // ensure user2 join is processed before user3's
		conn3.Send(Message{"join", "user3"})
		conn3.Send(Message{"message", "hi"})

		// assert
		rec1.OneToBe(Message{"user3", "hi"})
		rec2.OneToBe(Message{"user3", "hi"})
		rec3.OneNotToBe(Message{"user3", "hi"})
		RunChecks(mockT, 110*time.Millisecond)

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
		conn1.Send("paper")
		conn2.Send("rock")

		// assert
		rec1.OneToBe("won")
		rec2.OneToBe("lost")
		rec2.OneNotToBe("won")
		RunChecks(mockT, 50*time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("unexpected messages in RPS game, mockT output is:", getTestOutput(mockT))
		}

		// script
		conn1.Send("scissors")
		conn2.Send("paper")

		// assert
		rec1.OneToBe("won")
		rec1.OneNotToBe("lost")
		rec2.OneToBe("lost")
		RunChecks(mockT, 50*time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("unexpected messages in RPS game, mockT output is:", getTestOutput(mockT))
		}

		// script
		conn1.Send("paper")
		conn2.Send("paper")

		// assert
		rec1.OneToBe("draw")
		rec1.OneNotToBe("won")
		rec1.OneNotToBe("lost")
		rec2.OneToBe("draw")
		RunChecks(mockT, 50*time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("unexpected messages in RPS game, mockT output is:", getTestOutput(mockT))
		}
	})
}
