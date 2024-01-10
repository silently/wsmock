package integration_test

import (
	"fmt"
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func trueExceptOnEnd(end bool, _ any, _ []any) (done, passed bool, err string) {
	return true, !end, ""
}

func alwaysTrue(_ bool, _ any, _ []any) (done, passed bool, err string) {
	return true, true, ""
}

func alwaysFalse(_ bool, _ any, _ []any) (done, passed bool, err string) {
	return true, false, ""
}

func hasMoreMessagesOnEndThan(count int) ws.ConditionFunc {
	return func(end bool, _ any, all []any) (done, passed bool, err string) {
		if end {
			err = fmt.Sprintf("on end, the number of messages should be strictly more than: %v", count)
			return true, len(all) > count, err
		}
		return
	}
}

func TestPendingTooLate(t *testing.T) {
	t.Run("fails when last condition is not met", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		go conn.WriteJSON("hello")

		// declare expected chains
		rec.NewAssertion().
			OneToBe("hello").
			With(alwaysTrue).
			OneToBe("something")

		rec.RunAssertions(20 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("OneToBe*s* should not succeed since no message is sent")
		}
	})
}

func TestCustom_TrueExceptOnEnd(t *testing.T) {
	t.Run("succeeds when custom Condition does", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("pong")
		}()

		// assert
		rec.NewAssertion().With(trueExceptOnEnd)
		before := time.Now()
		rec.RunAssertions(10 * durationUnit)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("RunAssertions should have custom Condition trueExceptOnEnd succeed, mockT output is:\n", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed >= 5*durationUnit {
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
			time.Sleep(10 * durationUnit) // ------> longer than RunAssertions timeout
			conn.WriteJSON("pong")
		}()

		// assert
		rec.NewAssertion().With(trueExceptOnEnd)
		rec.RunAssertions(5 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("should have custom Condition trueExceptOnEnd fail when no message is received")
		}
	})
}

func TestCustom_AlwaysFalse(t *testing.T) {
	t.Run("fails when custom Condition does", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("pong")
		}()

		// assert
		rec.NewAssertion().With(alwaysFalse)
		before := time.Now()
		rec.RunAssertions(10 * durationUnit)
		after := time.Now()

		if !mockT.Failed() { // fail expected
			t.Error("RunAssertions should have custom finder alwaysFalse fail")
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed >= 3*durationUnit {
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
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong2")
			conn.WriteJSON("pong3")
			conn.WriteJSON("pong4")
		}()

		// assert
		rec.NewAssertion().With(hasMoreMessagesOnEndThan(3))
		rec.RunAssertions(5 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("should have custom Condition hasMoreMessagesOnEndThan succeed, mockT output is:\n", getTestOutput(mockT))
		}
	})

	t.Run("fails using hasMoreMessagesOnEndThan when asked too soon", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong2")
			time.Sleep(6 * durationUnit)
			conn.WriteJSON("pong3")
			conn.WriteJSON("pong4")
		}()

		// assert
		rec.NewAssertion().With(hasMoreMessagesOnEndThan(3))
		rec.RunAssertions(5 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("should have custom Condition hasMoreMessagesOnEndThan fail")
		}
	})

	t.Run("fails using hasMoreMessagesOnEndThan with not enough messages", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(1 * durationUnit)
			conn.WriteJSON("pong1")
			conn.WriteJSON("pong2")
			conn.WriteJSON("pong3")
			conn.WriteJSON("pong4")
		}()

		// assert
		rec.NewAssertion().With(hasMoreMessagesOnEndThan(10))
		rec.RunAssertions(5 * durationUnit)

		if !mockT.Failed() { // fail expected
			t.Error("should have custom Condition hasMoreMessagesOnEndThan fail")
		}
	})
}
