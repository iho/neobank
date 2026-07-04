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
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
)

func newSignedRequest(t *testing.T, secret []byte, deliveryID string, body []byte) *http.Request {
	t.Helper()

	ts := time.Now().Unix()
	sig := Sign(secret, ts, body)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/test", bytes.NewReader(body))
	req.Header.Set(HeaderTimestamp, strconv.FormatInt(ts, 10))
	req.Header.Set(HeaderSignature, sig)
	req.Header.Set(HeaderDeliveryID, deliveryID)

	return req
}

func TestVerifyWebhookAcceptsValidRequest(t *testing.T) {
	secret := []byte("test-secret")

	var calls int32

	handler := VerifyWebhook(VerifyWebhookConfig{Secret: secret})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusOK)
	}))

	body := []byte(`{"amount":"10.00"}`)
	req := newSignedRequest(t, secret, uuid.NewString(), body)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected inner handler called once, got %d", calls)
	}
}

func TestVerifyWebhookRejectsInvalidSignature(t *testing.T) {
	handler := VerifyWebhook(VerifyWebhookConfig{Secret: []byte("real-secret")})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	body := []byte(`{"amount":"10.00"}`)
	req := newSignedRequest(t, []byte("wrong-secret"), uuid.NewString(), body)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestVerifyWebhookShortCircuitsDuplicateDelivery(t *testing.T) {
	secret := []byte("test-secret")

	var calls int32

	handler := VerifyWebhook(VerifyWebhookConfig{
		Secret: secret,
		Nonces: NewMemoryNonceStore(),
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusOK)
	}))

	body := []byte(`{"amount":"10.00"}`)
	deliveryID := uuid.NewString()

	req1 := newSignedRequest(t, secret, deliveryID, body)
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)

	req2 := newSignedRequest(t, secret, deliveryID, body)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected inner handler called once despite duplicate delivery, got %d", calls)
	}

	if rr2.Code != http.StatusOK {
		t.Fatalf("expected duplicate delivery to still get 200, got %d", rr2.Code)
	}

	if rr2.Header().Get("X-Vendorsim-Duplicate") != "true" {
		t.Fatal("expected X-Vendorsim-Duplicate header on replay")
	}
}

func TestVerifyWebhookRejectsMissingDeliveryID(t *testing.T) {
	secret := []byte("test-secret")

	handler := VerifyWebhook(VerifyWebhookConfig{Secret: secret})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	body := []byte(`{}`)
	req := newSignedRequest(t, secret, "", body)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
