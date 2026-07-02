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

package fraud

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"github.com/iho/neobank/pkg/money"
)

type Decision int

const (
	DecisionAllow Decision = iota + 1
	DecisionReview
	DecisionDeny
)

// DefaultRuleSetVersion identifies the active rule configuration. Bump this
// when thresholds or rule logic change so past fraud_decisions rows remain
// explainable.
const DefaultRuleSetVersion = "mvp-1.0.0"

type Result struct {
	ReasonCode     string
	RuleSetVersion string
	Decision       Decision
	RiskScore      int
}

type VelocityStore interface {
	RecordAt(userID, amount string, at time.Time) error
	CountLastHour(userID string, now time.Time) int
	SumLast24h(userID string, now time.Time) decimal.Decimal
}

type EvaluateOpts struct {
	AccountCreatedAt *time.Time
	Now              time.Time
}

// Checker implements MVP rules-based fraud screening.
type Checker struct {
	blockedUsers map[string]struct{}
	velocity     VelocityStore
}

func NewChecker(blockedUsers ...string) *Checker {
	blocked := make(map[string]struct{}, len(blockedUsers))
	for _, u := range blockedUsers {
		blocked[u] = struct{}{}
	}

	return &Checker{
		blockedUsers: blocked,
		velocity:     NewMemoryVelocityStore(),
	}
}

func NewCheckerWithVelocity(store VelocityStore, blockedUsers ...string) *Checker {
	c := NewChecker(blockedUsers...)

	c.velocity = store

	return c
}

// RuleSetVersion returns the identifier of the rule configuration used by this
// checker. Persist it alongside every fraud evaluation.
func (c *Checker) RuleSetVersion() string {
	return DefaultRuleSetVersion
}

func (c *Checker) outcome(decision Decision, reason string, score int) Result {
	return Result{
		Decision:       decision,
		ReasonCode:     reason,
		RiskScore:      score,
		RuleSetVersion: c.RuleSetVersion(),
	}
}

func (c *Checker) Evaluate(userID, transactionType, amount, _ string, opts EvaluateOpts) (Result, error) {
	now := opts.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}

	if _, blocked := c.blockedUsers[userID]; blocked {
		return c.outcome(DecisionDeny, "USER_BLOCKED", 100), nil
	}

	amt, err := money.Parse(amount)
	if err != nil {
		return Result{}, fmt.Errorf("parse amount: %w", err)
	}

	if c.velocity != nil {
		_ = c.velocity.RecordAt(userID, amount, now)

		if c.velocity.CountLastHour(userID, now) > 10 {
			return c.outcome(DecisionDeny, "VELOCITY_HOURLY", 90), nil
		}

		if c.velocity.SumLast24h(userID, now).GreaterThan(decimal.NewFromInt(10000)) {
			return c.outcome(DecisionReview, "VELOCITY_DAILY", 75), nil
		}
	}

	if amt.GreaterThan(decimal.NewFromInt(5000)) {
		return c.outcome(DecisionReview, "HIGH_AMOUNT", 70), nil
	}

	if opts.AccountCreatedAt != nil && now.Sub(*opts.AccountCreatedAt) < 24*time.Hour {
		if amt.GreaterThan(decimal.NewFromInt(500)) {
			return c.outcome(DecisionReview, "NEW_ACCOUNT", 55), nil
		}
	}

	if transactionType == "p2p" && amt.GreaterThan(decimal.NewFromInt(500)) {
		return c.outcome(DecisionReview, "P2P_THRESHOLD", 40), nil
	}

	return c.outcome(DecisionAllow, "OK", 10), nil
}
