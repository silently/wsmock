package wsmock

import (
	"fmt"
	"testing"
	"time"
)

func alwaysTrue(_ bool, _ any, _ []any) (done, passed bool, errorMessage string) {
	return true, true, ""
}

func alwaysFalseWithEmptyError(_ bool, _ any, _ []any) (done, passed bool, errorMessage string) {
	return true, false, ""
}

func hasMoreMessagesOnEndThan(count int) AsserterFunc {
	return func(end bool, _ any, all []any) (done, passed bool, errorMessage string) {
		if end {
			errorMessage = fmt.Sprintf("on end, the number of messages should be strictly more than: %v", count)
			return true, len(all) > count, errorMessage
		}
		return
	}
}

func TestAssert(t *testing.T) {
	t.Run("succeeds when custom Asserter does", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		rec.AddAsserterFunc(alwaysTrue)
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("Assert should have custom finder alwaysTrue succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed >= 30*time.Millisecond {
				t.Error("Assert should be faster with alwaysTrue")
			}
		}
	})

	t.Run("fails when custom Asserter does", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		rec.AddAsserterFunc(alwaysFalseWithEmptyError)
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("Assert should have custom finder alwaysFalse fail")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed >= 30*time.Millisecond {
				t.Error("Assert should be faster with alwaysFalse")
			}
		}
	})

	t.Run("succeeds using hasMoreMessagesOnEndThan with enough messages", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		rec.AddAsserterFunc(hasMoreMessagesOnEndThan(3))
		rec.RunAssertions(70 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("should have custom Asserter hasMoreMessagesOnEndThan succeed, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("fails using hasMoreMessagesOnEndThan when asked too soon", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		rec.AddAsserterFunc(hasMoreMessagesOnEndThan(3))
		rec.RunAssertions(20 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("should have custom Asserter hasMoreMessagesOnEndThan fail")
		}
	})

	t.Run("fails using hasMoreMessagesOnEndThan with not enough messages", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsHistory(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		rec.AddAsserterFunc(hasMoreMessagesOnEndThan(10))
		rec.RunAssertions(70 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("should have custom Asserter hasMoreMessagesOnEndThan fail")
		}
	})

}
