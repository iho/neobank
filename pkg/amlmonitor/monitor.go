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

package amlmonitor

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"github.com/iho/neobank/pkg/money"
)

type Disposition int

const (
	DispositionClear Disposition = iota + 1
	DispositionReview
	DispositionReport
)

// DefaultRuleSetVersion identifies the active AML rule configuration.
const DefaultRuleSetVersion = "mvp-1.0.0"

var (
	ctrThreshold         = decimal.NewFromInt(10000)
	structuringBandMin   = decimal.NewFromInt(8000)
	structuringBandMax   = decimal.NewFromInt(9999)
	highCumulative24h    = decimal.NewFromInt(15000)
	structuringMinCount  = 2
)

type Result struct {
	Disposition    Disposition
	ReasonCode     string
	RuleSetVersion string
	RiskScore      int
}

type EvaluateOpts struct {
	Now time.Time
}

// Monitor implements MVP transaction-monitoring rules distinct from fraud.
type Monitor struct {
	history HistoryStore
}

func NewMonitor(history HistoryStore) *Monitor {
	store := history
	if store == nil {
		store = NewMemoryHistoryStore()
	}

	return &Monitor{history: store}
}

func (m *Monitor) RuleSetVersion() string {
	return DefaultRuleSetVersion
}

func (m *Monitor) outcome(disposition Disposition, reason string, score int) Result {
	return Result{
		Disposition:    disposition,
		ReasonCode:     reason,
		RiskScore:      score,
		RuleSetVersion: m.RuleSetVersion(),
	}
}

func (m *Monitor) Evaluate(userID, transactionType, amount, _ string, opts EvaluateOpts) (Result, error) {
	now := opts.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}

	amt, err := money.Parse(amount)
	if err != nil {
		return Result{}, fmt.Errorf("parse amount: %w", err)
	}

	if amt.GreaterThanOrEqual(ctrThreshold) {
		_ = m.history.RecordAt(userID, amount, now)
		return m.outcome(DispositionReport, "CTR_THRESHOLD", 95), nil
	}

	inStructuringBand := !amt.LessThan(structuringBandMin) && !amt.GreaterThan(structuringBandMax)
	if inStructuringBand {
		prior := m.history.CountInBandLast24h(userID, structuringBandMin, structuringBandMax, now)
		if prior+1 >= structuringMinCount {
			_ = m.history.RecordAt(userID, amount, now)
			return m.outcome(DispositionReport, "STRUCTURING", 90), nil
		}
	}

	priorSum := m.history.SumLast24h(userID, now)
	if priorSum.Add(amt).GreaterThan(highCumulative24h) {
		_ = m.history.RecordAt(userID, amount, now)
		return m.outcome(DispositionReview, "HIGH_CUMULATIVE_24H", 70), nil
	}

	_ = m.history.RecordAt(userID, amount, now)

	if transactionType == "p2p" && amt.GreaterThan(decimal.NewFromInt(5000)) {
		return m.outcome(DispositionReview, "P2P_HIGH_VALUE", 45), nil
	}

	return m.outcome(DispositionClear, "OK", 5), nil
}