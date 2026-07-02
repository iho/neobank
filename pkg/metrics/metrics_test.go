package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestMountAndMiddleware(t *testing.T) {
	t.Parallel()
	r := chi.NewRouter()
	r.Use(HTTPMiddleware("test"))
	Mount(r)
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("health status=%d", rr.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("metrics status=%d", rr.Code)
	}
	if body := rr.Body.String(); body == "" || body[0] != '#' {
		t.Fatalf("expected prometheus text, got %q", body[:min(40, len(body))])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}