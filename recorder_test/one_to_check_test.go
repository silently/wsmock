package recorder_test

import (
	"testing"
	"time"

	ws "github.com/silently/wsmock"
)

func stringLongerThan3(msg any) bool {
	if str, ok := msg.(string); ok {
		return len(str) > 3
	}
	return false
}

func TestOneToCheck_Success(t *testing.T) {
	t.Run("succeeds when checkin message is received before timeout", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// script
		go func() {
			conn.WriteJSON("sentence")
		}()

		// assert
		rec.Assert().OneToCheck(stringLongerThan3)
		before := time.Now()
		rec.RunAssertions(100 * time.Millisecond)
		after := time.Now()

		if mockT.Failed() { // fail not expected
			t.Error("OneToCheck should succeed, mockT output is:", getTestOutput(mockT))
		} else {
			// test timing
			elapsed := after.Sub(before)
			if elapsed > 20*time.Millisecond {
				t.Errorf("OneToCheck should succeed faster")
			}
		}
	})
}
