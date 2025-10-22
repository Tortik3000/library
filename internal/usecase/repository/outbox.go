package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/project/library/internal/metrics"
	"go.uber.org/zap"
)

var _ OutboxRepository = (*outboxRepository)(nil)

type outboxRepository struct {
	db     PgxIface
	logger *zap.Logger
}

func NewOutbox(db PgxIface, logger *zap.Logger) *outboxRepository {
	return &outboxRepository{
		db:     db,
		logger: logger,
	}
}

func (o *outboxRepository) SendMessage(
	ctx context.Context,
	idempotencyKey string,
	kind OutboxKind,
	message []byte,
	traceID string,
) error {
	const query = `
INSERT INTO outbox (idempotency_key, data, status, kind, trace_id)
VALUES($1, $2, 'CREATED', $3, $4)
ON CONFLICT (idempotency_key) DO NOTHING`

	var err error
	if tx, txErr := extractTx(ctx); txErr == nil {
		_, err = tx.Exec(ctx, query, idempotencyKey, message, kind, traceID)
	} else {
		_, err = o.db.Exec(ctx, query, idempotencyKey, message, kind, traceID)
	}

	if err != nil {
		return err
	}

	metrics.OutboxTasksCreated.WithLabelValues(kind.String()).Inc()

	return nil
}

func (o *outboxRepository) GetMessages(
	ctx context.Context, batchSize int,
	inProgressTTL time.Duration,
) ([]OutboxData, error) {
	const query = `
UPDATE outbox
SET status = 'IN_PROGRESS'
WHERE idempotency_key IN (
    SELECT idempotency_key
    FROM outbox
    WHERE
        (status = 'CREATED'
            OR (status = 'IN_PROGRESS' AND updated_at < now() - $1::interval))
    ORDER BY created_at
    LIMIT $2
    FOR UPDATE SKIP LOCKED
	)
	RETURNING idempotency_key, data, kind, trace_id;`

	interval := fmt.Sprintf("%d ms", inProgressTTL.Milliseconds())
	var (
		err  error
		rows pgx.Rows
	)
	tx, txErr := extractTx(ctx)
	if txErr == nil {
		rows, err = tx.Query(ctx, query, interval, batchSize)
	} else {
		rows, err = o.db.Query(ctx, query, interval, batchSize)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := make([]OutboxData, 0)

	for rows.Next() {
		var key string
		var rawData []byte
		var kind OutboxKind
		var traceID string

		if err := rows.Scan(&key, &rawData, &kind, &traceID); err != nil {
			return nil, err
		}

		result = append(result, OutboxData{
			IdempotencyKey: key,
			RawData:        rawData,
			Kind:           kind,
			TraceID:        traceID,
		})
	}

	return result, rows.Err()
}

func (o *outboxRepository) MarkAsProcessed(
	ctx context.Context,
	idempotencyKeys []string,
) error {
	if len(idempotencyKeys) == 0 {
		return nil
	}

	const query = `
UPDATE outbox
SET status = 'SUCCESS'
WHERE idempotency_key = ANY($1);
`

	var err error
	if tx, txErr := extractTx(ctx); txErr == nil {
		_, err = tx.Exec(ctx, query, idempotencyKeys)
	} else {
		_, err = o.db.Exec(ctx, query, idempotencyKeys)
	}

	if err != nil {
		return err
	}

	return nil
}
