package recorder_test

import (
	"fmt"
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func trueExceptOnEnd(end bool, _ any, _ []any) (done, passed bool, err string) {
	return true, !end, ""
}

func alwaysFalse(_ bool, _ any, _ []any) (done, passed bool, err string) {
	return true, false, ""
}

func hasMoreMessagesOnEndThan(count int) ws.AsserterFunc {
	return func(end bool, _ any, all []any) (done, passed bool, err string) {
		if end {
			err = fmt.Sprintf("on end, the number of messages should be strictly more than: %v", count)
			return true, len(all) > count, err
		}
		return
	}
}

func TestCustom_AlwaysTrue(t *testing.T) {
	t.Run("succeeds when custom Asserter does", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("pong")
		}()

		// assert
		rec.Assert().With(trueExceptOnEnd)
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("RunAssertions should have custom Asserter trueExceptOnEnd succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed >= 30*time.Millisecond {
				t.Error("RunAssertions should be faster with trueExceptOnEnd")
			}
		}
	})

	t.Run("fails when no message is received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(100 * time.Millisecond) // ------> longer than RunAssertions timeout
			conn.WriteJSON("pong")
		}()

		// assert
		rec.Assert().With(trueExceptOnEnd)
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("should have custom Asserter trueExceptOnEnd fail when no message is received")
		}
	})
}

func TestCustom_AlwaysFalse(t *testing.T) {
	t.Run("fails when custom Asserter does", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("pong")
		}()

		// assert
		rec.Assert().With(alwaysFalse)
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("RunAssertions should have custom finder alwaysFalse fail")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed >= 30*time.Millisecond {
				t.Error("RunAssertions should be faster with alwaysFalse")
			}
		}
	})
}

func TestCustom_CountMessages(t *testing.T) {
	t.Run("succeeds using hasMoreMessagesOnEndThan with enough messages", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong2")
			conn.WriteJSON("pong3")
			conn.WriteJSON("pong4")
		}()

		// assert
		rec.Assert().With(hasMoreMessagesOnEndThan(3))
		rec.RunAssertions(50 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("should have custom Asserter hasMoreMessagesOnEndThan succeed, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("fails using hasMoreMessagesOnEndThan when asked too soon", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong2")
			time.Sleep(60 * time.Millisecond)
			conn.WriteJSON("pong3")
			conn.WriteJSON("pong4")
		}()

		// assert
		rec.Assert().With(hasMoreMessagesOnEndThan(3))
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("should have custom Asserter hasMoreMessagesOnEndThan fail")
		}
	})

	t.Run("fails using hasMoreMessagesOnEndThan with not enough messages", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong2")
			conn.WriteJSON("pong3")
			conn.WriteJSON("pong4")
		}()

		// assert
		rec.Assert().With(hasMoreMessagesOnEndThan(10))
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("should have custom Asserter hasMoreMessagesOnEndThan fail")
		}
	})
}
