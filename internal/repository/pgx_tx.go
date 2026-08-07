package repository

import (
	"context"

	db "github.com/P4rz1val22/outreach-tool/db/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgxTxRunner struct {
	Pool *pgxpool.Pool
}

func (r *PgxTxRunner) RunTx(ctx context.Context, fn func(qtx db.Querier) error) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := db.New(tx)

	if err := fn(qtx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
