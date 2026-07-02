package fraud

import "github.com/shopspring/decimal"

const (
	HourlyTransferCountLimit = 10
	DailyTransferAmountLimit = 10000
	HighAmountReviewThreshold = 5000
	P2PReviewThreshold        = 500
	NewAccountReviewThreshold = 500
)

func DailyTransferAmountLimitDecimal() decimal.Decimal {
	return decimal.NewFromInt(DailyTransferAmountLimit)
}

func HighAmountReviewThresholdDecimal() decimal.Decimal {
	return decimal.NewFromInt(HighAmountReviewThreshold)
}