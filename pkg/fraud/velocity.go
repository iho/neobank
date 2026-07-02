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

package fraud

import (
	"sync"
	"time"

	"github.com/shopspring/decimal"

	"github.com/iho/neobank/pkg/money"
)

type velocityEntry struct {
	at     time.Time
	userID string
	amount decimal.Decimal
}

// MemoryVelocityStore tracks recent transaction attempts in-process (MVP).
type MemoryVelocityStore struct {
	entries []velocityEntry
	mu      sync.Mutex
}

func NewMemoryVelocityStore() *MemoryVelocityStore {
	return &MemoryVelocityStore{}
}

func (s *MemoryVelocityStore) Record(userID, amount string) error {
	return s.RecordAt(userID, amount, time.Now().UTC())
}

func (s *MemoryVelocityStore) RecordAt(userID, amount string, at time.Time) error {
	amt, err := money.Parse(amount)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries = append(s.entries, velocityEntry{
		userID: userID,
		amount: amt,
		at:     at.UTC(),
	})
	s.pruneLocked(at.UTC().Add(-24 * time.Hour))

	return nil
}

func (s *MemoryVelocityStore) CountLastHour(userID string, now time.Time) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := now.Add(-time.Hour)

	count := 0
	for _, e := range s.entries {
		if e.userID == userID && !e.at.Before(cutoff) {
			count++
		}
	}

	return count
}

func (s *MemoryVelocityStore) SumLast24h(userID string, now time.Time) decimal.Decimal {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := now.Add(-24 * time.Hour)

	sum := decimal.Zero
	for _, e := range s.entries {
		if e.userID == userID && !e.at.Before(cutoff) {
			sum = sum.Add(e.amount)
		}
	}

	return sum
}

func (s *MemoryVelocityStore) pruneLocked(cutoff time.Time) {
	kept := s.entries[:0]
	for _, e := range s.entries {
		if !e.at.Before(cutoff) {
			kept = append(kept, e)
		}
	}

	s.entries = kept
}
