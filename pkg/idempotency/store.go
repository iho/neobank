package idempotency

import (
	"context"
	"time"
)

// CachedResponse stores a prior successful HTTP response.
type CachedResponse struct {
	Fingerprint string
	StatusCode  int
	Body        []byte
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