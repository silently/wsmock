package integration_test

import (
	"testing"
	"time"

	"github.com/gorilla/websocket"
	ws "github.com/silently/wsmock"
)

func TestOneNotToMatch_Success(t *testing.T) {
	t.Run("succeeds fast when non matching string is received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		go func() {
			conn.Send("shoot")
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("goal")
			conn.WriteJSON("missed")
		}()

		// assert
		rec.NewAssertion().OneNotToMatch(goalRE)
		before := time.Now()
		rec.RunAssertions(10 * durationUnit)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("OneNotToMatch should succeed, mockT output is:\n", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 5*durationUnit {
				t.Errorf("OneToMatch should succeed faster")
			}
		}
	})

	t.Run("succeeds when non matching []byte is received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.Send("shoot")
			time.Sleep(1 * durationUnit)
			w, _ := conn.NextWriter(websocket.TextMessage)
			w.Write([]byte("missed"))
			w.Close()
		}()

		// assert
		rec.NewAssertion().OneNotToMatch(goalRE)
		rec.RunAssertions(5 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("OneNotToMatch should succeed, mockT output is:\n", getTestOutput(mockT))
		}
	})
}

func TestOneNotToMatch_Failure(t *testing.T) {
	t.Run("fails when only matching message is received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		go func() {
			conn.Send("shoot")
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("goal")
		}()

		// assert
		rec.NewAssertion().OneNotToMatch(goalRE)
		rec.RunAssertions(5 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("OneNotToMatch should fail because there is no matching message")
		}
	})
}
