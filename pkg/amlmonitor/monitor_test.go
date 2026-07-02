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

package amlmonitor_test

import (
	"testing"
	"time"

	"github.com/iho/neobank/pkg/amlmonitor"
)

func TestMonitor_CTRThreshold(t *testing.T) {
	monitor := amlmonitor.NewMonitor(nil)

	result, err := monitor.Evaluate("user-1", "p2p", "10000.00", "USD", amlmonitor.EvaluateOpts{})
	if err != nil {
		t.Fatal(err)
	}

	if result.Disposition != amlmonitor.DispositionReport {
		t.Fatalf("disposition = %v", result.Disposition)
	}
	if result.ReasonCode != "CTR_THRESHOLD" {
		t.Fatalf("reason = %q", result.ReasonCode)
	}
}

func TestMonitor_Structuring(t *testing.T) {
	store := amlmonitor.NewMemoryHistoryStore()
	monitor := amlmonitor.NewMonitor(store)
	now := time.Date(2026, 7, 2, 12, 0, 0, 0, time.UTC)
	opts := amlmonitor.EvaluateOpts{Now: now}

	if _, err := monitor.Evaluate("user-2", "p2p", "9000.00", "USD", opts); err != nil {
		t.Fatal(err)
	}

	result, err := monitor.Evaluate("user-2", "p2p", "9000.00", "USD", opts)
	if err != nil {
		t.Fatal(err)
	}

	if result.Disposition != amlmonitor.DispositionReport {
		t.Fatalf("disposition = %v", result.Disposition)
	}
	if result.ReasonCode != "STRUCTURING" {
		t.Fatalf("reason = %q", result.ReasonCode)
	}
}

func TestMonitor_ClearSmallTransfer(t *testing.T) {
	monitor := amlmonitor.NewMonitor(nil)

	result, err := monitor.Evaluate("user-3", "p2p", "50.00", "USD", amlmonitor.EvaluateOpts{})
	if err != nil {
		t.Fatal(err)
	}

	if result.Disposition != amlmonitor.DispositionClear {
		t.Fatalf("disposition = %v", result.Disposition)
	}
}

func TestMonitor_ResultIncludesRuleSetVersion(t *testing.T) {
	monitor := amlmonitor.NewMonitor(nil)
	if monitor.RuleSetVersion() != amlmonitor.DefaultRuleSetVersion {
		t.Fatalf("RuleSetVersion() = %q", monitor.RuleSetVersion())
	}

	result, err := monitor.Evaluate("user-4", "p2p", "50.00", "USD", amlmonitor.EvaluateOpts{})
	if err != nil {
		t.Fatal(err)
	}

	if result.RuleSetVersion != amlmonitor.DefaultRuleSetVersion {
		t.Fatalf("result.RuleSetVersion = %q", result.RuleSetVersion)
	}
}