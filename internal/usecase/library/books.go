package library

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/repository"
)

func (l *libraryImpl) AddBook(
	ctx context.Context,
	name string,
	authorIDs []string,
) (*entity.Book, error) {
	var book *entity.Book

	err := l.transactor.WithTx(ctx, func(ctx context.Context) error {
		l.logger.Info("Transaction started for AddBook")

		var txErr error
		book, txErr = l.booksRepository.AddBook(ctx, &entity.Book{
			Name:      name,
			AuthorIDs: authorIDs,
		})
		if txErr != nil {
			l.logger.Error("Error adding book to repository:",
				zap.Error(txErr))
			return txErr
		}

		serialized, txErr := json.Marshal(book)
		if txErr != nil {
			l.logger.Error("Error serializing book data")
			return txErr
		}

		idempotencyKey := repository.OutboxKindBook.String() + "_" + book.ID
		txErr = l.outboxRepository.SendMessage(
			ctx, idempotencyKey, repository.OutboxKindBook, serialized)
		if txErr != nil {
			l.logger.Error("Error sending message to outbox",
				zap.Error(txErr))
			return txErr
		}

		return nil
	})

	if err != nil {
		l.logger.Error("Failed to add book", zap.Error(err))
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
