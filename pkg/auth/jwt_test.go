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