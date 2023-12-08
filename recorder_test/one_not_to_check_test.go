package recorder_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func TestOneNotToCheck_Success(t *testing.T) {
	t.Run("succeeds when valid message is received among others", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("long")
			conn.WriteJSON("longer")
			conn.WriteJSON("no")
		}()

		// assert
		rec.Assert().OneNotToCheck(stringLongerThan3)
		rec.RunAssertions(50 * time.Millisecond)

		if mockT.Failed() { // fail not expected
			t.Error("OneNotToCheck should succeed, mockT output is:", getTestOutput(mockT))
		}
	})
}

func TestOneNotToCheck_Failure(t *testing.T) {
	t.Run("fails when only invalid messages are received", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			time.Sleep(10 * time.Millisecond)
			conn.WriteJSON("long")
			conn.WriteJSON("longer")
			conn.WriteJSON("longest")
		}()

		rec.Assert().OneNotToCheck(stringLongerThan3)
		rec.RunAssertions(50 * time.Millisecond)

		if !mockT.Failed() { // fail expected
			t.Error("OneNotToCheck should fail because of unexpected message")
		}
	})
}
