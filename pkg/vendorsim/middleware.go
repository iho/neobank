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

package vendorsim

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// NonceStore tracks recently seen delivery IDs so consumers can reject or
// no-op replayed webhooks (a delivery may arrive more than once under
// at-least-once delivery or simulated chaos duplication).
type NonceStore interface {
	// SeenBefore reports whether deliveryID was already recorded within ttl,
	// and records it if not.
	SeenBefore(ctx context.Context, deliveryID string, ttl time.Duration) (bool, error)
}

// MemoryNonceStore is an in-process NonceStore for local development and
// single-instance consumers.
type MemoryNonceStore struct {
	seen map[string]time.Time
	mu   sync.Mutex
}

func NewMemoryNonceStore() *MemoryNonceStore {
	return &MemoryNonceStore{seen: make(map[string]time.Time)}
}

func (s *MemoryNonceStore) SeenBefore(_ context.Context, deliveryID string, ttl time.Duration) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	for id, at := range s.seen {
		if now.Sub(at) > ttl {
			delete(s.seen, id)
		}
	}

	if _, ok := s.seen[deliveryID]; ok {
		return true, nil
	}

	s.seen[deliveryID] = now

	return false, nil
}

// VerifyWebhookConfig configures the consumer-side verification middleware.
type VerifyWebhookConfig struct {
	Nonces   NonceStore
	Secret   []byte
	MaxSkew  time.Duration
	NonceTTL time.Duration
}

// VerifyWebhook checks the signature/timestamp headers a vendorsim Dispatcher
// attaches and short-circuits already-seen deliveries (idempotent consumption
// under duplicate/at-least-once delivery). Handlers behind it can assume the
// payload is authentic and not-yet-processed.
func VerifyWebhook(cfg VerifyWebhookConfig) func(http.Handler) http.Handler {
	maxSkew := cfg.MaxSkew
	if maxSkew <= 0 {
		maxSkew = 5 * time.Minute
	}

	nonceTTL := cfg.NonceTTL
	if nonceTTL <= 0 {
		nonceTTL = 24 * time.Hour
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, `{"error":"read_body_failed"}`, http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(body))

			err = VerifySignature(cfg.Secret, r.Header.Get(HeaderTimestamp), r.Header.Get(HeaderSignature), body, maxSkew)
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error":"invalid_signature","detail":%q}`, err.Error()), http.StatusUnauthorized)
				return
			}

			deliveryID := r.Header.Get(HeaderDeliveryID)
			if deliveryID == "" {
				http.Error(w, `{"error":"missing_delivery_id"}`, http.StatusBadRequest)
				return
			}

			if cfg.Nonces != nil {
				dup, err := cfg.Nonces.SeenBefore(r.Context(), deliveryID, nonceTTL)
				if err != nil {
					http.Error(w, `{"error":"nonce_store_error"}`, http.StatusInternalServerError)
					return
				}

				if dup {
					w.Header().Set("X-Vendorsim-Duplicate", "true")
					w.WriteHeader(http.StatusOK)

					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
