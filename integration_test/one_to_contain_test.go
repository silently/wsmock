package integration_test

import (
	"testing"
	"time"

	"github.com/gorilla/websocket"
	ws "github.com/silently/wsmock"
)

func TestOneToContain_Success(t *testing.T) {
	t.Run("succeeds fast when containing messages are received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("pong")
			conn.WriteJSON(Message{"nothing", "special"})
		}()

		// assert
		rec.NewAssertion().OneToContain("ong")
		rec.NewAssertion().OneToContain("spec")
		before := time.Now()
		rec.RunAssertions(10 * durationUnit)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("OneToContain should succeed, mockT output is:\n", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 50*time.Millisecond {
				t.Errorf("OneToContain should succeed faster")
			}
		}
	})

	t.Run("succeeds when containing message is received among others", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong2")
			conn.WriteJSON("pong3")
		}()

		// assert
		rec.NewAssertion().OneToContain("ng2")
		rec.RunAssertions(10 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("OneToContain should succeed, mockT output is:\n", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when testing JSON field names", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON(Message{"nothing", "special"})
		}()

		// assert
		rec.NewAssertion().OneToContain("kind") // json field is lowercased
		rec.NewAssertion().OneToContain("payload")
		rec.RunAssertions(5 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("OneToContain should succeed, mockT output is:\n", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when pointer to containing JSON message is written", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON(&Message{"pointer", "sent"})
		}()

		// assert
		rec.NewAssertion().OneToContain("kind") // json field is lowercased
		rec.NewAssertion().OneToContain("pointer")
		rec.NewAssertion().OneToContain("payload")
		rec.NewAssertion().OneToContain("sent")
		rec.RunAssertions(5 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("OneToContain should succeed, mockT output is:\n", getTestOutput(mockT))
		}
	})

	t.Run("succeeds fast when containing bytes is received", func(t *testing.T) {
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
		rec.NewAssertion().OneToContain("byte")
		rec.NewAssertion().OneToContain("byte2")
		before := time.Now()
		rec.RunAssertions(10 * durationUnit)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("OneToContain should succeed, mockT output is:\n", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 50*time.Millisecond {
				t.Errorf("OneToContain should succeed faster")
			}
		}
	})
}

func TestOneToContain_Failure(t *testing.T) {
	t.Run("fails when no message is received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// dumb script
		go conn.Send("ping")

		// assert
		rec.NewAssertion().OneToContain("pong")
		rec.RunAssertions(5 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("OneToContain should fail because no message is received")
		}
	})

	t.Run("fails when no containing message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON(Message{"nothing", "special"})
		}()

		// assert
		rec.NewAssertion().OneToContain("notfound")
		rec.RunAssertions(5 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("OneToContain should fail because substr is not found")
		}
	})

	t.Run("fails when no containing string", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("pong")
		}()

		// assert
		rec.NewAssertion().OneToContain("notfound")
		rec.RunAssertions(5 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("OneToContain should fail because substr is not found")
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
		rec.NewAssertion().OneToContain("notfound")
		rec.RunAssertions(5 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("OneToContain should fail because message can not be marshalled")
		}
	})

	t.Run("fails when timeout occurs before containing message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(6 * durationUnit)
			conn.WriteJSON("pong")
		}()

		// assert
		rec.NewAssertion().OneToContain("pong")
		rec.RunAssertions(5 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("OneToContain should fail because of timeout")
		}
	})

	t.Run("fails fast when conn is closed before message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(5 * durationUnit)
			conn.WriteJSON("pong")
		}()
		go func() {
			time.Sleep(2 * durationUnit)
			conn.Close()
		}()

		// assert
		rec.NewAssertion().OneToContain("ng")
		before := time.Now()
		rec.RunAssertions(10 * durationUnit)
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("OneToContain should fail because of Close")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 50*time.Millisecond {
				t.Error("OneToContain should fail faster because of Close")
			}
		}
	})
}
