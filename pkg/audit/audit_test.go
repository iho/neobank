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

package audit

import (
	"testing"

	"github.com/iho/neobank/pkg/reqctx"
)

func TestResolveUsesActorFromContext(t *testing.T) {
	ctx := reqctx.WithActor(t.Context(), "550e8400-e29b-41d4-a716-446655440000")

	ctx = reqctx.WithCorrelationID(ctx, "corr-1")

	entry := Resolve(ctx, Entry{
		EntityType: "transfer",
		EntityID:   "tx-1",
		Action:     "complete",
		ToStatus:   "completed",
	})

	if entry.Actor != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("Actor = %q", entry.Actor)
	}

	if entry.CorrelationID != "corr-1" {
		t.Fatalf("CorrelationID = %q", entry.CorrelationID)
	}
}

func TestResolvePreservesExplicitActor(t *testing.T) {
	ctx := reqctx.WithActor(t.Context(), "from-context")

	entry := Resolve(ctx, Entry{
		Actor:  "explicit-actor",
		Action: "freeze",
	})
	if entry.Actor != "explicit-actor" {
		t.Fatalf("Actor = %q, want explicit-actor", entry.Actor)
	}
}
