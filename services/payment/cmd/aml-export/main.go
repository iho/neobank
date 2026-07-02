// aml-export prints open CTR/SAR/review cases in a FinCEN-style JSON envelope
// for compliance operators. Intended for cron or manual export, not a long-lived
// process.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/iho/neobank/services/payment/internal/config"
	"github.com/iho/neobank/services/payment/internal/gen/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

type exportEnvelope struct {
	GeneratedAt time.Time    `json:"generated_at"`
	ReportTypes []string     `json:"report_types"`
	Cases       []exportCase `json:"cases"`
}

type exportCase struct {
	CaseID          string    `json:"case_id"`
	EvaluationID    string    `json:"evaluation_id"`
	CaseType        string    `json:"case_type"`
	Status          string    `json:"status"`
	ReasonCode      string    `json:"reason_code"`
	UserID          string    `json:"user_id"`
	EntityType      string    `json:"entity_type"`
	EntityID        string    `json:"entity_id"`
	TransactionType string    `json:"transaction_type"`
	Amount          string    `json:"amount"`
	Currency        string    `json:"currency"`
	Disposition     string    `json:"disposition"`
	RuleSetVersion  string    `json:"rule_set_version"`
	CorrelationID   string    `json:"correlation_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

func main() {
	typesFlag := flag.String("types", "ctr,sar", "comma-separated case types to export (ctr,sar,review)")
	flag.Parse()

	cfg := config.Load()
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "database connect failed: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	types := splitCSV(*typesFlag)
	if len(types) == 0 {
		fmt.Fprintf(os.Stderr, "no case types specified\n")
		os.Exit(1)
	}

	queries := sqlc.New(pool)
	rows, err := queries.ListOpenAMLCasesForExport(ctx, types)
	if err != nil {
		fmt.Fprintf(os.Stderr, "list aml cases: %v\n", err)
		os.Exit(1)
	}

	cases := make([]exportCase, 0, len(rows))
	for _, row := range rows {
		cases = append(cases, exportCase{
			CaseID:          row.ID.String(),
			EvaluationID:    row.EvaluationID.String(),
			CaseType:        row.CaseType,
			Status:          row.Status,
			ReasonCode:      row.ReasonCode,
			UserID:          row.UserID.String(),
			EntityType:      row.EntityType,
			EntityID:        row.EntityID,
			TransactionType: row.TransactionType,
			Amount:          row.Amount,
			Currency:        row.Currency,
			Disposition:     row.Disposition,
			RuleSetVersion:  row.RuleSetVersion,
			CorrelationID:   row.CorrelationID,
			CreatedAt:       row.CreatedAt.Time,
		})
	}

	out := exportEnvelope{
		GeneratedAt: time.Now().UTC(),
		ReportTypes: types,
		Cases:       cases,
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "encode export: %v\n", err)
		os.Exit(1)
	}
}

func splitCSV(s string) []string {
	var out []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			part := s[start:i]
			if part != "" {
				out = append(out, part)
			}
			start = i + 1
		}
	}
	return out
}