package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iho/neobank/pkg/auth"
	"github.com/iho/neobank/pkg/reqctx"
)

func TestActorMiddlewareSetsActorFromJWT(t *testing.T) {
	jwtAuth := auth.NewJWT("gateway-test-secret")
	access, _, err := jwtAuth.Issue("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatal(err)
	}

	var gotActor string
	handler := Actor(jwtAuth, false)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotActor = reqctx.Actor(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+access)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if gotActor != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("Actor = %q", gotActor)
	}
}

func TestActorMiddlewareSetsActorFromXUserIDInDev(t *testing.T) {
	var gotActor string
	handler := Actor(auth.NewJWT("secret"), true)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotActor = reqctx.Actor(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(reqctx.UserIDHeader, "dev-user")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if gotActor != "dev-user" {
		t.Fatalf("Actor = %q, want dev-user", gotActor)
	}
}

func TestActorPropagatesToDownstreamViaTransport(t *testing.T) {
	jwtAuth := auth.NewJWT("gateway-test-secret")
	access, _, err := jwtAuth.Issue("user-downstream")
	if err != nil {
		t.Fatal(err)
	}

	var downstreamActor string
	downstream := reqctx.Middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		downstreamActor = reqctx.Actor(r.Context())
	}))

	gateway := Actor(jwtAuth, false)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, "http://payment.local/transfer", nil)
		if err != nil {
			t.Fatal(err)
		}
		client := &http.Client{Transport: reqctx.Transport(handlerRoundTripper{downstream})}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
	}))

	inbound := httptest.NewRequest(http.MethodPost, "/v1/transfers", nil)
	inbound.Header.Set("Authorization", "Bearer "+access)
	rr := httptest.NewRecorder()
	gateway.ServeHTTP(rr, inbound)

	if downstreamActor != "user-downstream" {
		t.Fatalf("downstream Actor = %q, want user-downstream", downstreamActor)
	}
}

type handlerRoundTripper struct {
	h http.Handler
}

func (t handlerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rr := httptest.NewRecorder()
	t.h.ServeHTTP(rr, req)
	return rr.Result(), nil
}

func TestActorMiddlewareLeavesSystemForUnauthenticated(t *testing.T) {
	var gotActor string
	handler := Actor(auth.NewJWT("secret"), false)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotActor = reqctx.Actor(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if gotActor != "system" {
		t.Fatalf("Actor = %q, want system", gotActor)
	}
}