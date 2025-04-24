package library

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/repository"
)

func (l *libraryImpl) RegisterAuthor(
	ctx context.Context,
	authorName string,
) (*entity.Author, error) {
	var author *entity.Author

	err := l.transactor.WithTx(ctx, func(ctx context.Context) error {
		l.logger.Info("Transaction started for RegisterAuthor")

		var txErr error
		author, txErr = l.authorRepository.RegisterAuthor(ctx, &entity.Author{
			Name: authorName,
		})
		if txErr != nil {
			l.logger.Error("Error adding author to repository:",
				zap.Error(txErr))
			return txErr
		}

		serialized, txErr := json.Marshal(author)
		if txErr != nil {
			l.logger.Error("Error serializing author data")
			return txErr
		}

		idempotencyKey := repository.OutboxKindAuthor.String() + "_" + author.ID
		txErr = l.outboxRepository.SendMessage(
			ctx, idempotencyKey, repository.OutboxKindAuthor, serialized)
		if txErr != nil {
			l.logger.Error("Error sending message to outbox",
				zap.Error(txErr))
			return txErr
		}

		return nil
	})
	if err != nil {
		l.logger.Error("Failed to register author",
			zap.Error(err))
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
