package tracker

import (
	"fmt"
	"sync"
	"time"

	"github.com/ConsenSys/orchestrate/pkg/types/tx"
)

type Tracker struct {
	// Output envelopes
	output map[string]chan *tx.Envelope

	// Envelope that can be diagnosed (last one retrieved from an out channel)
	Current *tx.Envelope

	mux *sync.RWMutex
}

func NewTracker() *Tracker {
	t := &Tracker{
		output: make(map[string]chan *tx.Envelope),
		mux:    &sync.RWMutex{},
	}
	return t
}

func (t *Tracker) AddOutput(key string, ch chan *tx.Envelope) {
	t.mux.Lock()
	defer t.mux.Unlock()
	t.output[key] = ch
}

func (t *Tracker) get(key string, timeout time.Duration) (*tx.Envelope, error) {
	t.mux.RLock()
	defer t.mux.RUnlock()
	ch, ok := t.output[key]
	if !ok {
		return nil, fmt.Errorf("output %q not tracked", key)
	}

	select {
	case e := <-ch:
		return e, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("no envelope available in output %q. Timeout after %s", key, timeout)
	}
}

func (t *Tracker) Load(key string, timeout time.Duration) error {
	e, err := t.get(key, timeout)
	if err != nil {
		return err
	}

	// Set Current envelope
	t.Current = e

	return nil
}
