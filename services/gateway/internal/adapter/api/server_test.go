package api

import (
	"testing"

	"github.com/iho/neobank/pkg/auth"
)

func TestResolveUserIDFromJWT(t *testing.T) {
	jwtAuth := auth.NewJWT("gateway-test-secret")
	access, _, err := jwtAuth.Issue("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatal(err)
	}

	srv := &Server{jwt: jwtAuth}
	authHeader := "Bearer " + access
	userID := srv.resolveUserID(&authHeader, nil)
	if userID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("userID = %q", userID)
	}
}

func TestResolveUserIDLegacyDevToken(t *testing.T) {
	srv := &Server{jwt: auth.NewJWT("secret"), allowDevAuth: true}
	authHeader := "Bearer access.legacy-user-id.devsecret.123"
	userID := srv.resolveUserID(&authHeader, nil)
	if userID != "legacy-user-id" {
		t.Fatalf("userID = %q", userID)
	}
}

func TestResolveUserIDLegacyDevTokenBlockedOutsideDev(t *testing.T) {
	srv := &Server{jwt: auth.NewJWT("secret"), allowDevAuth: false}
	authHeader := "Bearer access.legacy-user-id.devsecret.123"
	userID := srv.resolveUserID(&authHeader, nil)
	if userID != "" {
		t.Fatalf("expected legacy dev token to be rejected outside dev, got userID = %q", userID)
	}
}

func TestResolveUserIDXUserIDBlockedOutsideDev(t *testing.T) {
	srv := &Server{jwt: auth.NewJWT("secret"), allowDevAuth: false}
	xUserID := "some-user-id"
	userID := srv.resolveUserID(nil, &xUserID)
	if userID != "" {
		t.Fatalf("expected X-User-Id bypass to be rejected outside dev, got userID = %q", userID)
	}
}
