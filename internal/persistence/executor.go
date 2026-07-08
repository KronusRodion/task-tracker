package persistence

import (
	"context"
	"database/sql"
)

type TxExecutor interface {
	BeginTx(ctx context.Context,opts *sql.TxOptions) (*sql.Tx, error)
	Executor
}

type Executor interface {
	ExecContext(
		ctx context.Context,
		query string,
		args ...any,
	) (sql.Result, error)

	QueryContext(
		ctx context.Context,
		query string,
		args ...any,
	) (*sql.Rows, error)

	QueryRowContext(
		ctx context.Context,
		query string,
		args ...any,
	) *sql.Row
}
