package recorder_test

import (
	"testing"
	"time"

	"github.com/gorilla/websocket"
	ws "github.com/silently/wsmock"
)

func TestOneToContain_Success(t *testing.T) {
	t.Run("succeeds fast when containing messages are received before timeout", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("pong")
			conn.WriteJSON(Message{"nothing", "special"})
		}()

		// assert
		rec.Assert().OneToContain("ong")
		rec.Assert().OneToContain("spec")
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("OneToContain should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 30*time.Millisecond {
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
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong2")
			conn.WriteJSON("pong3")
		}()

		// assert
		rec.Assert().OneToContain("ng2")
		rec.RunAssertions(100 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("OneToContain should succeed, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when testing JSON field names", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON(Message{"nothing", "special"})
		}()

		// assert
		rec.Assert().OneToContain("kind") // json field is lowercased
		rec.Assert().OneToContain("payload")
		rec.RunAssertions(50 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("OneToContain should succeed, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when pointer to containing JSON message is written", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON(&Message{"pointer", "sent"})
		}()

		// assert
		rec.Assert().OneToContain("kind") // json field is lowercased
		rec.Assert().OneToContain("pointer")
		rec.Assert().OneToContain("payload")
		rec.Assert().OneToContain("sent")
		rec.RunAssertions(50 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("OneToContain should succeed, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("succeeds fast when containing bytes is received before timeout", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			w, _ := conn.NextWriter(websocket.TextMessage)
			w.Write([]byte("byte1"))
			w.Close()
			w, _ = conn.NextWriter(websocket.TextMessage)
			w.Write([]byte("byte2"))
			w.Close()
		}()

		// assert
		rec.Assert().OneToContain("byte")
		rec.Assert().OneToContain("byte2")
		before := time.Now()
		rec.RunAssertions(50 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("OneToContain should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 30*time.Millisecond {
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
		rec.Assert().OneToContain("pong")
		rec.RunAssertions(50 * time.Millisecond)

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
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON(Message{"nothing", "special"})
		}()

		// assert
		rec.Assert().OneToContain("notfound")
		rec.RunAssertions(50 * time.Millisecond)

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
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("pong")
		}()

		// assert
		rec.Assert().OneToContain("notfound")
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("OneToContain should fail because substr is not found")
		}
	})

	t.Run("fails when timeout occurs before containing message", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(60 * time.Millisecond)
			conn.WriteJSON("pong")
		}()

		// assert
		rec.Assert().OneToContain("pong")
		rec.RunAssertions(50 * time.Millisecond)

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
			time.Sleep(50 * time.Millisecond)
			conn.WriteJSON("pong")
		}()
		go func() {
			time.Sleep(20 * time.Millisecond)
			conn.Close()
		}()

		// assert
		rec.Assert().OneToContain("ng")
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("OneToContain should fail because of Close")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 30*time.Millisecond {
				t.Error("OneToContain should fail faster because of Close")
			}
		}
	})
}
