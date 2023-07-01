package wsmock

import (
	"io"
	"net"
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

type Asserter func(latest any, messages []any) (done, result bool, errMessage string)

type Finder func(messages []any) bool
