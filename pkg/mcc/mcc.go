package mcc

import "strings"

// CategoryLabel returns a human-readable merchant category for a 4-digit MCC code.
func CategoryLabel(code string) string {
	code = strings.TrimSpace(code)
	switch code {
	case "5411", "5422", "5441", "5451", "5462", "5499":
		return "Groceries"
	case "5812", "5813", "5814":
		return "Restaurants"
	case "4111", "4121", "4131":
		return "Transport"
	case "4511", "4722", "7011":
		return "Travel"
	case "5912", "5122", "5911":
		return "Pharmacy"
	case "5732", "5734", "5735":
		return "Electronics"
	case "6011", "6012":
		return "ATM"
	default:
		if code == "" {
			return "Purchase"
		}
		return "Retail"
	}
}