package fraud_test

import (
	"testing"
	"time"

	"github.com/iho/neobank/pkg/fraud"
)

func TestChecker_VelocityHourlyDeny(t *testing.T) {
	store := fraud.NewMemoryVelocityStore()
	checker := fraud.NewCheckerWithVelocity(store)
	now := time.Date(2026, 7, 2, 12, 0, 0, 0, time.UTC)

	for i := 0; i < 10; i++ {
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