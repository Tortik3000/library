package library

import (
	"context"
	"encoding/json"
	"fmt"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"go.uber.org/zap"

	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/repository"
)

func (l *libraryImpl) RegisterAuthor(
	ctx context.Context,
	authorName string,
) (*entity.Author, error) {
	span := trace.SpanFromContext(ctx)
	l.logger.Info("start to register author",
		zap.String("trace_id", span.SpanContext().TraceID().String()),
	)

	var author *entity.Author

	err := l.transactor.WithTx(ctx, func(ctx context.Context) error {
		l.logger.Info("transaction started to register author",
			zap.String("layer", "transaction"),
			zap.String("trace_id", span.SpanContext().TraceID().String()))

		var txErr error
		author, txErr = l.authorRepository.RegisterAuthor(ctx, &entity.Author{
			Name: authorName,
		})
		if txErr != nil {
			span.RecordError(fmt.Errorf("error register author to repostitory: %w", txErr))
			return txErr
		}

		serialized, txErr := json.Marshal(author)
		if txErr != nil {
			span.RecordError(fmt.Errorf("error serializing author: %w", txErr))
			return txErr
		}

		idempotencyKey := repository.OutboxKindAuthor.String() + "_" + author.ID
		txErr = l.outboxRepository.SendMessage(
			ctx, idempotencyKey, repository.OutboxKindAuthor, serialized)
		if txErr != nil {
			span.RecordError(fmt.Errorf("error sending message to outbox: %w", txErr))
			return txErr
		}

		l.logger.Info("transaction finish to register author",
			zap.String("layer", "transaction"),
			zap.String("trace_id", span.SpanContext().TraceID().String()))

		return nil
	})

	if err != nil {
		span.RecordError(fmt.Errorf("error register author to repository: %w", err))
		return nil, err
	}

	span.SetAttributes(attribute.String("author.id", author.ID))
	l.logger.Info("book registered",
		zap.String("layer", "use_case_library"),
		zap.String("trace_id", span.SpanContext().TraceID().String()),
		zap.String("author_id", author.ID),
	)

	return author, nil
}

func (l *libraryImpl) GetAuthorInfo(ctx context.Context,
	authorID string,
) (*entity.Author, error) {
	return l.authorRepository.GetAuthorInfo(ctx, authorID)
}

func (l *libraryImpl) ChangeAuthor(ctx context.Context,
	authorID string,
	newAuthorName string,
) error {
	return l.authorRepository.ChangeAuthor(ctx, authorID, newAuthorName)
}

func (l *libraryImpl) GetAuthorBooks(ctx context.Context,
	authorID string,
) ([]*entity.Book, error) {
	return l.authorRepository.GetAuthorBooks(ctx, authorID)
}
