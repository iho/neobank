package middleware

import (
	"net/http"
	"strings"

	"github.com/iho/neobank/pkg/auth"
	"github.com/iho/neobank/pkg/reqctx"
)

// Actor resolves the authenticated user from the inbound request and attaches
// it to context so audit records and downstream service calls record who
// initiated the mutation.
func Actor(jwtAuth *auth.JWT, allowDevAuth bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			if actor := ResolveUserID(r.Header.Get("Authorization"), r.Header.Get(reqctx.UserIDHeader), jwtAuth, allowDevAuth); actor != "" {
				ctx = reqctx.WithActor(ctx, actor)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ResolveUserID extracts the caller identity from the Authorization bearer token
// (JWT or legacy dev token) or, in development only, the X-User-Id header.
func ResolveUserID(authHeader, xUserID string, jwtAuth *auth.JWT, allowDevAuth bool) string {
	if allowDevAuth && xUserID != "" {
		return xUserID
	}
	raw := strings.TrimSpace(authHeader)
	if !strings.HasPrefix(raw, "Bearer ") {
		return ""
	}
	token := strings.TrimSpace(strings.TrimPrefix(raw, "Bearer "))
	if token == "" {
		return ""
	}

	if allowDevAuth && strings.HasPrefix(token, "access.") {
		parts := strings.Split(token, ".")
		if len(parts) >= 2 {
			return parts[1]
		}
	}

	if jwtAuth != nil {
		userID, err := jwtAuth.ValidateAccessToken(token)
		if err == nil {
			return userID
		}
	}
	return ""
}