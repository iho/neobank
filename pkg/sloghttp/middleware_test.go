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

package sloghttp

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/iho/neobank/pkg/reqctx"
)

func TestAccessLogIncludesTraceFields(t *testing.T) {
	var buf bytes.Buffer

	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	handler := middleware.RequestID(
		reqctx.Middleware(
			AccessLog(logger, WithService("test"))(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusCreated)
				}),
			),
		),
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/transfers", http.NoBody)
	req.Header.Set(reqctx.CorrelationHeader, "corr-abc")
	req.Header.Set(reqctx.ActorHeader, "user-123")
	req.Header.Set("Idempotency-Key", "idem-xyz")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var entry map[string]any
	err := json.Unmarshal(buf.Bytes(), &entry)
	if err != nil {
		t.Fatalf("unmarshal log: %v body=%s", err, buf.String())
	}

	if entry["msg"] != "http_request" {
		t.Fatalf("msg = %v", entry["msg"])
	}

	if entry["correlation_id"] != "corr-abc" {
		t.Fatalf("correlation_id = %v", entry["correlation_id"])
	}

	if entry["user_id"] != "user-123" {
		t.Fatalf("user_id = %v", entry["user_id"])
	}

	if entry["idempotency_key"] != "idem-xyz" {
		t.Fatalf("idempotency_key = %v", entry["idempotency_key"])
	}

	if entry["service"] != "test" {
		t.Fatalf("service = %v", entry["service"])
	}
}

func TestAccessLogSkipsHealth(t *testing.T) {
	var buf bytes.Buffer

	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	handler := AccessLog(logger)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if buf.Len() != 0 {
		t.Fatalf("expected no log for /health, got %s", buf.String())
	}
}

func TestLoggerFromContext(t *testing.T) {
	ctx := reqctx.WithCorrelationID(t.Context(), "corr-1")

	ctx = reqctx.WithActor(ctx, "actor-1")

	base := slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil))

	enriched := Logger(ctx, base)
	if enriched == base {
		t.Fatal("expected enriched logger")
	}
}
