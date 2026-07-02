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

	"github.com/iho/neobank/pkg/fraud"
)

func TestChecker_BlockedUser(t *testing.T) {
	checker := fraud.NewChecker("user-1")

	result, err := checker.Evaluate("user-1", "p2p", "10.00", "USD", fraud.EvaluateOpts{})
	if err != nil {
		t.Fatal(err)
	}

	if result.Decision != fraud.DecisionDeny {
		t.Fatalf("decision = %v", result.Decision)
	}
}

func TestChecker_AllowSmallTransfer(t *testing.T) {
	checker := fraud.NewChecker()

	result, err := checker.Evaluate("user-2", "p2p", "25.00", "USD", fraud.EvaluateOpts{})
	if err != nil {
		t.Fatal(err)
	}

	if result.Decision != fraud.DecisionAllow {
		t.Fatalf("decision = %v", result.Decision)
	}
}

func TestChecker_ResultIncludesRuleSetVersion(t *testing.T) {
	checker := fraud.NewChecker()
	if checker.RuleSetVersion() != fraud.DefaultRuleSetVersion {
		t.Fatalf("RuleSetVersion() = %q", checker.RuleSetVersion())
	}

	result, err := checker.Evaluate("user-3", "p2p", "25.00", "USD", fraud.EvaluateOpts{})
	if err != nil {
		t.Fatal(err)
	}

	if result.RuleSetVersion != fraud.DefaultRuleSetVersion {
		t.Fatalf("result.RuleSetVersion = %q", result.RuleSetVersion)
	}
}
