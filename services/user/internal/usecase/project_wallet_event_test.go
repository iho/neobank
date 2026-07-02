package usecase_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/events"
	"github.com/iho/neobank/pkg/walletprojection"
	"github.com/iho/neobank/services/user/internal/domain"
	"github.com/iho/neobank/services/user/internal/usecase"
)

type fakeWalletTxRepo struct {
	inserts int
}

func (f *fakeWalletTxRepo) Insert(_ context.Context, _ walletprojection.Row) error {
	f.inserts++
	return nil
}

func (f *fakeWalletTxRepo) ApplyCapture(_ context.Context, _ walletprojection.CaptureUpdate) error {
	return nil
}

func (f *fakeWalletTxRepo) ListByUser(_ context.Context, _ string, _ int, _ *time.Time, _ string) ([]domain.WalletTransaction, error) {
	return nil, nil
}

func (f *fakeWalletTxRepo) ListByUserInRange(_ context.Context, _ string, _, _ time.Time) ([]domain.WalletTransaction, error) {
	return nil, nil
}

type fakeInbox struct {
	exists   bool
	recorded bool
}

func (f *fakeInbox) Exists(_ context.Context, _ string) (bool, error) {
	return f.exists, nil
}

func (f *fakeInbox) Record(_ context.Context, _, _ string) error {
	f.recorded = true
	f.exists = true
	return nil
}

func TestProjectWalletEventUseCase_DedupSkipsProjection(t *testing.T) {
	repo := &fakeWalletTxRepo{}
	inbox := &fakeInbox{exists: true}
	uc := usecase.NewProjectWalletEventUseCase(repo, inbox)

	payload, _ := json.Marshal(events.TransferCompleted{
		TransferID:      uuid.NewString(),
		SenderUserID:    uuid.NewString(),
		RecipientUserID: uuid.NewString(),
		Amount:          "10.00",
		Currency:        "USD",
	})
	err := uc.Execute(context.Background(), events.Envelope{
		EventID:   uuid.NewString(),
		EventType: events.TypeTransferCompleted,
		Payload:   payload,
	})
	if err != nil {
		t.Fatal(err)
	}
	if repo.inserts != 0 {
		t.Fatalf("inserts = %d, want 0", repo.inserts)
	}
}