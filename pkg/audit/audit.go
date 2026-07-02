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
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/iho/neobank/pkg/reqctx"
)

// Entry is one append-only audit record. Actor and CorrelationID are filled
// in from ctx (see pkg/reqctx) by Recorder.Record if left blank, so call
// sites only need to describe the entity/transition/why.
type Entry struct {
	Metadata      map[string]any
	EntityType    string
	EntityID      string
	Action        string
	FromStatus    string
	ToStatus      string
	Actor         string
	CorrelationID string
}

// Recorder persists audit entries. Implementations must support being
// wrapped in an open transaction (see outbox.TxPublisher for the pattern)
// so the audit row commits atomically with the domain mutation it describes.
type Recorder interface {
	Record(ctx context.Context, e Entry) error
	WithTx(tx pgx.Tx) Recorder
}

// Resolve fills Actor/CorrelationID from ctx when the caller didn't set them explicitly.
func Resolve(ctx context.Context, e Entry) Entry {
	if e.Actor == "" {
		e.Actor = reqctx.Actor(ctx)
	}

	if e.CorrelationID == "" {
		e.CorrelationID = reqctx.CorrelationID(ctx)
	}

	if e.Metadata == nil {
		e.Metadata = map[string]any{}
	}

	return e
}
