package wsmock

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

type GorillaMock struct {
	recorder *Recorder
}

func NewGorillaMockWithRecorder(t *testing.T) (*GorillaMock, *Recorder) {
	recorder := NewRecorder(t)
	conn := &GorillaMock{recorder}

	return conn, recorder
}

// Stub API (used by server)

// blocking till next message sent from client
func (conn *GorillaMock) ReadJSON(m any) error {
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

func (conn *GorillaMock) WriteJSON(m any) error {
	for {
		select {
		case conn.recorder.serverWriteCh <- m:
			return nil
		case <-conn.recorder.closedCh:
			return errors.New("[wsmock] conn closed while writing")
		}
	}
}

func (conn *GorillaMock) Close() error {
	if !conn.recorder.closed {
		conn.recorder.closed = true
		close(conn.recorder.closedCh)
	}
	return nil
}

// Client-side API

func (conn *GorillaMock) Send(m any) {
	w := &bytes.Buffer{}
	json.NewEncoder(w).Encode(m)
	conn.recorder.serverReadCh <- w
}
