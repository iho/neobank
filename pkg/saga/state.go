package saga

import "sync"

// State is a thread-safe key-value bag passed between saga steps.
type State struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewState(initial map[string]string) *State {
	data := make(map[string]string, len(initial))
	for k, v := range initial {
		data[k] = v
	}
	return &State{data: data}
}

func (s *State) Get(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[key]
}

func (s *State) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

func (s *State) Has(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.data[key]
	return ok
}

func (s *State) Snapshot() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]string, len(s.data))
	for k, v := range s.data {
		out[k] = v
	}
	return out
}