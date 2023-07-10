package wsmock

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestGorillaConnWrite(t *testing.T) {
	t.Run("NextWriter can write to recorder", func(t *testing.T) {
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)

		w, err := conn.NextWriter(websocket.TextMessage)
		if err != nil {
			t.Error(err)
		}
		w.Write([]byte("log1"))
		if err := w.Close(); err != nil {
			t.Error(err)
		}

		if len(rec.serverWriteCh) != 1 {
			t.Error("recorder should contain one write")
		}
	})

	t.Run("WriteMessage can write TextMessage to recorder", func(t *testing.T) {
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)

		err := conn.WriteMessage(websocket.TextMessage, []byte("message"))
		if err != nil {
			t.Error(err)
		}

		if len(rec.serverWriteCh) != 1 {
			t.Error("recorder should contain one write")
		}
	})

	t.Run("WriteMessage can write TextMessage to recorder", func(t *testing.T) {
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)

		err := conn.WriteMessage(websocket.BinaryMessage, []byte{0, 255})
		if err != nil {
			t.Error(err)
		}

		if len(rec.serverWriteCh) != 1 {
			t.Error("recorder should contain one write")
		}
	})

	t.Run("WriteMessage should error when conn is closed", func(t *testing.T) {
		mockT := &testing.T{}
		conn, _ := NewGorillaMockAndRecorder(mockT)
		conn.Close()

		err := conn.WriteMessage(websocket.TextMessage, []byte("message"))
		if err == nil || !strings.Contains(err.Error(), "conn closed") {
			t.Error("WriteMessage should return 'conn closed' error")
		}
	})
}

func TestGorillaConnRead(t *testing.T) {
	t.Run("ReadMessage can read BinaryMessage sent to conn", func(t *testing.T) {
		mockT := &testing.T{}
		conn, _ := NewGorillaMockAndRecorder(mockT)

		original := []byte("bytes")
		conn.Send(original)

		mType, message, err := conn.ReadMessage()
		if err != nil {
			t.Error(err)
		}
		if mType != websocket.BinaryMessage {
			t.Errorf("wrong message type, expected %v but got %v", websocket.BinaryMessage, mType)
		}
		if !bytes.Equal(message, original) {
			t.Errorf("wrong message, expected %v but got %v", original, message)
		}
	})

	t.Run("ReadMessage can read TextMessage sent to conn", func(t *testing.T) {
		mockT := &testing.T{}
		conn, _ := NewGorillaMockAndRecorder(mockT)

		original := "string"
		conn.Send(original)

		mType, bytes, err := conn.ReadMessage()
		if err != nil {
			t.Error(err)
		}
		message := string(bytes)
		if mType != websocket.TextMessage {
			t.Errorf("wrong message type, expected %v but got %v", websocket.TextMessage, mType)
		}
		if message != original {
			t.Errorf("wrong message, expected %v but got %v", original, message)
		}
	})

	t.Run("ReadMessage can read interface sent to conn", func(t *testing.T) {
		mockT := &testing.T{}
		conn, _ := NewGorillaMockAndRecorder(mockT)

		original := Message{"kind", "payload"}
		conn.Send(original)

		mType, bytes, err := conn.ReadMessage()
		if err != nil {
			t.Error(err)
		}
		var msg Message
		json.Unmarshal(bytes, &msg)
		if mType != websocket.TextMessage {
			t.Errorf("wrong message type, expected %v but got %v", websocket.TextMessage, mType)
		}
		if msg != original {
			t.Errorf("wrong message, expected %v but got %v", original, msg)
		}
	})

	t.Run("ReadMessage fails when interface can't be marshalled", func(t *testing.T) {
		mockT := &testing.T{}
		conn, _ := NewGorillaMockAndRecorder(mockT)

		var original complex64
		conn.Send(original)

		_, _, err := conn.ReadMessage()
		if err == nil {
			t.Error("ReadMessage should fail")
		}
	})

	t.Run("ReadMessage should error when conn is closed", func(t *testing.T) {
		mockT := &testing.T{}
		conn, _ := NewGorillaMockAndRecorder(mockT)
		conn.Close()

		mType, _, err := conn.ReadMessage()
		if err == nil || !strings.Contains(err.Error(), "conn closed") {
			t.Error("ReadMessage should return 'conn closed' error")
		}
		if mType != -1 {
			t.Errorf("wrong message type, expected %v but got %v", -1, mType)
		}
	})

	t.Run("ReadJSON fails when interface can't be marshalled", func(t *testing.T) {
		mockT := &testing.T{}
		conn, _ := NewGorillaMockAndRecorder(mockT)

		var original complex64
		conn.Send(original)

		var msg Message
		err := conn.ReadJSON(&msg)
		if err == nil {
			t.Error("ReadJSON unmarshalling should fail")
		}
	})

	t.Run("ReadJSON fails when argument is not a pointer", func(t *testing.T) {
		mockT := &testing.T{}
		conn, _ := NewGorillaMockAndRecorder(mockT)

		conn.Send(Message{"kind", "payload"})

		var msg Message
		err := conn.ReadJSON(msg)
		if err == nil {
			t.Error("ReadJSON fails with non pointer argument")
		}
	})
}
