package recorder_test

import (
	"sync"
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

type chatStub struct {
	sync.Mutex
	connIndex map[*ws.GorillaConn]string
}

func newChatWsStub() *chatStub {
	return &chatStub{sync.Mutex{}, make(map[*ws.GorillaConn]string)}
}

func (s *chatStub) handle(conn *ws.GorillaConn) {
	for {
		var m Message
		err := conn.ReadJSON(&m)
		if err != nil {
			// client left (or needs to stop loop anyway)
			return
		} else if m.Kind == "join" {
			s.Lock()
			s.connIndex[conn] = m.Payload
			s.Unlock()
			conn.WriteJSON(Message{"joined", m.Payload})
		} else if m.Kind == "message" {
			s.Lock()
			from := s.connIndex[conn]
			for c := range s.connIndex {
				if c != conn {
					c.WriteJSON(Message{from, m.Payload})
				}
			}
			s.Unlock()
		}
	}
}

func TestMulti_Chat(t *testing.T) {
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
		rec2.Assert().NoneToBe(Message{"user1", "hello"}) // user2 has not joined
		ws.RunAssertions(mockT, 110*time.Millisecond)

		// script
		conn2.Send(Message{"join", "user2"})
		time.Sleep(10 * time.Millisecond) // ensure user2 join is processed before user3's
		conn3.Send(Message{"join", "user3"})
		conn3.Send(Message{"message", "hi"})

		// assert
		rec1.Assert().OneToBe(Message{"user3", "hi"})
		rec2.Assert().OneToBe(Message{"user3", "hi"})
		rec3.Assert().NoneToBe(Message{"user3", "hi"})
		ws.RunAssertions(mockT, 110*time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("unexpected messages in chat room, mockT output is:", getTestOutput(mockT))
		}
	})
}
