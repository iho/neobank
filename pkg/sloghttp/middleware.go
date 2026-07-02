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

package sloghttp

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/iho/neobank/pkg/reqctx"
)

const (
	idempotencyKeyHeader = "Idempotency-Key"
)

// Options configure access-log middleware.
type Options struct {
	SkipPaths map[string]struct{}
	Service   string
}

// Option mutates Options.
type Option func(*Options)

// WithService tags each log line with the service name.
func WithService(name string) Option {
	return func(o *Options) {
		o.Service = name
	}
}

// WithSkipPaths excludes paths from access logs (e.g. health checks).
func WithSkipPaths(paths ...string) Option {
	return func(o *Options) {
		if o.SkipPaths == nil {
			o.SkipPaths = make(map[string]struct{}, len(paths))
		}

		for _, p := range paths {
			o.SkipPaths[p] = struct{}{}
		}
	}
}

// AccessLog emits one structured log line per HTTP request with correlation_id,
// user_id (actor), idempotency_key, method, path, status, and duration.
func AccessLog(logger *slog.Logger, opts ...Option) func(http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	cfg := Options{
		SkipPaths: map[string]struct{}{"/health": {}},
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, skip := cfg.SkipPaths[r.URL.Path]; skip {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)

			attrs := []any{
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"duration_ms", time.Since(start).Milliseconds(),
				"bytes", rec.bytes,
			}
			if cfg.Service != "" {
				attrs = append(attrs, "service", cfg.Service)
			}

			if requestID := middleware.GetReqID(r.Context()); requestID != "" {
				attrs = append(attrs, "request_id", requestID)
			}

			if correlationID := reqctx.CorrelationID(r.Context()); correlationID != "" {
				attrs = append(attrs, "correlation_id", correlationID)
			}

			if actor := reqctx.Actor(r.Context()); actor != "" && actor != "system" {
				attrs = append(attrs, "user_id", actor)
			}

			if idem := r.Header.Get(idempotencyKeyHeader); idem != "" {
				attrs = append(attrs, "idempotency_key", idem)
			}

			if replay := w.Header().Get("X-Idempotent-Replay"); replay == "true" {
				attrs = append(attrs, "idempotent_replay", true)
			}

			level := slog.LevelInfo
			switch {
			case rec.status >= 500:
				level = slog.LevelError
			case rec.status >= 400:
				level = slog.LevelWarn
			}

			logger.Log(r.Context(), level, "http_request", attrs...)
		})
	}
}

type responseRecorder struct {
	http.ResponseWriter

	status int
	bytes  int
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(p []byte) (int, error) {
	n, err := r.ResponseWriter.Write(p)

	r.bytes += n

	return n, err
}
