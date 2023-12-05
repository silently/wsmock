package recorder_test

import (
	"slices"
	"sync"
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

type rpsWsStub struct {
	sync.Mutex
	firstConn      *ws.GorillaConn
	firstConnThrow string
}

func newRpsWsStub() *rpsWsStub {
	return &rpsWsStub{sync.Mutex{}, nil, ""}
}

// -1 if throw1 is the winner, 1 for throw2, 0 for draw
func (s *rpsWsStub) decideWinner(throw1, throw2 string) int {
	if throw1 == throw2 {
		return 0
	}
	switch throw1 + throw2 {
	case "rockpaper":
		return 1
	case "rockscissors":
		return -1
	case "paperrock":
		return -1
	case "paperscissors":
		return 1
	case "scissorsrock":
		return 1
	case "scissorspaper":
		return -1
	default:
		return 0
	}
}

func (s *rpsWsStub) handle(conn *ws.GorillaConn) {
	throws := []string{"rock", "paper", "scissors"}
	go func() {
		for {
			var m string
			err := conn.ReadJSON(&m)
			if err != nil {
				// client left (or needs to stop loop anyway)
				return
			} else if slices.Contains(throws, m) {
				s.Lock()
				if s.firstConn == nil {
					s.firstConn = conn
					s.firstConnThrow = m
				} else {
					if s.firstConn == conn {
						break
					}
					switch s.decideWinner(s.firstConnThrow, m) {
					case -1:
						s.firstConn.WriteJSON("win")
						conn.WriteJSON("loss")
					case 1:
						s.firstConn.WriteJSON("loss")
						conn.WriteJSON("win")
					case 0:
						s.firstConn.WriteJSON("draw")
						conn.WriteJSON("draw")
					}
					s.firstConn = nil
					s.firstConnThrow = ""
				}
				s.Unlock()
			}
		}
	}()
}

func TestMulti_RPS(t *testing.T) {
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
