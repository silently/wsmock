package wsmock

import (
	"sync"
	"testing"
)

var store recorderStore = recorderStore{sync.RWMutex{}, make(map[*testing.T]map[*Recorder]bool)}

type recorderStore struct {
	mu    sync.RWMutex
	index map[*testing.T]map[*Recorder]bool
}

func indexRecorder(t *testing.T, r *Recorder) {
	store.mu.Lock()
	defer store.mu.Unlock()

	if _, ok := store.index[t]; !ok {
		store.index[t] = make(map[*Recorder]bool)
	}

	store.index[t][r] = true
}

func getIndexedRecorders(t *testing.T) (recorders []*Recorder) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	for r := range store.index[t] {
		recorders = append(recorders, r)
	}

	return
}

func unindexRecorder(t *testing.T, r *Recorder) {
	store.mu.Lock()
	defer store.mu.Unlock()

	delete(store.index[t], r)
}
