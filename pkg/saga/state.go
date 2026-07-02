//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package saga

import "sync"

// State is a thread-safe key-value bag passed between saga steps.
type State struct {
	data map[string]string
	mu   sync.RWMutex
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
