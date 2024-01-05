package integration_test

import (
	"testing"

	ws "github.com/silently/wsmock"
)

// this test should be skipped, it's only there to inspect wsmock failing output
func TestFailing(t *testing.T) {
	t.Run("should fail", func(t *testing.T) {
		t.Skip()
		conn, rec := ws.NewGorillaMockAndRecorder(t)

		go func() {
			conn.WriteJSON("1")
			conn.WriteJSON("2")
			conn.WriteJSON("3")
		}()

		// assert
		rec.NewAssertion().OneNotToBe(Message{"chat", "sentence1"})
		rec.NewAssertion().NextToBe(Message{"chat", "notfound"})
		rec.NewAssertion().OneToCheck(stringLongerThan3)
		rec.RunAssertions(10 * durationUnit)
	})
}
