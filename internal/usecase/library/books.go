package library

import (
	"context"
	"encoding/json"
	"fmt"

	"go.opentelemetry.io/otel/trace"

	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/repository"
)

func (l *libraryImpl) AddBook(
	ctx context.Context,
	name string,
	authorIDs []string,
) (*entity.Book, error) {
	span := trace.SpanFromContext(ctx)
	traceID := span.SpanContext().TraceID().String()

	var book *entity.Book

	err := l.transactor.WithTx(ctx, func(ctx context.Context) error {
		var txErr error
		book, txErr = l.booksRepository.AddBook(ctx, &entity.Book{
			Name:      name,
			AuthorIDs: authorIDs,
		})
		if txErr != nil {
			return txErr
		}

		serialized, txErr := json.Marshal(book)
		if txErr != nil {
			span.RecordError(fmt.Errorf("error serializing book: %w", txErr))
			return txErr
		}

		idempotencyKey := repository.OutboxKindBook.String() + "_" + book.ID
		txErr = l.outboxRepository.SendMessage(
			ctx, idempotencyKey, repository.OutboxKindBook, serialized, traceID)
		if txErr != nil {
			return txErr
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return book, nil
}

func (l *libraryImpl) GetBook(
	ctx context.Context,
	bookID string,
) (*entity.Book, error) {
	return l.booksRepository.GetBook(ctx, bookID)
}

func (l *libraryImpl) UpdateBook(
	ctx context.Context,
	bookID string,
	newBookName string,
	authorIDs []string,
) error {
	return l.booksRepository.UpdateBook(ctx, bookID, newBookName, authorIDs)
}
