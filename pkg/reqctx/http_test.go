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
	"net/http/httptest"
	"testing"
)

func TestMiddlewareReadsActorHeader(t *testing.T) {
	var gotActor string

	handler := Middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotActor = Actor(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(ActorHeader, "user-123")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if gotActor != "user-123" {
		t.Fatalf("Actor = %q, want user-123", gotActor)
	}
}

func TestMiddlewareFallsBackToUserIDHeader(t *testing.T) {
	var gotActor string

	handler := Middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotActor = Actor(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(UserIDHeader, "user-456")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if gotActor != "user-456" {
		t.Fatalf("Actor = %q, want user-456", gotActor)
	}
}

func TestMiddlewarePrefersActorOverUserID(t *testing.T) {
	var gotActor string

	handler := Middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotActor = Actor(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(ActorHeader, "actor-1")
	req.Header.Set(UserIDHeader, "user-2")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if gotActor != "actor-1" {
		t.Fatalf("Actor = %q, want actor-1", gotActor)
	}
}

func TestTransportForwardsActor(t *testing.T) {
	var gotHeader string

	inner := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		gotHeader = req.Header.Get(ActorHeader)
		return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
	})

	ctx := WithActor(t.Context(), "user-789")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Transport(inner).RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}

	if gotHeader != "user-789" {
		t.Fatalf("X-Actor-Id header = %q, want user-789", gotHeader)
	}
}

func TestTransportSkipsDefaultSystemActor(t *testing.T) {
	var gotHeader string

	inner := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		gotHeader = req.Header.Get(ActorHeader)
		return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
	})

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Transport(inner).RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}

	if gotHeader != "" {
		t.Fatalf("X-Actor-Id header = %q, want empty for default system actor", gotHeader)
	}
}
