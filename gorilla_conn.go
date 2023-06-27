package wsmock

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

type GorillaConn struct {
	recorder *Recorder
}

func NewGorillaMockAndRecorder(t *testing.T) (*GorillaConn, *Recorder) {
	recorder := NewRecorder(t)
	conn := &GorillaConn{recorder}

	return conn, recorder
}

// Stub API (used by server)

// blocking till next message sent from client
func (conn *GorillaConn) ReadJSON(m any) error {
	for {
		select {
		case read := <-conn.recorder.serverReadCh:
			b := read.(*bytes.Buffer)
			json.NewDecoder(b).Decode(m)

			return nil
		case <-conn.recorder.closedCh:
			return errors.New("[wsmock] conn closed while reading")
		}
	}
}

func (conn *GorillaConn) WriteJSON(m any) error {
	for {
		select {
		case conn.recorder.serverWriteCh <- m:
			return nil
		case <-conn.recorder.closedCh:
			return errors.New("[wsmock] conn closed while writing")
		}
	}
}

func (conn *GorillaConn) Close() error {
	if !conn.recorder.closed {
		conn.recorder.closed = true
		close(conn.recorder.closedCh)
	}
	return nil
}

// Client-side API

func (conn *GorillaConn) Send(m any) {
	w := &bytes.Buffer{}
	json.NewEncoder(w).Encode(m)
	conn.recorder.serverReadCh <- w
}

// IGorilla noop implementations

func (conn *GorillaConn) CloseHandler() func(code int, text string) error {
	return func(code int, text string) error {
		return conn.Close()
	}
}
func (conn *GorillaConn) EnableWriteCompression(enable bool) {}
func (conn *GorillaConn) LocalAddr() net.Addr {
	return &net.IPAddr{}
}
func (conn *GorillaConn) NextReader() (messageType int, r io.Reader, err error) {
	return 1, &io.PipeReader{}, nil
}
func (conn *GorillaConn) NextWriter(messageType int) (io.WriteCloser, error) {
	return &io.PipeWriter{}, nil
}
func (conn *GorillaConn) PingHandler() func(appData string) error {
	return func(appData string) error {
		return errors.New("")
	}
}
func (conn *GorillaConn) PongHandler() func(appData string) error {
	return func(appData string) error {
		return nil
	}
}
func (conn *GorillaConn) ReadMessage() (messageType int, p []byte, err error) {
	return 1, []byte{}, nil
}
func (conn *GorillaConn) RemoteAddr() net.Addr {
	return &net.IPAddr{}
}
func (conn *GorillaConn) SetCloseHandler(h func(code int, text string) error) {}
func (conn *GorillaConn) SetCompressionLevel(level int) error {
	return nil
}
func (conn *GorillaConn) SetPingHandler(h func(appData string) error) {}
func (conn *GorillaConn) SetPongHandler(h func(appData string) error) {}
func (conn *GorillaConn) SetReadDeadline(t time.Time) error {
	return nil
}
func (conn *GorillaConn) SetReadLimit(limit int64) {}
func (conn *GorillaConn) SetWriteDeadline(t time.Time) error {
	return nil
}
func (conn *GorillaConn) Subprotocol() string {
	return ""
}
func (conn *GorillaConn) UnderlyingConn() net.Conn {
	return &net.TCPConn{}
}
func (conn *GorillaConn) WriteControl(messageType int, data []byte, deadline time.Time) error {
	return nil
}
func (conn *GorillaConn) WriteMessage(messageType int, data []byte) error {
	return nil
}
func (conn *GorillaConn) WritePreparedMessage(pm *websocket.PreparedMessage) error {
	return nil
}
