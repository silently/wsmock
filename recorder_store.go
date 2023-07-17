package wsmock

import (
	"sync"
	"testing"
)

var store recorderStore = recorderStore{sync.RWMutex{}, make(map[*testing.T][]*Recorder)}

type recorderStore struct {
	mu    sync.RWMutex
	index map[*testing.T][]*Recorder
}

// returns the index/position of recorder for the given *testing.T test
func indexRecorder(t *testing.T, r *Recorder) (index int) {
	t.Helper()

	store.mu.Lock()
	defer store.mu.Unlock()

	length := len(store.index[t])

	if length == 0 { // do it once
		t.Cleanup(func() {
			unindexRecorders(t)
		})
	}

	store.index[t] = append(store.index[t], r)

	return length
}

func getIndexedRecorders(t *testing.T) (recorders []*Recorder) {
	t.Helper()

	store.mu.RLock()
	defer store.mu.RUnlock()

	return store.index[t]
}

func unindexRecorders(t *testing.T) {
	t.Helper()

	store.mu.Lock()
	defer store.mu.Unlock()

	delete(store.index, t)
}
