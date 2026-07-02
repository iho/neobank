package sqlcrepo

import (
	"github.com/iho/neobank/services/payment/internal/gen/sqlc"
	"github.com/jackc/pgx/v5"
)

func withTx(q sqlc.Querier, tx pgx.Tx) sqlc.Querier {
	if queries, ok := q.(*sqlc.Queries); ok {
		return queries.WithTx(tx)
	}
	return q
}
