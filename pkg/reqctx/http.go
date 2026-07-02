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

package reqctx

import (
	"net/http"

	"github.com/google/uuid"
)

const (
	CorrelationHeader = "X-Correlation-Id"
	CausationHeader   = "X-Causation-Id"
	ActorHeader       = "X-Actor-Id"
	// UserIDHeader is set by the gateway on internal service calls to identify
	// the subject user. Used as an actor fallback when ActorHeader is absent.
	UserIDHeader = "X-User-Id"
)

// Middleware extracts X-Correlation-Id / X-Causation-Id from the inbound
// request into context (generating a correlation ID if the caller didn't
// send one), and echoes the correlation ID back on the response so it can be
// surfaced in client-side logs and support tickets.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := r.Header.Get(CorrelationHeader)
		if correlationID == "" {
			correlationID = uuid.NewString()
		}

		ctx := WithCorrelationID(r.Context(), correlationID)
		if causationID := r.Header.Get(CausationHeader); causationID != "" {
			ctx = WithCausationID(ctx, causationID)
		}

		if actor := r.Header.Get(ActorHeader); actor != "" {
			ctx = WithActor(ctx, actor)
		} else if userID := r.Header.Get(UserIDHeader); userID != "" {
			// Internal service calls from the gateway always set X-User-Id;
			// treat it as the actor when X-Actor-Id was not forwarded yet.
			ctx = WithActor(ctx, userID)
		}

		w.Header().Set(CorrelationHeader, correlationID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Transport wraps an http.RoundTripper and forwards the correlation ID (and,
// if set, the causation ID) from the outgoing request's context onto the
// request headers, so downstream services can pick them back up via
// Middleware. Pass nil to wrap http.DefaultTransport.
func Transport(base http.RoundTripper) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}

	return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		ctx := req.Context()
		correlationID := CorrelationID(ctx)
		causationID := CausationID(ctx)
		actor := Actor(ctx)

		needsClone := correlationID != "" ||
			(causationID != "" && req.Header.Get(CausationHeader) == "") ||
			(actor != "" && actor != "system")
		if needsClone {
			req = req.Clone(ctx)
		}

		if correlationID != "" {
			req.Header.Set(CorrelationHeader, correlationID)
		}

		if causationID != "" && req.Header.Get(CausationHeader) == "" {
			req.Header.Set(CausationHeader, causationID)
		}

		if actor != "" && actor != "system" {
			req.Header.Set(ActorHeader, actor)
		}

		return base.RoundTrip(req)
	})
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
