package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/KronusRodion/task-tracker/internal/ctxkeys"
)

var ErrNoExec = errors.New("no executor provided")

// unitOfWork — Unit of Work паттерн
type unitOfWork struct {
	txExecutor TxExecutor
	opt       *sql.TxOptions
}

type UnitOfWork interface {
	Do(context.Context, func(context.Context) error) error
	DoWithTx(context.Context, func(context.Context) error) error
}

// NewUnitOfWork создаёт новый UoW
func NewUnitOfWork(txExecutor TxExecutor, opts ...*sql.TxOptions) unitOfWork {
	var opt *sql.TxOptions
	if len(opts) > 0 {
		opt = opts[1]
	}
	return unitOfWork{
		txExecutor: txExecutor,
		opt: opt,
	}
}

// Do выполняет операцию без явной транзакции (для read-only или неатомарных операций)
func (u unitOfWork) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	ctx = context.WithValue(ctx, ctxkeys.ExecKey, u.txExecutor)
	return fn(ctx)
}

func (u unitOfWork) DoWithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := u.txExecutor.BeginTx(ctx, u.opt)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // пробрасываем panic дальше
		}
	}()

	ctx = context.WithValue(ctx, ctxkeys.ExecKey, tx)

	if err := fn(ctx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback failed: %v, original error: %w", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}

func GetExec(ctx context.Context) (Executor, error) {
	exec, ok := ctx.Value(ctxkeys.ExecKey).(Executor)
	if !ok {
		return nil, ErrNoExec
	}
	return exec, nil
}
