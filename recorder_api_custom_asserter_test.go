package wsmock

import (
	"fmt"
	"testing"
	"time"
)

func alwaysTrue(_ bool, _ any, _ []any) (done, passed bool, errorMessage string) {
	return true, true, ""
}

func alwaysFalse(_ bool, _ any, _ []any) (done, passed bool, errorMessage string) {
	return true, false, "[wsmock] always false"
}

func hasMoreMessagesOnEndThan(count int) Asserter {
	return func(end bool, _ any, allWrites []any) (done, passed bool, errorMessage string) {
		if end {
			errorMessage = fmt.Sprintf("[wsmock] on end, the number of messages should be strictly more than: %v", count)
			return true, len(allWrites) > count, errorMessage
		}
		return
	}
}

func TestAssertWith(t *testing.T) {
	t.Run("succeeds when custom Asserter does", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		rec.AssertWith(alwaysTrue)
		before := time.Now()
		rec.Run(100 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("AssertWith should have custom finder alwaysTrue succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed >= 30*time.Millisecond {
				t.Error("AssertWith should be faster with alwaysTrue")
			}
		}
	})

	t.Run("fails when custom Asserter does", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		rec.AssertWith(alwaysFalse)
		before := time.Now()
		rec.Run(100 * time.Millisecond)
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("AssertWith should have custom finder alwaysFalse fail")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed >= 30*time.Millisecond {
				t.Error("AssertWith should be faster with alwaysFalse")
			}
		}
	})

	t.Run("succeeds using hasMoreMessagesOnEndThan with enough messages", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		rec.AssertWith(hasMoreMessagesOnEndThan(3))
		rec.Run(70 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("should have custom Asserter hasMoreMessagesOnEndThan succeed, mockT output is:", getTestOutput(mockT))
		}
	})

	t.Run("fails using hasMoreMessagesOnEndThan when asked too soon", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		rec.AssertWith(hasMoreMessagesOnEndThan(3))
		rec.Run(20 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("should have custom Asserter hasMoreMessagesOnEndThan fail")
		}
	})

	t.Run("fails using hasMoreMessagesOnEndThan with not enough messages", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})

		// assert
		rec.AssertWith(hasMoreMessagesOnEndThan(10))
		rec.Run(70 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("should have custom Asserter hasMoreMessagesOnEndThan fail")
		}
	})

}
