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

// Interface satisfied both by Gorilla websocket.Conn and wsmock.GorillaConn,
// enabling the possibility to pass the latter in place of the former for testing purposes.
//
// It declares all methods available on https://pkg.go.dev/github.com/gorilla/websocket#Conn
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

// Mock for Gorilla websocket.Conn
type GorillaConn struct {
	serverReadCh chan any
	recorder     *Recorder
	closed       bool
	closedCh     chan struct{}
}

type gorillaWriteCloser struct {
	messageType int
	conn        *GorillaConn
	data        []byte
}

type gorillaReader struct {
	data      []byte
	readIndex int64
}

func (w *gorillaWriteCloser) Write(data []byte) (n int, err error) {
	w.data = append(w.data, data...)
	return len(data), nil
}

func (w *gorillaWriteCloser) Close() error {
	return w.conn.WriteMessage(w.messageType, w.data)
}

func (r *gorillaReader) Read(p []byte) (n int, err error) {
	if r.readIndex >= int64(len(r.data)) {
		err = io.EOF
		return
	}

	n = copy(p, r.data[r.readIndex:])
	r.readIndex += int64(n)
	return
}

// Returns a mock to be used in place of a Gorilla websocket.Conn in tests. GorillaConn has an extra Send()
// method to simulate client-side sent messages, and *Recorder provides an API to declare assertions about
// what is written by the server to the mock.
//
// Binding these resources to a given *testing.T helps cleaning them when the test is over.
func NewGorillaMockAndRecorder(t *testing.T) (*GorillaConn, *Recorder) {
	recorder := newRecorder(t)
	conn := &GorillaConn{
		serverReadCh: make(chan any, 256),
		recorder:     recorder,
		closedCh:     make(chan struct{}),
	}

	return conn, recorder
}

// Client-side API

// Send does not make any asumption on its message argument type (and does not serializes it),
// this will be decided upon what Read* function is used to retrieve it
func (conn *GorillaConn) Send(message any) {
	conn.serverReadCh <- message
}

// Stub API (used by server)

// Close the conn, preventing further reads or writes.
func (conn *GorillaConn) Close() error {
	if !conn.closed {
		conn.closed = true
		close(conn.closedCh)
		conn.recorder.stop()
	}
	return nil
}

// Parses as JSON the first message available on conn and stores the result in the value pointed to by v
// While waiting for it, it can return sooner if conn is closed
func (conn *GorillaConn) ReadJSON(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return errors.New("ReadJSON: argument should be a pointer")
	}
	for {
		select {
		case read := <-conn.serverReadCh:
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
		case read := <-conn.serverReadCh:
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
			return -1, nil, errors.New("[wsmock] conn closed while reading")
		}
	}
}

// Returns an io.Reader used to Read the next data message
func (conn *GorillaConn) NextReader() (messageType int, r io.Reader, err error) {
	messageType, p, err := conn.ReadMessage()
	r = &gorillaReader{p, 0}
	return
}

// Returns an io.WriteCloser used to Write the next data message
func (conn *GorillaConn) NextWriter(messageType int) (io.WriteCloser, error) {
	return &gorillaWriteCloser{messageType, conn, nil}, nil
}

// Writes the JSON encoding of m as a message to its recorder, but returns an error if conn is closed.
func (conn *GorillaConn) WriteJSON(m any) error {
	if conn.closed {
		return errors.New("[wsmock] conn closed while writing")
	}
	conn.recorder.serverWriteCh <- m
	return nil
}

// Writes a []byte msg to its recorder, but returns an error if conn is closed.
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

// IGorilla noop implementations

// Mock not implemented yet
func (conn *GorillaConn) CloseHandler() func(code int, text string) error {
	return func(code int, text string) error {
		return conn.Close()
	}
}

// Mock not implemented yet
func (conn *GorillaConn) EnableWriteCompression(enable bool) {}

// Mock not implemented yet
func (conn *GorillaConn) LocalAddr() net.Addr {
	return &net.IPAddr{}
}

// Mock not implemented yet
func (conn *GorillaConn) PingHandler() func(appData string) error {
	return func(appData string) error {
		return errors.New("")
	}
}

// Mock not implemented yet
func (conn *GorillaConn) PongHandler() func(appData string) error {
	return func(appData string) error {
		return nil
	}
}

// Mock not implemented yet
func (conn *GorillaConn) RemoteAddr() net.Addr {
	return &net.IPAddr{}
}

// Mock not implemented yet
func (conn *GorillaConn) SetCloseHandler(h func(code int, text string) error) {}

// Mock not implemented yet
func (conn *GorillaConn) SetCompressionLevel(level int) error {
	return nil
}

// Mock not implemented yet
func (conn *GorillaConn) SetPingHandler(h func(appData string) error) {}

// Mock not implemented yet
func (conn *GorillaConn) SetPongHandler(h func(appData string) error) {}

// Mock not implemented yet
func (conn *GorillaConn) SetReadDeadline(t time.Time) error {
	return nil
}

// Mock not implemented yet
func (conn *GorillaConn) SetReadLimit(limit int64) {}

// Mock not implemented yet
func (conn *GorillaConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// Mock not implemented yet
func (conn *GorillaConn) Subprotocol() string {
	return ""
}

// Mock not implemented yet
func (conn *GorillaConn) UnderlyingConn() net.Conn {
	return &net.TCPConn{}
}

// Mock not implemented yet
func (conn *GorillaConn) WriteControl(messageType int, data []byte, deadline time.Time) error {
	return nil
}

// Mock not implemented yet
func (conn *GorillaConn) WritePreparedMessage(pm *websocket.PreparedMessage) error {
	return nil
}
