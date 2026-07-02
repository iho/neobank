package usecase

import (
	"context"
	"fmt"

	"github.com/iho/neobank/pkg/events"
	"github.com/iho/neobank/pkg/ledgerclient"
	"github.com/iho/neobank/services/card/internal/domain"
)

type CaptureAuthorizationInput struct {
	UserID          string
	AuthorizationID string
	SettlementAcct  string
}

type CaptureAuthorizationUseCase struct {
	auths      AuthorizationRepository
	ledger     *ledgerclient.Client
	outbox     OutboxPublisher
	settlement string
}

func NewCaptureAuthorizationUseCase(
	auths AuthorizationRepository,
	ledger *ledgerclient.Client,
	outbox OutboxPublisher,
	settlementAccountID string,
) *CaptureAuthorizationUseCase {
	return &CaptureAuthorizationUseCase{
		auths:      auths,
		ledger:     ledger,
		outbox:     outbox,
		settlement: settlementAccountID,
	}
}

func (uc *CaptureAuthorizationUseCase) Execute(ctx context.Context, in CaptureAuthorizationInput) (*domain.Authorization, error) {
	if in.AuthorizationID == "" {
		return nil, fmt.Errorf("authorization_id is required")
	}

	auth, err := uc.auths.GetByID(ctx, in.AuthorizationID)
	if err != nil {
		return nil, err
	}
	if in.UserID != "" && auth.UserID != in.UserID {
		return nil, fmt.Errorf("authorization not found")
	}
	if auth.Status == domain.AuthStatusCaptured {
		return auth, nil
	}
	if auth.Status != domain.AuthStatusAuthorized {
		return nil, fmt.Errorf("authorization cannot be captured")
	}
	if auth.LedgerHoldID == "" {
		return nil, fmt.Errorf("missing ledger hold")
	}
	if uc.ledger == nil {
		return nil, fmt.Errorf("ledger unavailable")
	}
	if uc.settlement == "" {
		return nil, fmt.Errorf("settlement account not configured")
	}

	transfer, err := uc.ledger.CaptureHold(ctx, ledgerclient.CaptureHoldInput{
		HoldID:      auth.LedgerHoldID,
		ToAccountID: uc.settlement,
	})
	if err != nil {
		return nil, err
	}

	if err := uc.auths.MarkCaptured(ctx, auth.ID, transfer.Id); err != nil {
		return nil, err
	}

	_ = uc.outbox.Publish(ctx, events.CardAuthCaptured{
		AuthorizationID:  auth.ID,
		CardID:           auth.CardID,
		UserID:           auth.UserID,
		Amount:           auth.Amount,
		Currency:         auth.Currency,
		LedgerTransferID: transfer.Id,
	})

	return uc.auths.GetByID(ctx, auth.ID)
}