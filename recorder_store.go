package wsmock

import (
	"sync"
	"testing"
)

var store recorderStore = recorderStore{sync.Mutex{}, make(map[*testing.T]map[*Recorder]bool)}

type recorderStore struct {
	sync.Mutex
	index map[*testing.T]map[*Recorder]bool
}

func indexRecorder(t *testing.T, r *Recorder) {
	t.Helper()

	store.Lock()
	defer store.Unlock()

	if _, ok := store.index[t]; !ok {
		store.index[t] = make(map[*Recorder]bool)
	}

	store.index[t][r] = true
}

func getIndexedRecorders(t *testing.T) (recorders []*Recorder) {
	t.Helper()

	store.Lock()
	defer store.Unlock()

	for r := range store.index[t] {
		recorders = append(recorders, r)
	}

	return
}

func unindexRecorder(t *testing.T, r *Recorder) {
	t.Helper()

	store.Lock()
	defer store.Unlock()

	delete(store.index[t], r)
}
