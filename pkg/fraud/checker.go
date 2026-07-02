package fraud

import (
	"fmt"
	"time"

	"github.com/iho/neobank/pkg/money"
	"github.com/shopspring/decimal"
)

type Decision int

const (
	DecisionAllow Decision = iota + 1
	DecisionReview
	DecisionDeny
)

type Result struct {
	Decision   Decision
	ReasonCode string
	RiskScore  int
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

func (c *Checker) Evaluate(userID, transactionType, amount, _ string, opts EvaluateOpts) (Result, error) {
	now := opts.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}

	if _, blocked := c.blockedUsers[userID]; blocked {
		return Result{Decision: DecisionDeny, ReasonCode: "USER_BLOCKED", RiskScore: 100}, nil
	}

	amt, err := money.Parse(amount)
	if err != nil {
		return Result{}, fmt.Errorf("parse amount: %w", err)
	}

	if c.velocity != nil {
		_ = c.velocity.RecordAt(userID, amount, now)

		if c.velocity.CountLastHour(userID, now) > 10 {
			return Result{Decision: DecisionDeny, ReasonCode: "VELOCITY_HOURLY", RiskScore: 90}, nil
		}

		if c.velocity.SumLast24h(userID, now).GreaterThan(decimal.NewFromInt(10000)) {
			return Result{Decision: DecisionReview, ReasonCode: "VELOCITY_DAILY", RiskScore: 75}, nil
		}
	}

	if amt.GreaterThan(decimal.NewFromInt(5000)) {
		return Result{Decision: DecisionReview, ReasonCode: "HIGH_AMOUNT", RiskScore: 70}, nil
	}

	if opts.AccountCreatedAt != nil && now.Sub(*opts.AccountCreatedAt) < 24*time.Hour {
		if amt.GreaterThan(decimal.NewFromInt(500)) {
			return Result{Decision: DecisionReview, ReasonCode: "NEW_ACCOUNT", RiskScore: 55}, nil
		}
	}

	if transactionType == "p2p" && amt.GreaterThan(decimal.NewFromInt(500)) {
		return Result{Decision: DecisionReview, ReasonCode: "P2P_THRESHOLD", RiskScore: 40}, nil
	}

	return Result{Decision: DecisionAllow, ReasonCode: "OK", RiskScore: 10}, nil
}