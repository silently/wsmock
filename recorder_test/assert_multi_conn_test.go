package recorder_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestMultiConn_Chat(t *testing.T) {
	t.Run("succeeds when testing messages written before and after other users join", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		server := newChatWsStub()
		conn1, rec1 := ws.NewGorillaMockAndRecorder(mockT)
		conn2, rec2 := ws.NewGorillaMockAndRecorder(mockT)
		conn3, rec3 := ws.NewGorillaMockAndRecorder(mockT)
		go server.handle(conn1)
		go server.handle(conn2)
		go server.handle(conn3)

		// script
		conn1.Send(Message{"join", "user1"})
		conn1.Send(Message{"message", "hello"})

		// assert
		rec1.Assert().OneToBe(Message{"joined", "user1"})
		rec2.Assert().OneNotToBe(Message{"user1", "hello"}) // user2 has not joined
		ws.RunAssertions(mockT, 110*time.Millisecond)

		// script
		conn2.Send(Message{"join", "user2"})
		time.Sleep(10 * time.Millisecond) // ensure user2 join is processed before user3's
		conn3.Send(Message{"join", "user3"})
		conn3.Send(Message{"message", "hi"})

		// assert
		rec1.Assert().OneToBe(Message{"user3", "hi"})
		rec2.Assert().OneToBe(Message{"user3", "hi"})
		rec3.Assert().OneNotToBe(Message{"user3", "hi"})
		ws.RunAssertions(mockT, 110*time.Millisecond)

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
		conn1, rec1 := ws.NewGorillaMockAndRecorder(mockT)
		conn2, rec2 := ws.NewGorillaMockAndRecorder(mockT)
		go server.handle(conn1)
		go server.handle(conn2)

		// script
		conn1.Send("paper")
		conn2.Send("rock")

		// assert
		rec1.Assert().OneToBe("win")
		rec2.Assert().OneToBe("loss")
		rec2.Assert().OneNotToBe("win")
		ws.RunAssertions(mockT, 50*time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("unexpected messages in RPS game, mockT output is:", getTestOutput(mockT))
		}

		// script
		conn1.Send("scissors")
		conn2.Send("paper")

		// assert
		rec1.Assert().OneToBe("win")
		rec1.Assert().OneNotToBe("loss")
		rec2.Assert().OneToBe("loss")
		ws.RunAssertions(mockT, 50*time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("unexpected messages in RPS game, mockT output is:", getTestOutput(mockT))
		}

		// script
		conn1.Send("paper")
		conn2.Send("paper")

		// assert
		rec1.Assert().OneToBe("draw")
		rec1.Assert().OneNotToBe("win")
		rec1.Assert().OneNotToBe("loss")
		rec2.Assert().OneToBe("draw")
		ws.RunAssertions(mockT, 50*time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("unexpected messages in RPS game, mockT output is:", getTestOutput(mockT))
		}
	})
}
