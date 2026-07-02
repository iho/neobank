//
// Copyright (c) 2026 Sumicare
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

package idempotency

import (
	"context"
	"time"
)

// CachedResponse stores a prior successful HTTP response.
type CachedResponse struct {
	Fingerprint string
	Body        []byte
	StatusCode  int
}

// Store persists idempotency records.
type Store interface {
	Get(ctx context.Context, key string) (CachedResponse, error)
	Set(ctx context.Context, key string, resp CachedResponse, ttl time.Duration) error
}

// MemoryStore is an in-process store for local development.
type MemoryStore struct {
	data map[string]CachedResponse
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{data: make(map[string]CachedResponse)}
}

func (s *MemoryStore) Get(_ context.Context, key string) (CachedResponse, error) {
	resp, ok := s.data[key]
	if !ok {
		return CachedResponse{}, ErrNotFound
	}

	return resp, nil
}

func (s *MemoryStore) Set(_ context.Context, key string, resp CachedResponse, _ time.Duration) error {
	s.data[key] = resp
	return nil
}
