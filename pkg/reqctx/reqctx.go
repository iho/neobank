// Package reqctx carries request-scoped tracing identifiers (correlation ID,
// causation ID, actor) through context.Context so that every audit record,
// outbox event, and log line written while handling a request can be tied
// back to the request that caused it.
package reqctx

import "context"

type ctxKey int

const (
	correlationIDKey ctxKey = iota
	causationIDKey
	actorKey
)

// WithCorrelationID attaches the correlation ID for the current request chain.
func WithCorrelationID(ctx context.Context, id string) context.Context {
	if id == "" {
		return ctx
	}
	return context.WithValue(ctx, correlationIDKey, id)
}

// CorrelationID returns the correlation ID stored on ctx, or "" if absent.
func CorrelationID(ctx context.Context) string {
	v, _ := ctx.Value(correlationIDKey).(string)
	return v
}

// WithCausationID attaches the ID of the event/request that directly caused
// the current unit of work (distinct from the correlation ID, which spans
// the whole chain).
func WithCausationID(ctx context.Context, id string) context.Context {
	if id == "" {
		return ctx
	}
	return context.WithValue(ctx, causationIDKey, id)
}

// CausationID returns the causation ID stored on ctx, or "" if absent.
func CausationID(ctx context.Context) string {
	v, _ := ctx.Value(causationIDKey).(string)
	return v
}

// WithActor attaches the identity (user ID, or "system"/worker name)
// responsible for the current unit of work, for audit logging.
func WithActor(ctx context.Context, actor string) context.Context {
	if actor == "" {
		return ctx
	}
	return context.WithValue(ctx, actorKey, actor)
}

// Actor returns the actor stored on ctx, defaulting to "system" if absent.
func Actor(ctx context.Context) string {
	v, _ := ctx.Value(actorKey).(string)
	if v == "" {
		return "system"
	}
	return v
}
