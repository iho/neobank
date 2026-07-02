package saga

import (
	"context"
	"fmt"
)

// StepFunc executes or compensates a single saga step.
type StepFunc func(ctx context.Context, state *State) error

// Step defines one unit of work with an optional compensating action.
type Step struct {
	Name       string
	Execute    StepFunc
	Compensate StepFunc
}

// InstanceStore persists saga execution state for retries and idempotency.
type InstanceStore interface {
	GetOrCreate(ctx context.Context, sagaType, idempotencyKey string, initial *State) (*Instance, error)
	Save(ctx context.Context, inst *Instance) error
}

// Instance tracks saga progress.
type Instance struct {
	ID             string
	SagaType       string
	IdempotencyKey string
	Status         string
	CompletedSteps map[string]bool
	State          *State
}

func (i *Instance) IsStepCompleted(name string) bool {
	return i.CompletedSteps[name]
}

func (i *Instance) MarkStepCompleted(name string) {
	i.CompletedSteps[name] = true
}

func (i *Instance) MarkStepCompensated(name string) {
	delete(i.CompletedSteps, name)
}

// Orchestrator runs steps in order and compensates on failure.
type Orchestrator struct {
	sagaType string
	steps    []Step
	store    InstanceStore
}

func New(sagaType string, steps []Step, store InstanceStore) *Orchestrator {
	return &Orchestrator{sagaType: sagaType, steps: steps, store: store}
}

func (o *Orchestrator) Run(ctx context.Context, idempotencyKey string, initial *State) error {
	inst, err := o.store.GetOrCreate(ctx, o.sagaType, idempotencyKey, initial)
	if err != nil {
		return err
	}
	if inst.Status == "completed" {
		return nil
	}

	for idx, step := range o.steps {
		if inst.IsStepCompleted(step.Name) {
			continue
		}
		if err := step.Execute(ctx, inst.State); err != nil {
			if compErr := o.compensate(ctx, inst, idx); compErr != nil {
				return compErr
			}
			return err
		}
		inst.MarkStepCompleted(step.Name)
		if err := o.store.Save(ctx, inst); err != nil {
			return err
		}
	}

	inst.Status = "completed"
	return o.store.Save(ctx, inst)
}

func (o *Orchestrator) compensate(ctx context.Context, inst *Instance, failedIdx int) error {
	inst.Status = "compensating"
	for i := failedIdx; i >= 0; i-- {
		if o.steps[i].Compensate == nil || !inst.IsStepCompleted(o.steps[i].Name) {
			continue
		}
		if err := o.steps[i].Compensate(ctx, inst.State); err != nil {
			inst.Status = "failed"
			_ = o.store.Save(ctx, inst)
			return fmt.Errorf("compensate %s: %w", o.steps[i].Name, err)
		}
		inst.MarkStepCompensated(o.steps[i].Name)
	}
	inst.Status = "failed"
	return o.store.Save(ctx, inst)
}