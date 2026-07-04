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

type ReverseAuthorizationInput struct {
	AuthorizationID string
	Reason          string
}

// ReverseAuthorizationUseCase releases a hold without ever capturing it —
// the cardproc simulator's "reversal" flow (a merchant voiding a pending
// auth, or a hold expiring unused).
type ReverseAuthorizationUseCase struct {
	auths  port.AuthorizationRepository
	ledger *ledgerclient.Client
	outbox outbox.TxPublisher
	audit  audit.Recorder
	tx     *pgutil.TxRunner
}

func NewReverseAuthorizationUseCase(
	auths port.AuthorizationRepository,
	ledger *ledgerclient.Client,
	outboxPublisher outbox.TxPublisher,
	auditRecorder audit.Recorder,
	tx *pgutil.TxRunner,
) *ReverseAuthorizationUseCase {
	return &ReverseAuthorizationUseCase{
		auths:  auths,
		ledger: ledger,
		outbox: outboxPublisher,
		audit:  auditRecorder,
		tx:     tx,
	}
}

func (uc *ReverseAuthorizationUseCase) Execute(ctx context.Context, in ReverseAuthorizationInput) (*domain.Authorization, error) {
	if in.AuthorizationID == "" {
		return nil, fmt.Errorf("authorization_id is required")
	}

	auth, err := uc.auths.GetByID(ctx, in.AuthorizationID)
	if err != nil {
		return nil, err
	}
	if auth.Status == domain.AuthStatusVoided {
		return auth, nil
	}
	if auth.Status != domain.AuthStatusAuthorized {
		return nil, fmt.Errorf("authorization cannot be reversed")
	}
	if auth.LedgerHoldID == "" {
		return nil, fmt.Errorf("missing ledger hold")
	}
	if uc.ledger == nil {
		return nil, fmt.Errorf("ledger unavailable")
	}

	if err := uc.ledger.VoidHold(ctx, auth.LedgerHoldID); err != nil {
		return nil, fmt.Errorf("void hold: %w", err)
	}

	reason := in.Reason
	if reason == "" {
		reason = "reversed"
	}

	voidedEvent := events.CardAuthVoided{
		AuthorizationID: auth.ID,
		CardID:          auth.CardID,
		UserID:          auth.UserID,
		Amount:          auth.Amount,
		Currency:        auth.Currency,
	}
	if err := uc.tx.Run(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if err := uc.auths.WithTx(tx).MarkVoided(ctx, auth.ID, reason); err != nil {
			return err
		}
		if err := uc.audit.WithTx(tx).Record(ctx, audit.Entry{
			EntityType: "authorization",
			EntityID:   auth.ID,
			Action:     "voided",
			FromStatus: string(domain.AuthStatusAuthorized),
			ToStatus:   string(domain.AuthStatusVoided),
			Metadata:   map[string]any{"reason": reason},
		}); err != nil {
			return err
		}
		return uc.outbox.WithTx(tx).Publish(ctx, voidedEvent)
	}); err != nil {
		return nil, err
	}

	return uc.auths.GetByID(ctx, auth.ID)
}
