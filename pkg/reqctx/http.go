package reqctx

import (
	"net/http"

	"github.com/google/uuid"
)

const (
	CorrelationHeader = "X-Correlation-Id"
	CausationHeader   = "X-Causation-Id"
)

// Middleware extracts X-Correlation-Id / X-Causation-Id from the inbound
// request into context (generating a correlation ID if the caller didn't
// send one), and echoes the correlation ID back on the response so it can be
// surfaced in client-side logs and support tickets.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := r.Header.Get(CorrelationHeader)
		if correlationID == "" {
			correlationID = uuid.NewString()
		}
		ctx := WithCorrelationID(r.Context(), correlationID)
		if causationID := r.Header.Get(CausationHeader); causationID != "" {
			ctx = WithCausationID(ctx, causationID)
		}
		w.Header().Set(CorrelationHeader, correlationID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Transport wraps an http.RoundTripper and forwards the correlation ID (and,
// if set, the causation ID) from the outgoing request's context onto the
// request headers, so downstream services can pick them back up via
// Middleware. Pass nil to wrap http.DefaultTransport.
func Transport(base http.RoundTripper) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		ctx := req.Context()
		if id := CorrelationID(ctx); id != "" {
			req = req.Clone(ctx)
			req.Header.Set(CorrelationHeader, id)
		}
		if id := CausationID(ctx); id != "" {
			if req.Header.Get(CausationHeader) == "" {
				req.Header.Set(CausationHeader, id)
			}
		}
		return base.RoundTrip(req)
	})
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
