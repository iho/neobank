package pgutil

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func Text(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

func NumericFromString(amount string) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	if err := n.Scan(amount); err != nil {
		return pgtype.Numeric{}, fmt.Errorf("parse numeric %q: %w", amount, err)
	}
	return n, nil
}

func ParseUUID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}