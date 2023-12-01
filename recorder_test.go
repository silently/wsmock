package wsmock

import (
	"testing"
	"time"
)

func TestNewRecorderCleanup(t *testing.T) {
	var subT *testing.T
	t.Run("name", func(t *testing.T) {
		subT = t
		_, rec := NewGorillaMockAndRecorder(subT)
		rec.RunAssertions(1 * time.Millisecond)
	})

	time.Sleep(1 * time.Second)
	t.Cleanup(func() {
		count := len(getIndexedRecorders(subT))
		if count > 0 {
			t.Error("there should be no indexed recorders when test is over")
		}
	})
}
