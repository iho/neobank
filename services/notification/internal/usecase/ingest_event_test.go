package usecase_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/events"
	"github.com/iho/neobank/services/notification/internal/domain"
	"github.com/iho/neobank/services/notification/internal/usecase"
)

type fakeNotificationRepo struct {
	created int
}

func (f *fakeNotificationRepo) Create(_ context.Context, _ domain.Notification, _ string) error {
	f.created++
	return nil
}

func (f *fakeNotificationRepo) ListByUser(_ context.Context, _ string, _ int, _ *time.Time, _ string) ([]domain.Notification, error) {
	return nil, nil
}

func (f *fakeNotificationRepo) CountUnread(_ context.Context, _ string) (int64, error) {
	return 0, nil
}

func (f *fakeNotificationRepo) MarkRead(_ context.Context, _, _ string) (domain.Notification, error) {
	return domain.Notification{}, nil
}

func (f *fakeNotificationRepo) MarkAllRead(_ context.Context, _ string) (int64, error) {
	return 0, nil
}

type fakeInbox struct {
	exists  bool
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

func TestIngestEventUseCase_DedupSkipsDispatch(t *testing.T) {
	repo := &fakeNotificationRepo{}
	inbox := &fakeInbox{exists: true}
	uc := usecase.NewIngestEventUseCase(repo, inbox, nil, nil)

	payload, _ := json.Marshal(events.KYCApproved{UserID: uuid.NewString()})
	err := uc.Execute(context.Background(), events.Envelope{
		EventID:   uuid.NewString(),
		EventType: events.TypeKYCApproved,
		Payload:   payload,
	})
	if err != nil {
		t.Fatal(err)
	}
	if repo.created != 0 {
		t.Fatalf("notifications created = %d, want 0", repo.created)
	}
}

func TestIngestEventUseCase_RecordsInboxAfterDispatch(t *testing.T) {
	repo := &fakeNotificationRepo{}
	inbox := &fakeInbox{}
	uc := usecase.NewIngestEventUseCase(repo, inbox, nil, nil)

	userID := uuid.NewString()
	payload, _ := json.Marshal(events.KYCApproved{UserID: userID})
	err := uc.Execute(context.Background(), events.Envelope{
		EventID:   uuid.NewString(),
		EventType: events.TypeKYCApproved,
		Payload:   payload,
	})
	if err != nil {
		t.Fatal(err)
	}
	if repo.created != 1 {
		t.Fatalf("notifications created = %d, want 1", repo.created)
	}
	if !inbox.recorded {
		t.Fatal("expected inbox record")
	}
}