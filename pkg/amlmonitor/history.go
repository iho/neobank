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

package amlmonitor

import (
	"sync"
	"time"

	"github.com/shopspring/decimal"

	"github.com/iho/neobank/pkg/money"
)

type historyEntry struct {
	at     time.Time
	userID string
	amount decimal.Decimal
}

// HistoryStore tracks completed transfers for structuring and velocity rules.
type HistoryStore interface {
	RecordAt(userID, amount string, at time.Time) error
	CountInBandLast24h(userID string, minAmt, maxAmt decimal.Decimal, now time.Time) int
	SumLast24h(userID string, now time.Time) decimal.Decimal
}

// MemoryHistoryStore tracks recent transfers in-process (MVP).
type MemoryHistoryStore struct {
	entries []historyEntry
	mu      sync.Mutex
}

func NewMemoryHistoryStore() *MemoryHistoryStore {
	return &MemoryHistoryStore{}
}

func (s *MemoryHistoryStore) RecordAt(userID, amount string, at time.Time) error {
	amt, err := money.Parse(amount)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries = append(s.entries, historyEntry{
		userID: userID,
		amount: amt,
		at:     at.UTC(),
	})
	s.pruneLocked(at.UTC().Add(-24 * time.Hour))

	return nil
}

func (s *MemoryHistoryStore) CountInBandLast24h(userID string, minAmt, maxAmt decimal.Decimal, now time.Time) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := now.Add(-24 * time.Hour)
	count := 0
	for _, e := range s.entries {
		if e.userID != userID || e.at.Before(cutoff) {
			continue
		}
		if !e.amount.LessThan(minAmt) && !e.amount.GreaterThan(maxAmt) {
			count++
		}
	}

	return count
}

func (s *MemoryHistoryStore) SumLast24h(userID string, now time.Time) decimal.Decimal {
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

func (s *MemoryHistoryStore) pruneLocked(cutoff time.Time) {
	kept := s.entries[:0]
	for _, e := range s.entries {
		if !e.at.Before(cutoff) {
			kept = append(kept, e)
		}
	}

	s.entries = kept
}