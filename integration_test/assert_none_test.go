package integration_test

import (
	"testing"

	ws "github.com/silently/wsmock"
)

func TestNoRunAssertionsion(t *testing.T) {
	t.Run("no assertion should succeed", func(t *testing.T) {
		// init
		mockT := &testing.T{}
		conn, rec := ws.NewGorillaMockAndRecorder(mockT)

		// dumb script
		go conn.Send("ping")

		// no assertion!
		rec.RunAssertions(1 * durationUnit)

		if mockT.Failed() { // fail not expected
			t.Error("NoAssertion can't fail")
		}
	})
}
