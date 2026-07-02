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

package fraud_test

import (
	"testing"
	"time"

	"github.com/iho/neobank/pkg/fraud"
)

func TestChecker_VelocityHourlyDeny(t *testing.T) {
	store := fraud.NewMemoryVelocityStore()
	checker := fraud.NewCheckerWithVelocity(store)
	now := time.Date(2026, time.July, 2, 12, 0, 0, 0, time.UTC)

	for i := range 10 {
		result, err := checker.Evaluate("user-1", "p2p", "1.00", "USD", fraud.EvaluateOpts{Now: now})
		if err != nil {
			t.Fatal(err)
		}

		if result.Decision != fraud.DecisionAllow {
			t.Fatalf("attempt %d decision = %v", i+1, result.Decision)
		}
	}

	result, err := checker.Evaluate("user-1", "p2p", "1.00", "USD", fraud.EvaluateOpts{Now: now})
	if err != nil {
		t.Fatal(err)
	}

	if result.Decision != fraud.DecisionDeny || result.ReasonCode != "VELOCITY_HOURLY" {
		t.Fatalf("decision = %v code = %s", result.Decision, result.ReasonCode)
	}
}
