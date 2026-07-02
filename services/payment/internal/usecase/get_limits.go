package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/iho/neobank/pkg/fraud"
	"github.com/iho/neobank/pkg/money"
)

type LimitGauge struct {
	Limit     string
	Used      string
	Remaining string
}

type TransferLimits struct {
	HourlyTransferCount LimitGauge
	DailyTransferAmount LimitGauge
	SingleTransferMax   string
}

type LimitsView struct {
	P2P TransferLimits
}

type GetLimitsUseCase struct {
	velocity fraud.VelocityStore
}

func NewGetLimitsUseCase(velocity fraud.VelocityStore) *GetLimitsUseCase {
	return &GetLimitsUseCase{velocity: velocity}
}

func (uc *GetLimitsUseCase) Execute(ctx context.Context, userID string) (LimitsView, error) {
	if userID == "" {
		return LimitsView{}, fmt.Errorf("user_id is required")
	}
	_ = ctx

	now := time.Now().UTC()
	hourlyUsed := 0
	dailyUsed := "0"
	if uc.velocity != nil {
		hourlyUsed = uc.velocity.CountLastHour(userID, now)
		sum := uc.velocity.SumLast24h(userID, now)
		dailyUsed = sum.StringFixed(2)
	}

	dailyLimit := fraud.DailyTransferAmountLimitDecimal()
	dailyUsedDec, err := money.Parse(dailyUsed)
	if err != nil {
		return LimitsView{}, err
	}
	remainingDaily := dailyLimit.Sub(dailyUsedDec)
	if remainingDaily.IsNegative() {
		remainingDaily = remainingDaily.Abs().Neg()
	}

	hourlyRemaining := fraud.HourlyTransferCountLimit - hourlyUsed
	if hourlyRemaining < 0 {
		hourlyRemaining = 0
	}

	return LimitsView{
		P2P: TransferLimits{
			HourlyTransferCount: LimitGauge{
				Limit:     fmt.Sprintf("%d", fraud.HourlyTransferCountLimit),
				Used:      fmt.Sprintf("%d", hourlyUsed),
				Remaining: fmt.Sprintf("%d", hourlyRemaining),
			},
			DailyTransferAmount: LimitGauge{
				Limit:     dailyLimit.StringFixed(2),
				Used:      dailyUsedDec.StringFixed(2),
				Remaining: remainingDaily.StringFixed(2),
			},
			SingleTransferMax: fmt.Sprintf("%d", fraud.HighAmountReviewThreshold),
		},
	}, nil
}