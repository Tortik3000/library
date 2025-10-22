package library

import (
	"context"
	"encoding/json"
	"fmt"

	"go.opentelemetry.io/otel/trace"

	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/repository"
)

func (l *libraryImpl) RegisterAuthor(
	ctx context.Context,
	authorName string,
) (*entity.Author, error) {
	span := trace.SpanFromContext(ctx)
	traceID := span.SpanContext().TraceID().String()
	var author *entity.Author

	err := l.transactor.WithTx(ctx, func(ctx context.Context) error {
		var txErr error
		author, txErr = l.authorRepository.RegisterAuthor(ctx, &entity.Author{
			Name: authorName,
		})
		if txErr != nil {
			return txErr
		}

		serialized, txErr := json.Marshal(author)
		if txErr != nil {
			span.RecordError(fmt.Errorf("error serializing author: %w", txErr))
			return txErr
		}

		idempotencyKey := repository.OutboxKindAuthor.String() + "_" + author.ID
		txErr = l.outboxRepository.SendMessage(
			ctx, idempotencyKey, repository.OutboxKindAuthor, serialized, traceID)
		if txErr != nil {
			return txErr
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

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
