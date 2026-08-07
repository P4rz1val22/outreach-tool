package repository

import (
	"context"

	db "github.com/P4rz1val22/outreach-tool/db/sqlc"
)

// TxRunner abstracts "run this block of work atomically."
// Production uses real Postgres transactions; tests use a fake that
// just calls the function directly against a mock Querier.
type TxRunner interface {
	RunTx(ctx context.Context, fn func(qtx db.Querier) error) error
}
