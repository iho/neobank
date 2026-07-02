package processor

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type VirtualCard struct {
	Ref         string
	PANToken    string
	LastFour    string
	ExpiryMonth int
	ExpiryYear  int
}

type Processor interface {
	CreateVirtualCard(ctx context.Context, userID, cardholderName string) (VirtualCard, error)
	CancelCard(ctx context.Context, ref string) error
}

type MockProcessor struct{}

func NewMock() *MockProcessor {
	return &MockProcessor{}
}

func (m *MockProcessor) CreateVirtualCard(_ context.Context, userID, _ string) (VirtualCard, error) {
	ref := uuid.NewString()
	lastFour := fmt.Sprintf("%04d", time.Now().UnixNano()%10000)
	now := time.Now().UTC()
	return VirtualCard{
		Ref:         ref,
		PANToken:    fmt.Sprintf("tok_%s_%s", userID[:8], ref[:8]),
		LastFour:    lastFour,
		ExpiryMonth: int(now.Month()),
		ExpiryYear:  now.Year() + 3,
	}, nil
}

func (m *MockProcessor) CancelCard(_ context.Context, _ string) error {
	return nil
}