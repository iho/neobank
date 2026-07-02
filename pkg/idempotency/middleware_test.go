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

package idempotency_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/iho/neobank/pkg/idempotency"
)

func TestMiddleware_ReplaysSuccessfulResponse(t *testing.T) {
	store := idempotency.NewMemoryStore()
	handler := idempotency.Middleware(store)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)

		_, _ = w.Write([]byte(`{"ok":true}`))
	}))

	body := strings.NewReader(`{"email":"a@b.com"}`)
	req1 := httptest.NewRequest(http.MethodPost, "/v1/auth/register", body)
	req1.Header.Set("Idempotency-Key", "key-1")

	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusCreated {
		t.Fatalf("first status = %d", rec1.Code)
	}

	body2 := strings.NewReader(`{"email":"a@b.com"}`)
	req2 := httptest.NewRequest(http.MethodPost, "/v1/auth/register", body2)
	req2.Header.Set("Idempotency-Key", "key-1")

	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusCreated {
		t.Fatalf("replay status = %d", rec2.Code)
	}

	if rec2.Header().Get("X-Idempotent-Replay") != "true" {
		t.Fatal("expected replay header")
	}
}
