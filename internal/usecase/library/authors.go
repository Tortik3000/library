package library

import (
	"context"

	"github.com/project/library/internal/entity"
)

func (l *libraryImpl) RegisterAuthor(ctx context.Context, authorName string) (*entity.Author, error) {
	return l.authorRepository.RegisterAuthor(ctx, &entity.Author{
		Name: authorName,
	})
}

func (l *libraryImpl) GetAuthorInfo(ctx context.Context, authorID string) (*entity.Author, error) {
	return l.authorRepository.GetAuthorInfo(ctx, authorID)
}

func (l *libraryImpl) ChangeAuthor(ctx context.Context, authorID string, newAuthorName string) error {
	return l.authorRepository.ChangeAuthor(ctx, authorID, newAuthorName)
}

func (l *libraryImpl) GetAuthorBooks(ctx context.Context, authorID string) ([]*entity.Book, error) {
	return l.authorRepository.GetAuthorBooks(ctx, authorID)
}
