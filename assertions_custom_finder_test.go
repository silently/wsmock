package wsmock

import (
	"testing"
	"time"
)

func alwaysTrue(messages []any) bool {
	return true
}

func alwaysFalse(messages []any) bool {
	return false
}

func lastMessage(expected any) Finder {
	return func(messages []any) bool {
		length := len(messages)
		if length == 0 {
			return false
		}
		return expected == messages[length-1]
	}
}

func moreMessagesThan(count int) Finder {
	return func(messages []any) bool {
		return len(messages) > count
	}
}

func TestAssertAfterTimeoutOrClose(t *testing.T) {
	t.Run("with alwaysTrue custom Finder", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})
		rec.AssertOnTimeoutOrClose("will be true", alwaysTrue)
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("should have custom finder alwaysTrue succeed")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed < 100*time.Millisecond {
				t.Error("AssertOnTimeoutOrClose should have waited for timeout")
			}
		}
	})

	t.Run("with alwaysFalse custom Finder", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})
		rec.AssertOnTimeoutOrClose("will be false", alwaysFalse)
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("should have custom finder alwaysFalse fail")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed < 100*time.Millisecond {
				t.Error("AssertOnTimeoutOrClose should have wait for timeout")
			}
		}
	})

	t.Run("with lastMessage custom Finder", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})
		rec.AssertOnTimeoutOrClose("wrong last message", lastMessage(Message{"chat", "sentence4"}))
		rec.RunAssertions(70 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("should have custom finder lastMessage succeed")
		}
	})

	t.Run("with lastMessage custom Finder, asserted too soon", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})
		rec.AssertOnTimeoutOrClose("wrong last message", lastMessage(Message{"chat", "sentence4"}))
		rec.RunAssertions(30 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("should have custom finder lastMessage fail")
		}
	})
}

func TestAssertOnWrite(t *testing.T) {
	t.Run("with alwaysTrue custom Finder", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})
		rec.AssertOnWrite("will be true", alwaysTrue)
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("should have custom finder alwaysTrue succeed")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 30*time.Millisecond {
				t.Error("AssertOnWrite should have been faster")
			}
		}
	})

	t.Run("with alwaysFalse custom Finder", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})
		rec.AssertOnWrite("will be true", alwaysFalse)
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("should have custom finder alwaysFalse fail")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed < 100*time.Millisecond {
				t.Error("AssertOnWrite should have waited for timeout")
			}
		}
	})

	t.Run("with moreMessagesThan custom Finder", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})
		rec.AssertOnWrite("not enough messages", moreMessagesThan(3))
		rec.RunAssertions(70 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("should have custom Finder moreMessagesThan succeed")
		}
	})

	t.Run("with moreMessagesThan custom Finder, but too late", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := NewGorillaMockAndRecorder(mockT)
		serveWsStub(conn)

		// script
		conn.Send(Message{"history", ""})
		rec.AssertOnWrite("not enough messages", moreMessagesThan(3))
		rec.RunAssertions(20 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("should have custom Finder moreMessagesThan fail")
		}
	})

}
