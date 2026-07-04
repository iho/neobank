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
	"errors"
	"strconv"
	"testing"
	"time"
)

func TestSignAndVerifySignatureRoundTrip(t *testing.T) {
	secret := []byte("s3cr3t")
	body := []byte(`{"hello":"world"}`)
	ts := time.Now().Unix()
	sig := Sign(secret, ts, body)

	if err := VerifySignature(secret, strconv.FormatInt(ts, 10), sig, body, time.Minute); err != nil {
		t.Fatalf("expected valid signature, got %v", err)
	}
}

func TestVerifySignatureRejectsTamperedBody(t *testing.T) {
	secret := []byte("s3cr3t")
	ts := time.Now().Unix()
	sig := Sign(secret, ts, []byte(`{"amount":"10.00"}`))

	err := VerifySignature(secret, strconv.FormatInt(ts, 10), sig, []byte(`{"amount":"99999.00"}`), time.Minute)
	if !errors.Is(err, ErrSignatureMismatch) {
		t.Fatalf("expected ErrSignatureMismatch, got %v", err)
	}
}

func TestVerifySignatureRejectsWrongSecret(t *testing.T) {
	ts := time.Now().Unix()
	body := []byte(`{}`)
	sig := Sign([]byte("secret-a"), ts, body)

	err := VerifySignature([]byte("secret-b"), strconv.FormatInt(ts, 10), sig, body, time.Minute)
	if !errors.Is(err, ErrSignatureMismatch) {
		t.Fatalf("expected ErrSignatureMismatch, got %v", err)
	}
}

func TestVerifySignatureRejectsExpiredTimestamp(t *testing.T) {
	secret := []byte("s3cr3t")
	ts := time.Now().Add(-time.Hour).Unix()
	body := []byte(`{}`)
	sig := Sign(secret, ts, body)

	err := VerifySignature(secret, strconv.FormatInt(ts, 10), sig, body, time.Minute)
	if !errors.Is(err, ErrSignatureExpired) {
		t.Fatalf("expected ErrSignatureExpired, got %v", err)
	}
}

func TestVerifySignatureRejectsMissingHeaders(t *testing.T) {
	err := VerifySignature([]byte("s"), "", "", []byte("{}"), time.Minute)
	if !errors.Is(err, ErrMissingSignature) {
		t.Fatalf("expected ErrMissingSignature, got %v", err)
	}
}

func TestVerifySignatureRejectsBadTimestamp(t *testing.T) {
	err := VerifySignature([]byte("s"), "not-a-number", "abc", []byte("{}"), time.Minute)
	if !errors.Is(err, ErrBadTimestamp) {
		t.Fatalf("expected ErrBadTimestamp, got %v", err)
	}
}
