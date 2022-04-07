package testutils

import (
	"fmt"
	"sync"
)

// ChanRegistry holds a set of indexed envelopes channels
// and allows to dispatch envelope in those channels
type chanRegistry struct {
	mux *sync.RWMutex

	chans map[string]chan interface{}
}

// newChanRegistry creates a new channel registry
func newChanRegistry() *chanRegistry {
	return &chanRegistry{
		mux:   &sync.RWMutex{},
		chans: make(map[string]chan interface{}),
	}
}

// Register register a new channel
func (r *chanRegistry) Register(key string, ch chan interface{}) {
	r.mux.Lock()
	defer r.mux.Unlock()

	r.chans[key] = ch
}

// HasChan returns whether a channel is registered for the given key
func (r *chanRegistry) HasChan(key string) bool {
	r.mux.RLock()
	defer r.mux.RUnlock()

	_, ok := r.chans[key]

	return ok
}

// HasChan returns whether a channel is registered for the given key
func (r *chanRegistry) GetChan(key string) chan interface{} {
	r.mux.RLock()
	defer r.mux.RUnlock()

	ch, ok := r.chans[key]
	if !ok {
		return nil
	}

	return ch
}

// Send envelope to channel registered for key
func (r *chanRegistry) Send(key string, e interface{}) error {
	r.mux.RLock()
	defer r.mux.RUnlock()

	ch, ok := r.chans[key]
	if !ok {
		return fmt.Errorf("no channel register for key %q", key)
	}

	// Send envelope into channel
	ch <- e

	return nil
}
