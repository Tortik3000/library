package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/project/library/metrics"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/project/library/internal/entity"
)

var _ AuthorRepository = (*postgresRepository)(nil)
var _ BooksRepository = (*postgresRepository)(nil)

var ErrForeignKeyViolation = &pgconn.PgError{Code: "23503"}

func measureQueryLatency(operation string, queryFunc func() error) error {
	start := time.Now()
	err := queryFunc()
	metrics.DBQueryLatency.WithLabelValues(operation).Observe(time.Since(start).Seconds())
	return err
}

type postgresRepository struct {
	db     PgxIface
	logger *zap.Logger
}

func NewPostgresRepository(
	db PgxIface,
	logger *zap.Logger,
) *postgresRepository {
	return &postgresRepository{
		db:     db,
		logger: logger,
	}
}

func (p *postgresRepository) AddBook(
	ctx context.Context,
	book *entity.Book,
) (respBook *entity.Book, txErr error) {
	span := trace.SpanFromContext(ctx)
	p.logger.Info("start to add book",
		zap.String("layer", "postgres"),
		zap.String("book_id", book.ID),
		zap.String("trace_id", span.SpanContext().TraceID().String()),
	)

	tx, rollback, err := p.beginTx(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	defer rollback(txErr)

	const insertBook = `
INSERT INTO book (name)
VALUES ($1)
RETURNING id, created_at, updated_at;
`

	id := uuid.UUID{}
	err = measureQueryLatency("insert_book", func() error {
		return tx.QueryRow(ctx, insertBook, book.Name).Scan(
			&id, &book.CreatedAt, &book.UpdatedAt)
	})

	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	book.ID = id.String()

	err = bulkInsertInAuthorBook(ctx, book.AuthorIDs, book.ID, tx)
	if err != nil {
		span.RecordError(err)
		return nil, mapPostgresError(err, entity.ErrAuthorNotFound)
	}

	p.logger.Info("finish to add book",
		zap.String("layer", "postgres"),
		zap.String("trace_id", span.SpanContext().TraceID().String()),
	)

	return book, nil
}

func (p *postgresRepository) GetBook(
	ctx context.Context,
	bookID string,
) (*entity.Book, error) {
	const GetBook = `
SELECT 
  book.id, 
  book.name, 
  book.created_at, 
  book.updated_at, 
  array_agg(author_book.author_id) AS author_ids
FROM 
  book
LEFT JOIN 
  author_book ON book.id = author_book.book_id
WHERE 
  book.id = $1
GROUP BY 
  book.id;
`
	var book entity.Book
	var authorIDs []uuid.UUID

	err := measureQueryLatency("get_book", func() error {
		return p.db.QueryRow(ctx, GetBook, bookID).Scan(&book.ID,
			&book.Name, &book.CreatedAt, &book.UpdatedAt, &authorIDs)
	})

	if err != nil {
		return nil, mapPostgresError(err, entity.ErrBookNotFound)
	}

	book.AuthorIDs = convertUUIDsToStrings(authorIDs)
	return &book, nil
}

func (p *postgresRepository) UpdateBook(
	ctx context.Context,
	bookID string,
	newBookName string,
	authorIDs []string,
) (txErr error) {
	tx, rollback, err := p.beginTx(ctx)
	if err != nil {
		return err
	}
	defer rollback(txErr)

	const UpdateBook = `
UPDATE book SET name = $1 WHERE id = $2;
`

	err = measureQueryLatency("update_book", func() error {
		_, err := tx.Exec(ctx, UpdateBook, newBookName, bookID)
		return err
	})

	if err != nil {
		return err
	}

	const UpdateAuthorBooks = `
WITH inserted AS (
	INSERT INTO author_book (author_id, book_id)
	SELECT unnest($1::uuid[]), $2
	ON CONFLICT (author_id, book_id) DO NOTHING
)
DELETE FROM author_book
WHERE book_id = $2
  AND author_id NOT IN (SELECT unnest($1::uuid[]));
`

	_, err = tx.Exec(ctx, UpdateAuthorBooks, authorIDs, bookID)
	if err != nil {
		return mapPostgresError(err, entity.ErrAuthorNotFound)
	}

	return nil
}

func (p *postgresRepository) RegisterAuthor(
	ctx context.Context,
	author *entity.Author,
) (respAuthor *entity.Author, txErr error) {
	span := trace.SpanFromContext(ctx)

	p.logger.Info("start to register author",
		zap.String("layer", "postgres"),
		zap.String("author_id", author.ID),
		zap.String("trace_id", span.SpanContext().TraceID().String()),
	)

	tx, rollback, err := p.beginTx(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	defer rollback(txErr)

	const InsertAuthor = `
INSERT INTO author ( name)
VALUES ($1)
RETURNING id;
`
	id := uuid.UUID{}
	err = measureQueryLatency("insert_author", func() error {
		return tx.QueryRow(ctx, InsertAuthor, author.Name).Scan(&id)
	})

	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	author.ID = id.String()

	p.logger.Info("finish to register author",
		zap.String("layer", "postgres"),
		zap.String("trace_id", span.SpanContext().TraceID().String()),
	)

	return author, nil
}

func (p *postgresRepository) GetAuthorInfo(
	ctx context.Context,
	authorID string,
) (*entity.Author, error) {
	const GetQueryAuthor = `
SELECT id, name
FROM author
WHERE id = $1;
`

	var author entity.Author

	err := measureQueryLatency("get_author_info", func() error {
		return p.db.QueryRow(ctx, GetQueryAuthor, authorID).
			Scan(&author.ID, &author.Name)
	})

	if err != nil {
		return nil, mapPostgresError(err, entity.ErrAuthorNotFound)
	}

	return &author, nil
}

func (p *postgresRepository) ChangeAuthor(
	ctx context.Context,
	authorID string,
	newAuthorName string,
) (txErr error) {
	tx, rollback, err := p.beginTx(ctx)
	if err != nil {
		return err
	}
	defer rollback(txErr)

	const UpdateAuthor = `
UPDATE author SET name = $1 WHERE id = $2;
`
	err = measureQueryLatency("update_author", func() error {
		_, err = tx.Exec(ctx, UpdateAuthor, newAuthorName, authorID)
		return err
	})

	if err != nil {
		return err
	}

	return nil
}

func (p *postgresRepository) GetAuthorBooks(
	ctx context.Context,
	authorID string,
) ([]*entity.Book, error) {
	const GetBooksWithAuthors = `
SELECT
	book.id,
	book.name,
	book.created_at,
	book.updated_at,
	array_agg(author_book.author_id)
FROM
	book
LEFT JOIN
	author_book ON book.id = author_book.book_id
WHERE
	book.id IN (
		SELECT book_id
		FROM author_book
		WHERE author_id = $1
	)
GROUP BY
	book.id;
`

	rows, err := p.db.Query(ctx, GetBooksWithAuthors, authorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	books := make([]*entity.Book, 0)
	for rows.Next() {
		var book entity.Book
		var authorIDs []uuid.UUID

		if err = rows.Scan(&book.ID, &book.Name, &book.CreatedAt,
			&book.UpdatedAt, &authorIDs); err != nil {
			return nil, err
		}

		book.AuthorIDs = convertUUIDsToStrings(authorIDs)
		books = append(books, &book)
	}

	return books, nil
}

func (p *postgresRepository) beginTx(
	ctx context.Context,
) (pgx.Tx, func(txErr error), error) {
	rollbackFunc := func(error) {}

	tx, err := extractTx(ctx)
	if err != nil {
		tx, err = p.db.Begin(ctx)
		if err != nil {
			return nil, nil, err
		}

		rollbackFunc = func(txErr error) {
			if txErr != nil {
				err := tx.Rollback(ctx)
				if err != nil {
					p.logger.Debug("failed to rollback transaction", zap.Error(err))
				}
				return
			}
			err := tx.Commit(ctx)
			if err != nil {
				p.logger.Debug("failed to commit transaction", zap.Error(err))
			}
		}
	}

	return tx, rollbackFunc, nil
}

func bulkInsertInAuthorBook(
	ctx context.Context,
	authorIDs []string,
	bookID string,
	tx pgx.Tx) error {
	rows := make([][]any, len(authorIDs))
	for i, authorID := range authorIDs {
		rows[i] = []any{authorID, bookID}
	}

	_, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"author_book"},
		[]string{"author_id", "book_id"},
		pgx.CopyFromRows(rows),
	)
	return err
}

func convertUUIDsToStrings(uuids []uuid.UUID) []string {
	strs := make([]string, len(uuids))
	for i, id := range uuids {
		strs[i] = id.String()
	}
	if len(strs) > 0 && strs[0] == uuid.Nil.String() {
		strs = make([]string, 0)
	}
	return strs
}

func mapPostgresError(err error, notFoundErr error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return notFoundErr
	}
	if errors.As(err, &ErrForeignKeyViolation) {
		return notFoundErr
	}
	return err
}
