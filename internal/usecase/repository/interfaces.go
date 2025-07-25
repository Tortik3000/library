package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/project/library/internal/entity"
)

//go:generate mockgen_uber -source=interfaces.go -destination=mocks/repository_mock.go -package=mocks

type (
	PgxIface interface {
		Begin(context.Context) (pgx.Tx, error)
		Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
		Query(context.Context, string, ...interface{}) (pgx.Rows, error)
		QueryRow(context.Context, string, ...interface{}) pgx.Row
	}

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

	Transactor interface {
		WithTx(ctx context.Context, function func(ctx context.Context) error) error
	}

	OutboxRepository interface {
		SendMessage(ctx context.Context, idempotencyKey string, kind OutboxKind, message []byte) error
		GetMessages(ctx context.Context, batchSize int, inProgressTTL time.Duration) ([]OutboxData, error)
		MarkAsProcessed(ctx context.Context, idempotencyKeys []string) error
	}

	OutboxData struct {
		IdempotencyKey string
		Kind           OutboxKind
		RawData        []byte
	}
)

type OutboxKind int

const (
	OutboxKindUndefined OutboxKind = iota
	OutboxKindBook
	OutboxKindAuthor
)

func (o OutboxKind) String() string {
	switch o {
	case OutboxKindBook:
		return "book"
	case OutboxKindAuthor:
		return "author"
	default:
		return "undefined"
	}
}
