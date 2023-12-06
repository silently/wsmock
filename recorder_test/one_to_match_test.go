package recorder_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	ws "github.com/silently/wsmock"
)

var goalRE *regexp.Regexp

func init() {
	goalRE, _ = regexp.Compile("g[oa]+l")
}

func TestOneToMatch_Success(t *testing.T) {
	t.Run("succeeds when matched string is received before timeout", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		go func() {
			conn.Send("shoot")
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("goooooal")
		}()

		// assert
		rec.Assert().OneToMatch(goalRE)
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("OneToMatch should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 30*time.Millisecond {
				t.Errorf("OneToMatch should succeed faster")
			}
		}
	})

	t.Run("succeeds when matched []byte is received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.Send("shoot")
			time.Sleep(10 * time.Millisecond)
			w, _ := conn.NextWriter(websocket.TextMessage)
			w.Write([]byte("goal"))
			w.Close()
		}()

		// assert
		rec.Assert().OneToMatch(goalRE)
		rec.RunAssertions(100 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("OneToMatch should succeed, mockT output is:", getTestOutput(mockT))
		}
	})
}

func TestOneToMatch_Failure(t *testing.T) {
	t.Run("fails when timeout occurs before matched message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		go func() {
			conn.Send("shoot")
			time.Sleep(50 * time.Millisecond)
			conn.WriteJSON("goooooal")
		}()

		// assert
		rec.Assert().OneToMatch(goalRE)
		rec.RunAssertions(30 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("OneToMatch should fail because of timeout")
		}
	})

	t.Run("fails when timeout occurs before matched message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		go func() {
			conn.Send("shoot")
			time.Sleep(50 * time.Millisecond)
			conn.WriteJSON("gl")
		}()

		// assert
		rec.Assert().OneToMatch(goalRE)
		rec.RunAssertions(30 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("OneToMatch should fail because there is no matching message")
		}
	})

	t.Run("fails when there is no", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.Send("ping")
		}()

		// assert
		rec.Assert().OneToMatch(goalRE)
		rec.RunAssertions(30 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("OneToMatch should fail because there is no received message")
		}
	})
}
