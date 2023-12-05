package recorder_test

import (
	"slices"
	"sync"
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

// -1 if throw1 is the winner, 1 for throw2, 0 for draw
func decideWinner(throw1, throw2 string) int {
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

var throws = []string{"rock", "paper", "scissors"}

type rpsStub struct {
	sync.Mutex
	throwIndex map[*ws.GorillaConn][]string
}

func newRpsWsStub() *rpsStub {
	return &rpsStub{sync.Mutex{}, make(map[*ws.GorillaConn][]string)}
}

func (s *rpsStub) handle(conn *ws.GorillaConn) {
	for {
		var m string
		err := conn.ReadJSON(&m)
		if err != nil {
			// client left (or needs to stop loop anyway)
			return
		} else if slices.Contains(throws, m) {
			s.Lock()
			s.throwIndex[conn] = append(s.throwIndex[conn], m)
			for otherConn := range s.throwIndex { // iterate to find other conn
				if conn != otherConn { // found
					otherThrows := s.throwIndex[otherConn]
					otherLength := len(otherThrows)
					if len(s.throwIndex[conn]) == otherLength { // players have thrown the same amount of times
						switch decideWinner(m, otherThrows[otherLength-1]) {
						case -1:
							conn.WriteJSON("win")
							otherConn.WriteJSON("loss")
						case 1:
							conn.WriteJSON("loss")
							otherConn.WriteJSON("win")
						case 0:
							conn.WriteJSON("draw")
							otherConn.WriteJSON("draw")
						}
					}
				}
			}
			s.Unlock()
		}
	}
}

// Test similar to the first one described in README.md
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
		conn2.Send("paper")
		// then
		conn1.Send("paper")
		conn2.Send("rock")

		// assert
		rec1.Assert().OneToBe("win")
		rec2.Assert().OneToBe("loss")
		ws.RunAssertions(mockT, 50*time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("unexpected messages in RPS game, mockT output is:", getTestOutput(mockT))
		}
	})
}
