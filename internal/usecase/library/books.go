package library

import (
	"context"

	"github.com/project/library/internal/entity"
)

func (l *libraryImpl) AddBook(ctx context.Context, name string, authorIDs []string) (*entity.Book, error) {
	return l.booksRepository.AddBook(ctx, &entity.Book{
		Name:      name,
		AuthorIDs: authorIDs,
	})
}

func (l *libraryImpl) GetBook(ctx context.Context, bookID string) (*entity.Book, error) {
	return l.booksRepository.GetBook(ctx, bookID)
}

func (l *libraryImpl) UpdateBook(ctx context.Context, bookID string, newBookName string, authorIDs []string) error {
	return l.booksRepository.UpdateBook(ctx, bookID, newBookName, authorIDs)
}
