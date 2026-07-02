//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pgutil

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// TxBeginner starts a database transaction.
type TxBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// RunInTx executes fn inside a transaction, rolling back on error.
func RunInTx(ctx context.Context, db TxBeginner, fn func(ctx context.Context, tx pgx.Tx) error) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if err := fn(ctx, tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// TxRunner wraps a TxBeginner for use-case injection.
type TxRunner struct {
	db TxBeginner
}

func NewTxRunner(db TxBeginner) *TxRunner {
	return &TxRunner{db: db}
}

func (r *TxRunner) Run(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	return RunInTx(ctx, r.db, fn)
}
