package pgutil

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type mockBeginner struct {
	tx pgx.Tx
}

func (m *mockBeginner) Begin(ctx context.Context) (pgx.Tx, error) {
	return m.tx, nil
}

type recordingTx struct {
	committed bool
	rollback  func()
}

func (t *recordingTx) Begin(ctx context.Context) (pgx.Tx, error) { return nil, errors.New("nested") }
func (t *recordingTx) Commit(ctx context.Context) error {
	t.committed = true
	return nil
}
func (t *recordingTx) Rollback(ctx context.Context) error {
	if t.rollback != nil {
		t.rollback()
	}
	return nil
}
func (t *recordingTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (t *recordingTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults { return nil }
func (t *recordingTx) LargeObjects() pgx.LargeObjects                               { return pgx.LargeObjects{} }
func (t *recordingTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (t *recordingTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return nil, nil
}
func (t *recordingTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row { return nil }
func (t *recordingTx) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (t *recordingTx) Conn() *pgx.Conn { return nil }

func TestRunInTxRollsBackOnError(t *testing.T) {
	rolledBack := false
	tx := &recordingTx{rollback: func() { rolledBack = true }}
	db := &mockBeginner{tx: tx}

	err := RunInTx(context.Background(), db, func(ctx context.Context, tx pgx.Tx) error {
		return errors.New("boom")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !rolledBack {
		t.Fatal("expected rollback")
	}
	if tx.committed {
		t.Fatal("expected no commit")
	}
}

func TestRunInTxCommitsOnSuccess(t *testing.T) {
	tx := &recordingTx{}
	db := &mockBeginner{tx: tx}

	if err := RunInTx(context.Background(), db, func(ctx context.Context, tx pgx.Tx) error {
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if !tx.committed {
		t.Fatal("expected commit")
	}
}
