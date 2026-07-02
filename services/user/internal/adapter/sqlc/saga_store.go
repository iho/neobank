package sqlcrepo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/saga"
	"github.com/iho/neobank/services/user/internal/gen/sqlc"
	"github.com/jackc/pgx/v5"
)

type SagaStore struct {
	q sqlc.Querier
}

func NewSagaStore(q sqlc.Querier) *SagaStore {
	return &SagaStore{q: q}
}

func (s *SagaStore) GetOrCreate(ctx context.Context, sagaType, idempotencyKey string, initial *saga.State) (*saga.Instance, error) {
	row, err := s.q.GetSagaByIdempotencyKey(ctx, idempotencyKey)
	if err == nil {
		return mapSagaRow(row.ID, row.SagaType, row.IdempotencyKey, row.Status, row.CompletedSteps, row.Context)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	inst := saga.Instance{
		ID:             uuid.NewString(),
		SagaType:       sagaType,
		IdempotencyKey: idempotencyKey,
		Status:         "running",
		CompletedSteps: map[string]bool{},
		State:          initial,
	}
	contextBytes, _ := json.Marshal(initial.Snapshot())
	id, err := uuid.Parse(inst.ID)
	if err != nil {
		return nil, err
	}
	if err := s.q.CreateSagaInstance(ctx, sqlc.CreateSagaInstanceParams{
		ID:             id,
		SagaType:       inst.SagaType,
		IdempotencyKey: inst.IdempotencyKey,
		Status:         inst.Status,
		Context:        contextBytes,
	}); err != nil {
		return nil, fmt.Errorf("insert saga instance: %w", err)
	}
	return &inst, nil
}

func (s *SagaStore) Save(ctx context.Context, inst *saga.Instance) error {
	completedJSON, _ := json.Marshal(inst.CompletedSteps)
	contextJSON, _ := json.Marshal(inst.State.Snapshot())
	id, err := uuid.Parse(inst.ID)
	if err != nil {
		return err
	}
	return s.q.UpdateSagaInstance(ctx, sqlc.UpdateSagaInstanceParams{
		ID:             id,
		Status:         inst.Status,
		CompletedSteps: completedJSON,
		Context:        contextJSON,
	})
}

func mapSagaRow(id uuid.UUID, sagaType, idempotencyKey, status string, completedSteps, contextJSON []byte) (*saga.Instance, error) {
	inst := saga.Instance{
		ID:             id.String(),
		SagaType:       sagaType,
		IdempotencyKey: idempotencyKey,
		Status:         status,
		CompletedSteps: map[string]bool{},
	}
	_ = json.Unmarshal(completedSteps, &inst.CompletedSteps)
	stateData := map[string]string{}
	_ = json.Unmarshal(contextJSON, &stateData)
	inst.State = saga.NewState(stateData)
	return &inst, nil
}