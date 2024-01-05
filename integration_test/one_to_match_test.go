package integration_test

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
	t.Run("succeeds fast when matching string is received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		go func() {
			conn.Send("shoot")
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("goooooal")
		}()

		// assert
		rec.NewAssertion().OneToMatch(goalRE)
		before := time.Now()
		rec.RunAssertions(10 * durationUnit)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("OneToMatch should succeed, mockT output is:\n", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 50*time.Millisecond {
				t.Errorf("OneToMatch should succeed faster")
			}
		}
	})

	t.Run("succeeds fast when matching string is received among others", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		go func() {
			conn.Send("shoot")
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("missed")
			conn.WriteJSON("gooooal")
			conn.WriteJSON("out")
		}()

		// assert
		rec.NewAssertion().OneToMatch(goalRE)
		rec.RunAssertions(5 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("OneToMatch should succeed, mockT output is:\n", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when matching []byte is received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.Send("shoot")
			time.Sleep(1 * durationUnit)
			w, _ := conn.NextWriter(websocket.TextMessage)
			w.Write([]byte("goal"))
			w.Close()
		}()

		// assert
		rec.NewAssertion().OneToMatch(goalRE)
		rec.RunAssertions(5 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("OneToMatch should succeed, mockT output is:\n", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when matching marshalled message is received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.Send("shoot")
			time.Sleep(1 * durationUnit)
			conn.WriteJSON(Message{"event", "gooal"})
		}()

		// assert
		rec.NewAssertion().OneToMatch(goalRE)
		rec.RunAssertions(5 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("OneToMatch should succeed, mockT output is:\n", getTestOutput(mockT))
		}
	})
}

func TestOneToMatch_Failure(t *testing.T) {
	t.Run("fails when no message is received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// dumb script
		go conn.Send("ping")

		// assert
		rec.NewAssertion().OneToMatch(goalRE)
		rec.RunAssertions(5 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("OneToMatch should fail because no message is received")
		}
	})

	t.Run("fails when no matching message is received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		go func() {
			conn.Send("shoot")
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("missed")
		}()

		// assert
		rec.NewAssertion().OneToMatch(goalRE)
		rec.RunAssertions(5 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("OneToMatch should fail because there is no matching message")
		}
	})

	t.Run("fails when message can not be marshalled", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON(make(chan bool)) // contrieved example for test coverage
		}()

		// assert
		rec.NewAssertion().OneToMatch(goalRE)
		rec.RunAssertions(5 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("OneToMatch should fail because message can not be marshalled")
		}
	})

	t.Run("fails when timeout occurs before matching message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		go func() {
			conn.Send("shoot")
			time.Sleep(6 * durationUnit)
			conn.WriteJSON("goooooal")
		}()

		// assert
		rec.NewAssertion().OneToMatch(goalRE)
		rec.RunAssertions(5 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("OneToMatch should fail because of timeout")
		}
	})

	t.Run("fails fast when conn is closed before message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.Send("shoot")
			time.Sleep(5 * durationUnit)
			conn.WriteJSON("pong")
		}()
		go func() {
			time.Sleep(2 * durationUnit)
			conn.Close()
		}()

		// assert
		rec.NewAssertion().OneToMatch(goalRE)
		before := time.Now()
		rec.RunAssertions(10 * durationUnit)
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("OneToMatch should fail because of Close")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 50*time.Millisecond {
				t.Error("OneToMatch should fail faster because of Close")
			}
		}
	})
}
