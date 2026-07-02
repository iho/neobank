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

package saga

import (
	"context"
	"fmt"
)

// StepFunc executes or compensates a single saga step.
type StepFunc func(ctx context.Context, state *State) error

// Step defines one unit of work with an optional compensating action.
type Step struct {
	Execute    StepFunc
	Compensate StepFunc
	Name       string
}

// InstanceStore persists saga execution state for retries and idempotency.
type InstanceStore interface {
	GetOrCreate(ctx context.Context, sagaType, idempotencyKey string, initial *State) (*Instance, error)
	Save(ctx context.Context, inst *Instance) error
}

// Instance tracks saga progress.
type Instance struct {
	CompletedSteps map[string]bool
	State          *State
	ID             string
	SagaType       string
	IdempotencyKey string
	Status         string
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
	store    InstanceStore
	sagaType string
	steps    []Step
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
		err := step.Execute(ctx, inst.State)
		if err != nil {
			compErr := o.compensate(ctx, inst, idx)
			if compErr != nil {
				return compErr
			}

			return err
		}

		inst.MarkStepCompleted(step.Name)
		err = o.store.Save(ctx, inst)
		if err != nil {
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
		err := o.steps[i].Compensate(ctx, inst.State)
		if err != nil {
			inst.Status = "failed"
			_ = o.store.Save(ctx, inst)

			return fmt.Errorf("compensate %s: %w", o.steps[i].Name, err)
		}

		inst.MarkStepCompensated(o.steps[i].Name)
	}

	inst.Status = "failed"

	return o.store.Save(ctx, inst)
}
