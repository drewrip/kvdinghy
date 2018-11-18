package main

import (
	"encoding/json"
	"io"
	"sync"
	"github.com/hashicorp/raft"
)

type fsm struct {
	mutex      sync.Mutex
	KV	map[string]int
}

type event struct {
	Key string `json:"key"`
	Value int `json:"value"`
}

type keyval struct {
	Key string `json:"key"`
	Value int `json:"value"`
}

// Apply applies a Raft log entry to the key-value store.
func (fsm *fsm) Apply(logEntry *raft.Log) interface{} {
	var e event
	if err := json.Unmarshal(logEntry.Data, &e); err != nil {
		panic("Failed unmarshaling Raft log entry. This is a bug.")
	}

	fsm.mutex.Lock()
	defer fsm.mutex.Unlock()
	fsm.KV[e.Key] = e.Value

	return nil

}

func (fsm *fsm) Snapshot() (raft.FSMSnapshot, error) {
	fsm.mutex.Lock()
	defer fsm.mutex.Unlock()
	kvarray := []keyval{}
	for k,v := range fsm.KV{
		kvarray = append(kvarray, keyval{Key: k, Value: v})
	}
	return &fsmSnapshot{KV: kvarray}, nil
}

// Restore stores the key-value store to a previous state.
func (tfsm *fsm) Restore(serialized io.ReadCloser) error {
	var snapshot fsmSnapshot
	if err := json.NewDecoder(serialized).Decode(&snapshot); err != nil {
		return err
	}
	mapshot := fsm{}
	for _,i := range snapshot.KV{
		mapshot.KV[i.Key] = i.Value
	}
	tfsm.KV = mapshot.KV
	return nil
}
