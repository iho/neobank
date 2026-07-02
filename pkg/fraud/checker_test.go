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