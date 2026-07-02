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
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"time"
)

// Middleware requires Idempotency-Key on mutating requests and replays cached responses.
func Middleware(store Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			key := r.Header.Get("Idempotency-Key")
			if key == "" {
				http.Error(w, `{"error":"idempotency_key_required"}`, http.StatusBadRequest)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, `{"error":"read_body_failed"}`, http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(body))

			fingerprint := hashBody(body)
			if cached, err := store.Get(r.Context(), key); err == nil {
				if cached.Fingerprint != fingerprint {
					http.Error(w, `{"error":"idempotency_key_reused"}`, http.StatusConflict)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Idempotent-Replay", "true")
				w.WriteHeader(cached.StatusCode)

				_, _ = w.Write(cached.Body)

				return
			} else if !errors.Is(err, ErrNotFound) {
				http.Error(w, `{"error":"idempotency_store_error"}`, http.StatusInternalServerError)
				return
			}

			rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)

			if rec.status >= 200 && rec.status < 300 {
				_ = store.Set(r.Context(), key, CachedResponse{
					Fingerprint: fingerprint,
					StatusCode:  rec.status,
					Body:        rec.body,
				}, 24*time.Hour)
			}
		})
	}
}

func hashBody(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

type responseRecorder struct {
	http.ResponseWriter
	body   []byte
	status int
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body = append(r.body, b...)
	return r.ResponseWriter.Write(b)
}
