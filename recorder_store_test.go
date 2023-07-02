package wsmock

import (
	"testing"
)

func TestStore(t *testing.T) {
	t.Run("TestStore size evolves with added/removed conns", func(t *testing.T) {
		mockT := &testing.T{}
		r1 := newRecorder(mockT)
		if len(store.index[mockT]) != 1 {
			t.Errorf("size: expected %v but got %v", 1, len(store.index[mockT]))
		}

		newRecorder(mockT)
		if len(store.index[mockT]) != 2 {
			t.Errorf("size: expected %v but got %v", 2, len(store.index[mockT]))
		}

		if len(getIndexedRecorders(mockT)) != 2 {
			t.Errorf("getConns size: expected %v but got %v", 2, len(store.index[mockT]))
		}

		unindexRecorder(mockT, r1)
		if len(store.index[mockT]) != 1 {
			t.Errorf("size: expected %v but got %v", 1, len(store.index[mockT]))
		}
	})
}
