package usecase

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"time"
)

type ExportWalletTransactionsInput struct {
	UserID string
	From   time.Time
	To     time.Time
	Format string
}

type ExportWalletTransactionsUseCase struct {
	repo WalletTransactionRepository
}

func NewExportWalletTransactionsUseCase(repo WalletTransactionRepository) *ExportWalletTransactionsUseCase {
	return &ExportWalletTransactionsUseCase{repo: repo}
}

func (uc *ExportWalletTransactionsUseCase) Execute(ctx context.Context, in ExportWalletTransactionsInput) ([]byte, string, error) {
	if in.UserID == "" {
		return nil, "", fmt.Errorf("user_id is required")
	}
	if in.Format != "" && in.Format != "csv" {
		return nil, "", fmt.Errorf("unsupported format %q", in.Format)
	}
	if in.From.IsZero() || in.To.IsZero() {
		return nil, "", fmt.Errorf("from and to are required")
	}
	if !in.To.After(in.From) {
		return nil, "", fmt.Errorf("to must be after from")
	}

	rows, err := uc.repo.ListByUserInRange(ctx, in.UserID, in.From.UTC(), in.To.UTC())
	if err != nil {
		return nil, "", err
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"date", "type", "amount", "currency", "direction", "status", "counterparty", "memo"})
	for _, row := range rows {
		_ = w.Write([]string{
			row.CreatedAt.UTC().Format(time.RFC3339),
			row.Type,
			row.Amount,
			row.Currency,
			row.Direction,
			row.Status,
			row.Counterparty,
			row.Memo,
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), "text/csv", nil
}