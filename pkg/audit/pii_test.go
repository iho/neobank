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

func TestResolvePIIAccessUsesActorFromContext(t *testing.T) {
	ctx := reqctx.WithActor(t.Context(), "550e8400-e29b-41d4-a716-446655440000")
	ctx = reqctx.WithCorrelationID(ctx, "corr-pii")

	entry := ResolvePIIAccess(ctx, PIIAccessEntry{
		SubjectUserID: "subj-1",
		Resource:      PIIResourceProfile,
	})

	if entry.Actor != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("Actor = %q", entry.Actor)
	}
	if entry.CorrelationID != "corr-pii" {
		t.Fatalf("CorrelationID = %q", entry.CorrelationID)
	}
}

func TestResolvePIIAccessDefaultsActorToSystem(t *testing.T) {
	entry := ResolvePIIAccess(t.Context(), PIIAccessEntry{
		SubjectUserID: "subj-1",
		Resource:      PIIResourceKYCStatus,
	})
	if entry.Actor != "system" {
		t.Fatalf("Actor = %q, want system", entry.Actor)
	}
}

func TestResolvePIIAccessPreservesExplicitActor(t *testing.T) {
	ctx := reqctx.WithActor(t.Context(), "from-context")

	entry := ResolvePIIAccess(ctx, PIIAccessEntry{
		Actor:         "explicit-actor",
		SubjectUserID: "subj-1",
		Resource:      PIIResourceInternalWallet,
	})
	if entry.Actor != "explicit-actor" {
		t.Fatalf("Actor = %q, want explicit-actor", entry.Actor)
	}
}