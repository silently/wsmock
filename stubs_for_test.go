package wsmock

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/exp/slices"
)

type Message struct {
	Kind    string `json:"kind"`
	Payload string `json:"payload"`
}

// stub for single conn tests
func serveWsHistory(conn IGorilla) {
	go func() {
		for {
			var m Message
			err := conn.ReadJSON(&m)
			if err != nil {
				// client left (or needs to stop loop anyway)
				return
			} else if m.Kind == "join" { // ~100ms
				time.Sleep(100 * time.Millisecond)
				conn.WriteJSON(Message{"joined", m.Payload})
			} else if m.Kind == "history" { // ~60ms
				time.Sleep(10 * time.Millisecond)
				conn.WriteJSON(Message{"chat", "sentence1"})
				time.Sleep(20 * time.Millisecond)
				conn.WriteJSON(Message{"chat", "sentence2"})
				time.Sleep(10 * time.Millisecond)
				conn.WriteJSON(Message{"chat", "sentence3"})
				time.Sleep(20 * time.Millisecond)
				conn.WriteJSON(Message{"chat", "sentence4"})
			} else if m.Kind == "quit" { // ~10ms
				time.Sleep(10 * time.Millisecond)
				conn.Close()
			}
		}
	}()
}

// stub for single conn tests
func serveWsLogStrings(conn IGorilla) {
	go func() {
		for {
			var m string
			err := conn.ReadJSON(&m)
			if err != nil {
				// client left (or needs to stop loop anyway)
				return
			} else if m == "logs" {
				conn.WriteJSON("log1")
				conn.WriteJSON("log2")
				conn.WriteJSON("log3")
				conn.WriteJSON("log4")
			}
		}
	}()
}

func serveWsLogBytes(conn IGorilla) {
	go func() {
		for {
			var m string
			err := conn.ReadJSON(&m)
			if err != nil {
				// client left (or needs to stop loop anyway)
				return
			} else if m == "logs" {
				w, err := conn.NextWriter(websocket.TextMessage)
				if err != nil {
					return
				}
				w.Write([]byte("log1"))
				w.Write([]byte("log2"))
				if err := w.Close(); err != nil {
					return
				}
			}
		}
	}()
}

// stub for multi conn tests
type chatWsStub struct {
	sync.Mutex
	connIndex map[*GorillaConn]string
}

func newChatWsStub() *chatWsStub {
	return &chatWsStub{sync.Mutex{}, make(map[*GorillaConn]string)}
}

func (s *chatWsStub) handle(conn *GorillaConn) {
	go func() {
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
	}()
}

// rps stub for multi conn tests
type rpsWsStub struct {
	sync.Mutex
	firstConn      *GorillaConn
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

func (s *rpsWsStub) handle(conn *GorillaConn) {
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
						s.firstConn.WriteJSON("won")
						conn.WriteJSON("lost")
					case 1:
						s.firstConn.WriteJSON("lost")
						conn.WriteJSON("won")
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
