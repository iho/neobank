package money

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// Parse validates and parses a decimal amount string.
func Parse(amount string) (decimal.Decimal, error) {
	d, err := decimal.NewFromString(amount)
	if err != nil {
		return decimal.Zero, fmt.Errorf("invalid amount %q: %w", amount, err)
	}
	if d.IsNegative() {
		return decimal.Zero, fmt.Errorf("amount must be positive: %s", amount)
	}
	return d, nil
}

// MustParse parses amount or panics (tests only).
func MustParse(amount string) decimal.Decimal {
	d, err := Parse(amount)
	if err != nil {
		panic(err)
	}
	return d
}