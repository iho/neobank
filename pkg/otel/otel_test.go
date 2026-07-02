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

package otel_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iho/neobank/pkg/otel"
	"github.com/iho/neobank/pkg/reqctx"
)

func TestEnabled(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	if otel.Enabled() {
		t.Fatal("expected disabled without endpoint")
	}

	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317")
	if !otel.Enabled() {
		t.Fatal("expected enabled with endpoint")
	}
}

func TestHTTPMiddleware_DisabledIsNoop(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")

	called := false
	handler := otel.HTTPMiddleware("test")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected handler to run")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestOutboundTransport_DisabledUsesReqctxOnly(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")

	var gotCorrelation string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCorrelation = r.Header.Get("X-Correlation-Id")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL, http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	req = req.WithContext(reqctx.WithCorrelationID(req.Context(), "corr-123"))

	client := &http.Client{Transport: otel.OutboundTransport(nil)}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if gotCorrelation != "corr-123" {
		t.Fatalf("correlation = %q", gotCorrelation)
	}
}