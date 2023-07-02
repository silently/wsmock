package wsmock

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

type IGorilla interface {
	Close() error
	CloseHandler() func(code int, text string) error
	EnableWriteCompression(enable bool)
	LocalAddr() net.Addr
	NextReader() (messageType int, r io.Reader, err error)
	NextWriter(messageType int) (io.WriteCloser, error)
	PingHandler() func(appData string) error
	PongHandler() func(appData string) error
	ReadJSON(any) error
	ReadMessage() (messageType int, p []byte, err error)
	RemoteAddr() net.Addr
	SetCloseHandler(h func(code int, text string) error)
	SetCompressionLevel(level int) error
	SetPingHandler(h func(appData string) error)
	SetPongHandler(h func(appData string) error)
	SetReadDeadline(t time.Time) error
	SetReadLimit(limit int64)
	SetWriteDeadline(t time.Time) error
	Subprotocol() string
	UnderlyingConn() net.Conn
	WriteControl(messageType int, data []byte, deadline time.Time) error
	WriteJSON(any) error
	WriteMessage(messageType int, data []byte) error
	WritePreparedMessage(pm *websocket.PreparedMessage) error
}

type GorillaConn struct {
	recorder *Recorder
	closed   bool
	closedCh chan struct{}
}

type gorillaWriter struct {
	messageType int
	conn        *GorillaConn
	data        []byte
}

func (w *gorillaWriter) Write(data []byte) (n int, err error) {
	w.data = append(w.data, data...)
	return len(data), nil
}

func (w *gorillaWriter) Close() error {
	return w.conn.WriteMessage(w.messageType, w.data)
}

func NewGorillaMockAndRecorder(t *testing.T) (*GorillaConn, *Recorder) {
	recorder := newRecorder(t)
	conn := &GorillaConn{
		recorder: recorder,
		closedCh: make(chan struct{}),
	}

	return conn, recorder
}

// Client-side API

// Send does not make any asumption on its message argument type (and does not serializes it),
// this will be decided upon what Read* function is used to retrieve it
func (conn *GorillaConn) Send(message any) {
	conn.recorder.serverReadCh <- message
}

// Stub API (used by server)

// Parses as JSON the first message available on conn and stores the result in the value pointed to by v
// While waiting for it, it can return sooner if conn is closed
func (conn *GorillaConn) ReadJSON(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return errors.New("ReadJSON: argument should be a pointer")
	}
	for {
		select {
		case read := <-conn.recorder.serverReadCh:
			b, err := json.Marshal(read)
			if err != nil {
				return err
			}
			return json.Unmarshal(b, v)
		case <-conn.closedCh:
			return errors.New("[wsmock] conn closed while reading")
		}
	}
}

// Returns the first message available on conn, as []byte:
// - []byte message returned as is
// - string message converted to [byte]
// - other message types are JSON marshalled
// While waiting for a message, it can return sooner if conn is closed
func (conn *GorillaConn) ReadMessage() (messageType int, p []byte, err error) {
	for {
		select {
		case read := <-conn.recorder.serverReadCh:
			switch v := read.(type) {
			case []byte:
				return websocket.BinaryMessage, v, nil
			case string:
				return websocket.TextMessage, []byte(v), nil
			default:
				b, err := json.Marshal(read)
				if err != nil {
					return -1, nil, err
				}
				return websocket.TextMessage, b, nil
			}
		case <-conn.closedCh:
			return -1, nil, nil
		}
	}
}

func (conn *GorillaConn) WriteJSON(m any) error {
	if conn.closed {
		return errors.New("[wsmock] conn closed while writing")
	}
	conn.recorder.serverWriteCh <- m
	return nil
}

func (conn *GorillaConn) WriteMessage(messageType int, data []byte) error {
	if conn.closed {
		return errors.New("[wsmock] conn closed while writing")
	}
	if messageType == websocket.TextMessage {
		conn.recorder.serverWriteCh <- string(data)
	} else {
		conn.recorder.serverWriteCh <- data
	}
	return nil
}

func (conn *GorillaConn) NextWriter(messageType int) (io.WriteCloser, error) {
	return &gorillaWriter{messageType, conn, nil}, nil
}

func (conn *GorillaConn) Close() error {
	if !conn.closed {
		conn.closed = true
		close(conn.closedCh)
		conn.recorder.close()
	}
	return nil
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
func (conn *GorillaConn) WritePreparedMessage(pm *websocket.PreparedMessage) error {
	return nil
}
