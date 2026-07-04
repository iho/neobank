package usecase

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestMidRateSameCurrencyIsOne(t *testing.T) {
	rate, err := MidRate("USD", "USD", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !rate.Equal(decimal.NewFromInt(1)) {
		t.Fatalf("expected 1, got %s", rate)
	}
}

func TestMidRateUnsupportedPair(t *testing.T) {
	if _, err := MidRate("USD", "JPY", time.Now()); err == nil {
		t.Fatal("expected error for unsupported pair")
	}
}

func TestMidRateStableWithinBucket(t *testing.T) {
	now := time.Now()

	first, err := MidRate("EUR", "USD", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	second, err := MidRate("EUR", "USD", now.Add(time.Second))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !first.Equal(second) {
		t.Fatalf("expected stable rate within bucket, got %s and %s", first, second)
	}
}

func TestMidRateNearSeed(t *testing.T) {
	rate, err := MidRate("EUR", "USD", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	seed := decimal.NewFromFloat(1.08)
	diff := rate.Sub(seed).Abs()
	maxDrift := seed.Mul(decimal.NewFromFloat(0.01))

	if diff.GreaterThan(maxDrift) {
		t.Fatalf("rate %s drifted too far from seed %s", rate, seed)
	}
}

func TestApplyClientSpreadReducesRate(t *testing.T) {
	mid := decimal.NewFromFloat(1.08)
	withSpread := ApplyClientSpread(mid, 50)

	if !withSpread.LessThan(mid) {
		t.Fatalf("expected spread-adjusted rate %s to be less than mid %s", withSpread, mid)
	}

	expected := mid.Mul(decimal.NewFromFloat(0.995)).Round(8)
	if !withSpread.Equal(expected) {
		t.Fatalf("expected %s, got %s", expected, withSpread)
	}
}

func TestApplyClientSpreadZeroIsMid(t *testing.T) {
	mid := decimal.NewFromFloat(1.08)
	if !ApplyClientSpread(mid, 0).Equal(mid) {
		t.Fatal("expected zero spread to leave rate unchanged")
	}
}
