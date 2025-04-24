package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ Transactor = (*transactor)(nil)

type transactor struct {
	db     PgxIface
	logger *zap.Logger
}

func NewTransactor(db PgxIface, logger *zap.Logger) *transactor {
	return &transactor{
		db:     db,
		logger: logger,
	}
}

type txInjector struct{}

var ErrTxNotFound = status.Error(codes.NotFound,
	"tx not found in ctx")

func (t transactor) WithTx(
	ctx context.Context,
	function func(ctx context.Context) error,
) (txErr error) {
	ctxWithTx, tx, err := injectTx(ctx, t.db)
	if err != nil {
		return fmt.Errorf(
			"can not inject transaction, error: %w", err)
	}

	defer func() {
		if txErr != nil {
			err = tx.Rollback(ctxWithTx)
			if err != nil {
				t.logger.Error("failed to rollback transaction", zap.Error(err))
			}
			return
		}

		err = tx.Commit(ctxWithTx)
		if err != nil {
			t.logger.Error("failed to commit transaction", zap.Error(err))
		}
	}()

	err = function(ctxWithTx)
	if err != nil {
		return fmt.Errorf("function execution error: %w", err)
	}

	return nil
}

func injectTx(
	ctx context.Context,
	pool PgxIface,
) (context.Context, pgx.Tx, error) {
	if tx, err := extractTx(ctx); err == nil {
		return ctx, tx, nil
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}

	return context.WithValue(ctx, txInjector{}, tx), tx, nil
}

func extractTx(ctx context.Context) (pgx.Tx, error) {
	tx, ok := ctx.Value(txInjector{}).(pgx.Tx)

	if !ok {
		return nil, ErrTxNotFound
	}

	return tx, nil
}
