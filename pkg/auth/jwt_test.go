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

package auth_test

import (
	"testing"

	"github.com/iho/neobank/pkg/auth"
)

func TestJWT_IssueAndValidateAccess(t *testing.T) {
	jwtAuth := auth.NewJWT("test-secret-key")

	access, refresh, err := jwtAuth.Issue("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatal(err)
	}

	if access == "" || refresh == "" {
		t.Fatal("expected non-empty tokens")
	}

	userID, err := jwtAuth.ValidateAccessToken(access)
	if err != nil {
		t.Fatal(err)
	}

	if userID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("userID = %q", userID)
	}

	if _, err := jwtAuth.ValidateAccessToken(refresh); err == nil {
		t.Fatal("expected refresh token to fail access validation")
	}

	refreshUserID, err := jwtAuth.ValidateRefreshToken(refresh)
	if err != nil {
		t.Fatal(err)
	}

	if refreshUserID != userID {
		t.Fatalf("refresh userID = %q", refreshUserID)
	}
}

func TestJWT_Refresh(t *testing.T) {
	jwtAuth := auth.NewJWT("test-secret-key")

	_, refresh, err := jwtAuth.Issue("user-refresh-test")
	if err != nil {
		t.Fatal(err)
	}

	access, newRefresh, userID, err := jwtAuth.Refresh(refresh)
	if err != nil {
		t.Fatal(err)
	}

	if userID != "user-refresh-test" || access == "" || newRefresh == "" {
		t.Fatalf("unexpected refresh output")
	}

	if _, err := jwtAuth.ValidateAccessToken(access); err != nil {
		t.Fatal(err)
	}
}
