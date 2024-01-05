package integration_test

import (
	"testing"

	ws "github.com/silently/wsmock"
)

func truthy(msg any) bool {
	switch v := msg.(type) {
	case string:
		return len(v) > 0
	case bool:
		return v
	case int:
		return v != 0
	default:
		return false
	}
}

func TestChainOneToBe(t *testing.T) {
	t.Run("succeeds when chain is an ordered subpart of messages", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		go func() {
			conn.WriteJSON(Message{"chat", "sentence1"})
			conn.WriteJSON(Message{"chat", "sentence2"})
			conn.WriteJSON(Message{"chat", "sentence3"})
		}()

		// declare expected chains
		rec.NewAssertion().
			OneToBe(Message{"chat", "sentence1"}).
			OneToBe(Message{"chat", "sentence2"})
		rec.NewAssertion().
			OneToBe(Message{"chat", "sentence1"}).
			OneToBe(Message{"chat", "sentence3"})
		rec.NewAssertion().
			OneToBe(Message{"chat", "sentence2"}).
			OneToBe(Message{"chat", "sentence3"})

		rec.RunAssertions(15 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("OneToBe*s* should succeed, mockT output is:\n", getTestOutput(mockT))
		}
	})
}

func TestChainVarious(t *testing.T) {
	t.Run("succeeds when chain1 fits pattern", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		go func() {
			conn.WriteJSON(Message{"chat", "sentence2"})
			conn.WriteJSON(Message{"chat", "sentence3"})
			conn.WriteJSON(Message{"chat", "sentence1"})
			conn.WriteJSON(Message{"chat", "sentence4"})
			conn.WriteJSON(Message{"chat", "sentence5"})
			conn.WriteJSON(Message{"chat", "sentence6"})
		}()

		// declare expected chains
		rec.NewAssertion().
			OneToBe(Message{"chat", "sentence2"}).
			NextToBe(Message{"chat", "sentence3"}).
			OneNotToBe(Message{"chat", "sentence3"}).
			NoneToBe(Message{"chat", "sentence1"})

		rec.RunAssertions(10 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("Chain1 should succeed, mockT output is:\n", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when chain2 fits `ToBe` pattern", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)
		conn.WriteJSON(Message{"chat", "sentence1"})
		conn.WriteJSON(Message{"chat", "sentence2"})
		conn.WriteJSON(Message{"chat", "sentence3"})
		conn.WriteJSON(Message{"chat", "sentence4"})
		conn.WriteJSON(Message{"chat", "sentence5"})
		conn.WriteJSON(Message{"chat", "sentence6"})

		// declare expected chains
		rec.NewAssertion().
			OneToBe(Message{"chat", "sentence2"}).
			NextToBe(Message{"chat", "sentence3"}).
			OneNotToBe(Message{"chat", "sentence3"}).
			NoneToBe(Message{"chat", "sentence1"})

		rec.RunAssertions(10 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("Chain2 should succeed, mockT output is:\n", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when chain3 fits `ToCheck` pattern", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		go func() {
			conn.WriteJSON(true)
			conn.WriteJSON(false)
			conn.WriteJSON(true)
			conn.WriteJSON(true)
		}()

		// declare expected chains
		rec.NewAssertion().
			NextToCheck(truthy).
			NextNotToCheck(truthy).
			AllToCheck(truthy)
		rec.NewAssertion().
			OneNotToCheck(truthy).
			LastToCheck(truthy)

		rec.RunAssertions(10 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("Chain3 should succeed, mockT output is:\n", getTestOutput(mockT))
		}
	})

	t.Run("succeeds when chain4 fits various patterns", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		go func() {
			conn.WriteJSON(true)
			conn.WriteJSON("ko")
			conn.WriteJSON(false)
			conn.WriteJSON("long")
			conn.WriteJSON("goooal")
			conn.WriteJSON("goooaal")
		}()

		// declare expected chains
		rec.NewAssertion().
			NextToCheck(truthy).
			NextNotToCheck(stringLongerThan3).
			OneToCheck(stringLongerThan3).
			AllToMatch(goalRE)
		rec.NewAssertion().
			OneNotToCheck(truthy).
			LastToCheck(stringLongerThan3)

		rec.RunAssertions(10 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("Chain4 should succeed, mockT output is:\n", getTestOutput(mockT))
		}
	})
}
