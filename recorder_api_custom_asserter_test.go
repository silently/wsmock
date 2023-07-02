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

func hasMoreMessagesThan(count int) Asserter {
	return func(end bool, latestWrite any, allWrites []any) (done, passed bool, errorMessage string) {
		errorMessage = fmt.Sprintf("[wsmock] number of messages should be strictly more than: %v", count)
		return true, len(allWrites) > count, errorMessage
	}
}

func TestAssertWith(t *testing.T) {
	t.Run("with alwaysTrue custom Asserter", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})
		rec.AssertWith(alwaysTrue)
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("AssertWith should have custom finder alwaysTrue succeed")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed >= 30*time.Millisecond {
				t.Error("AssertWith should be faster with alwaysTrue")
			}
		}
	})

	t.Run("with alwaysFalse custom Asserter", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})
		rec.AssertWith(alwaysFalse)
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
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

	t.Run("with hasMoreMessagesThan custom Asserter", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})
		rec.AssertWith(hasMoreMessagesThan(3))
		rec.RunAssertions(70 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("should have custom Asserter hasMoreMessagesThan succeed")
		}
	})

	t.Run("with hasMoreMessagesThan custom Asserter, but too soon", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})
		rec.AssertWith(hasMoreMessagesThan(3))
		rec.RunAssertions(20 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("should have custom Asserter hasMoreMessagesThan fail")
		}
	})

}
