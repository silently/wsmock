package recorder_test

import (
	"time"

	"github.com/gorilla/websocket"

	w "github.com/silently/wsmock"
)

// stub for single conn tests
func serveWsHistory(conn w.IGorilla) {
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
		} else if m.Kind == "pointer" { // ~10ms
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON(&Message{"pointer", "sent"})
		}
	}
}

// stub for single conn tests
func serveWsLogStrings(conn w.IGorilla) {
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
}

func serveWsLogBytes(conn w.IGorilla) {
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
}
