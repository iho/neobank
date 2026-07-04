package usecase

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/iho/neobank/services/simulators/cardproc/internal/domain"
	"github.com/iho/neobank/services/simulators/cardproc/internal/port"
)

type IssueCardInput struct {
	ExternalRef    string
	CardholderName string
}

// IssueCardUseCase mints a new virtual card; the card service calls this
// once per card issuance (see services/card/internal/adapter/processor/httpclient.go).
type IssueCardUseCase struct {
	cards port.CardRepository
}

func NewIssueCardUseCase(cards port.CardRepository) *IssueCardUseCase {
	return &IssueCardUseCase{cards: cards}
}

func (uc *IssueCardUseCase) Execute(ctx context.Context, in IssueCardInput) (domain.Card, error) {
	if in.ExternalRef == "" {
		return domain.Card{}, fmt.Errorf("external_ref is required")
	}

	panToken, lastFour, err := generatePAN()
	if err != nil {
		return domain.Card{}, err
	}

	now := time.Now().UTC()

	return uc.cards.Create(ctx, in.ExternalRef, in.CardholderName, panToken, lastFour, int(now.Month()), now.Year()+3)
}

// generatePAN produces a token and last-four for simulation purposes only —
// not a real PAN.
func generatePAN() (token, lastFour string, err error) {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("generate pan: %w", err)
	}

	token = fmt.Sprintf("tok_sim_%x", buf)
	lastFour = fmt.Sprintf("%04d", (int(buf[0])<<8|int(buf[1]))%10000)

	return token, lastFour, nil
}
