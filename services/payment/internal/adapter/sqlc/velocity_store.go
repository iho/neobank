package sqlcrepo

import (
	"context"
	"time"

	"github.com/iho/neobank/pkg/fraud"
	"github.com/iho/neobank/pkg/money"
	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/services/payment/internal/gen/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

type PGVelocityStore struct {
	q sqlc.Querier
}

func NewPGVelocityStore(q sqlc.Querier) *PGVelocityStore {
	return &PGVelocityStore{q: q}
}

func (s *PGVelocityStore) RecordAt(userID, amount string, at time.Time) error {
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return err
	}
	numeric, err := pgutil.NumericFromString(amount)
	if err != nil {
		return err
	}
	ctx := context.Background()
	if err := s.q.InsertVelocityEvent(ctx, sqlc.InsertVelocityEventParams{
		UserID:     uid,
		Amount:     numeric,
		RecordedAt: pgtype.Timestamptz{Time: at.UTC(), Valid: true},
	}); err != nil {
		return err
	}
	cutoff := at.UTC().Add(-24 * time.Hour)
	_ = s.q.PruneVelocityEventsOlderThan(ctx, pgtype.Timestamptz{Time: cutoff, Valid: true})
	return nil
}

func (s *PGVelocityStore) CountLastHour(userID string, now time.Time) int {
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return 0
	}
	count, err := s.q.CountVelocityEventsLastHour(context.Background(), sqlc.CountVelocityEventsLastHourParams{
		UserID:     uid,
		RecordedAt: pgtype.Timestamptz{Time: now.UTC().Add(-time.Hour), Valid: true},
	})
	if err != nil {
		return 0
	}
	return int(count)
}

func (s *PGVelocityStore) SumLast24h(userID string, now time.Time) decimal.Decimal {
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return decimal.Zero
	}
	total, err := s.q.SumVelocityEventsLast24h(context.Background(), sqlc.SumVelocityEventsLast24hParams{
		UserID:     uid,
		RecordedAt: pgtype.Timestamptz{Time: now.UTC().Add(-24 * time.Hour), Valid: true},
	})
	if err != nil {
		return decimal.Zero
	}
	amt, err := money.Parse(total)
	if err != nil {
		return decimal.Zero
	}
	return amt
}

var _ fraud.VelocityStore = (*PGVelocityStore)(nil)