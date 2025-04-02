package repository

import (
	"context"

	"github.com/project/library/internal/entity"
)

//go:generate mockgen_uber -source=interfaces.go -destination=mocks/repository_mock.go -package=mocks

type (
	AuthorRepository interface {
		RegisterAuthor(ctx context.Context, author *entity.Author) (*entity.Author, error)
		GetAuthorInfo(ctx context.Context, authorID string) (*entity.Author, error)
		ChangeAuthor(ctx context.Context, authorID string, newAuthorName string) error
		GetAuthorBooks(ctx context.Context, authorID string) ([]*entity.Book, error)
	}

	BooksRepository interface {
		AddBook(ctx context.Context, book *entity.Book) (*entity.Book, error)
		GetBook(ctx context.Context, bookID string) (*entity.Book, error)
		UpdateBook(ctx context.Context, bookID string, newBookName string, authorIDs []string) error
	}
)
