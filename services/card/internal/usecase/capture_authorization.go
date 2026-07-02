package usecase

import (
	"context"
	"fmt"

	"github.com/iho/neobank/pkg/audit"
	"github.com/iho/neobank/pkg/events"
	"github.com/iho/neobank/pkg/ledgerclient"
	"github.com/iho/neobank/pkg/outbox"
	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/services/card/internal/domain"
	"github.com/iho/neobank/services/card/internal/port"
	"github.com/jackc/pgx/v5"
)

type CaptureAuthorizationInput struct {
	UserID          string
	AuthorizationID string
	SettlementAcct  string
}

type CaptureAuthorizationUseCase struct {
	auths      port.AuthorizationRepository
	ledger     *ledgerclient.Client
	outbox     outbox.TxPublisher
	audit      audit.Recorder
	settlement string
	tx         *pgutil.TxRunner
}

func NewCaptureAuthorizationUseCase(
	auths port.AuthorizationRepository,
	ledger *ledgerclient.Client,
	outboxPublisher outbox.TxPublisher,
	auditRecorder audit.Recorder,
	settlementAccountID string,
	tx *pgutil.TxRunner,
) *CaptureAuthorizationUseCase {
	return &CaptureAuthorizationUseCase{
		auths:      auths,
		ledger:     ledger,
		outbox:     outboxPublisher,
		audit:      auditRecorder,
		settlement: settlementAccountID,
		tx:         tx,
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

	capturedEvent := events.CardAuthCaptured{
		AuthorizationID:  auth.ID,
		CardID:           auth.CardID,
		UserID:           auth.UserID,
		Amount:           auth.Amount,
		Currency:         auth.Currency,
		LedgerTransferID: transfer.Id,
	}
	if err := uc.tx.Run(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if err := uc.auths.WithTx(tx).MarkCaptured(ctx, auth.ID, transfer.Id); err != nil {
			return err
		}
		if err := uc.audit.WithTx(tx).Record(ctx, audit.Entry{
			EntityType: "authorization",
			EntityID:   auth.ID,
			Action:     "captured",
			FromStatus: string(domain.AuthStatusAuthorized),
			ToStatus:   string(domain.AuthStatusCaptured),
			Metadata:   map[string]any{"ledger_transfer_id": transfer.Id},
		}); err != nil {
			return err
		}
		return uc.outbox.WithTx(tx).Publish(ctx, capturedEvent)
	}); err != nil {
		return nil, err
	}

	return uc.auths.GetByID(ctx, auth.ID)
}
