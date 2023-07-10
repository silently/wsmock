package wsmock

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	Kind    string `json:"kind"`
	Payload string `json:"payload"`
}

// stub for single conn tests
func serveWsStub(conn IGorilla) {
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
func serveWsStrings(conn IGorilla) {
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

func serveWsBytes(conn IGorilla) {
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
				if err := w.Close(); err != nil {
					return
				}
			}
		}
	}()
}

// stub for multi conn tests
type chatServerStub struct {
	sync.Mutex
	connIndex map[*GorillaConn]string
}

func newChatServerStub() *chatServerStub {
	return &chatServerStub{sync.Mutex{}, make(map[*GorillaConn]string)}
}

func (s *chatServerStub) handle(conn *GorillaConn) {
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
