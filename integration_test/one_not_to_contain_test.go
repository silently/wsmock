package integration_test

import (
	"testing"
	"time"

	"github.com/gorilla/websocket"
	ws "github.com/silently/wsmock"
)

func TestOneNotToContain_Success(t *testing.T) {
	t.Run("succeeds when not containing message is received among others", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong2")
		}()

		// assert
		rec.NewAssertion().OneNotToContain("ng1")
		rec.RunAssertions(10 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("OneNotToContain should succeed, mockT output is:\n", getTestOutput(mockT))
		}
	})

	t.Run("succeeds fast when not containing bytes is received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			w, _ := conn.NextWriter(websocket.TextMessage)
			w.Write([]byte("byte1"))
			w.Close()
			w, _ = conn.NextWriter(websocket.TextMessage)
			w.Write([]byte("byte2"))
			w.Close()
		}()

		// assert
		rec.NewAssertion().OneNotToContain("byte1")
		before := time.Now()
		rec.RunAssertions(10 * durationUnit)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("OneNotToContain should succeed, mockT output is:\n", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 50*time.Millisecond {
				t.Errorf("OneNotToContain should succeed faster")
			}
		}
	})
}

func TestOneNotToContain_Failure(t *testing.T) {
	t.Run("fails when only containing message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON(Message{"nothing", "special"})
		}()

		// assert
		rec.NewAssertion().OneNotToContain("nothing")
		rec.RunAssertions(5 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("OneNotToContain should fail because substr is found")
		}
	})
}
