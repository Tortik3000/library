package library

import (
	"context"

	"go.uber.org/zap"

	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/repository"
)

//go:generate mockgen_uber -source=interfaces.go -destination=mocks/library_mock.go -package=mocks

var _ AuthorUseCase = (*libraryImpl)(nil)
var _ BooksUseCase = (*libraryImpl)(nil)

type (
	AuthorUseCase interface {
		RegisterAuthor(ctx context.Context, authorName string) (*entity.Author, error)
		GetAuthorInfo(ctx context.Context, authorID string) (*entity.Author, error)
		ChangeAuthor(ctx context.Context, authorID string, newAuthorName string) error
		GetAuthorBooks(ctx context.Context, authorID string) ([]*entity.Book, error)
	}

	BooksUseCase interface {
		AddBook(ctx context.Context, name string, authorIDs []string) (*entity.Book, error)
		GetBook(ctx context.Context, bookID string) (*entity.Book, error)
		UpdateBook(ctx context.Context, bookID string, newBookName string, authorIDs []string) error
	}
)

type libraryImpl struct {
	logger           *zap.Logger
	authorRepository repository.AuthorRepository
	booksRepository  repository.BooksRepository
	outboxRepository repository.OutboxRepository
	transactor       repository.Transactor
}

func New(
	logger *zap.Logger,
	authorRepository repository.AuthorRepository,
	booksRepository repository.BooksRepository,
	outboxRepository repository.OutboxRepository,
	transactor repository.Transactor,
) *libraryImpl {
	return &libraryImpl{
		logger:           logger,
		authorRepository: authorRepository,
		booksRepository:  booksRepository,
		outboxRepository: outboxRepository,
		transactor:       transactor,
	}
}
