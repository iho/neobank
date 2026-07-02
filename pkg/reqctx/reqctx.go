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
