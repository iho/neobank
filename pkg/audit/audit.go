// Package audit defines the shared shape of the append-only audit_log table
// that each service (user, payment, card) writes to in the same transaction
// as any status-changing mutation, so lifecycle history survives destructive
// UPDATEs on the domain tables.
package audit

import (
	"context"

	"github.com/iho/neobank/pkg/reqctx"
	"github.com/jackc/pgx/v5"
)

// Entry is one append-only audit record. Actor and CorrelationID are filled
// in from ctx (see pkg/reqctx) by Recorder.Record if left blank, so call
// sites only need to describe the entity/transition/why.
type Entry struct {
	EntityType    string
	EntityID      string
	Action        string
	FromStatus    string
	ToStatus      string
	Actor         string
	CorrelationID string
	Metadata      map[string]any
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
