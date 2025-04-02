package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/project/library/internal/entity"
)

var _ AuthorRepository = (*postgresRepository)(nil)
var _ BooksRepository = (*postgresRepository)(nil)

var ErrForeignKeyViolation = &pgconn.PgError{Code: "23503"}

type postgresRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewPostgresRepository(db *pgxpool.Pool, logger *zap.Logger) *postgresRepository {
	return &postgresRepository{
		db:     db,
		logger: logger,
	}
}

func (p *postgresRepository) AddBook(ctx context.Context, book *entity.Book) (*entity.Book, error) {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return &entity.Book{}, err
	}

	defer p.rollbackTx(ctx, tx)

	const insertQueryBook = `
INSERT INTO book (name)
VALUES ($1)
RETURNING id, created_at, updated_at;
`

	id := uuid.UUID{}
	err = tx.QueryRow(ctx, insertQueryBook, book.Name).Scan(&id, &book.CreatedAt, &book.UpdatedAt)
	if err != nil {
		return &entity.Book{}, err
	}
	book.ID = id.String()

	err = bulkInsertInAuthorBook(ctx, book.AuthorIDs, book.ID, tx)
	if err != nil {
		if errors.As(err, &ErrForeignKeyViolation) {
			return &entity.Book{}, entity.ErrAuthorNotFound
		}
		return &entity.Book{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		return &entity.Book{}, err
	}

	return book, nil
}

func (p *postgresRepository) GetBook(ctx context.Context, bookID string) (*entity.Book, error) {
	const GetQueryBook = `
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
	err := p.db.QueryRow(ctx, GetQueryBook, bookID).
		Scan(&book.ID, &book.Name, &book.CreatedAt, &book.UpdatedAt, &authorIDs)
	if errors.Is(err, sql.ErrNoRows) {
		return &entity.Book{}, entity.ErrBookNotFound
	}
	if err != nil {
		return &entity.Book{}, err
	}

	book.AuthorIDs = convertUUIDsToStrings(authorIDs)
	return &book, nil
}

func (p *postgresRepository) UpdateBook(ctx context.Context, bookID string, newBookName string, authorIDs []string) error {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer p.rollbackTx(ctx, tx)

	const UpdateQueryBook = `
UPDATE book SET name = $1 WHERE id = $2;
`

	_, err = tx.Exec(ctx, UpdateQueryBook, newBookName, bookID)
	if err != nil {
		return err
	}

	const insert = `
        INSERT INTO author_book (author_id, book_id)
        SELECT unnest($1::uuid[]), $2
        ON CONFLICT (author_id, book_id) DO NOTHING;
    `
	_, err = tx.Exec(ctx, insert, authorIDs, bookID)
	if err != nil {
		if errors.As(err, &ErrForeignKeyViolation) {
			return entity.ErrAuthorNotFound
		}
		return err
	}

	const DeleteQueryFromAuthorBooks = `
	  DELETE FROM author_book
	  WHERE book_id = $1
	  AND author_id NOT IN (SELECT unnest($2::uuid[]));
	`

	_, err = tx.Exec(ctx, DeleteQueryFromAuthorBooks, bookID, authorIDs)
	if err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (p *postgresRepository) RegisterAuthor(ctx context.Context, author *entity.Author) (*entity.Author, error) {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return &entity.Author{}, err
	}

	defer p.rollbackTx(ctx, tx)

	const InsertQueryAuthor = `
INSERT INTO author ( name)
VALUES ($1)
RETURNING id;
`

	id := uuid.UUID{}
	err = tx.QueryRow(ctx, InsertQueryAuthor, author.Name).Scan(&id)
	if err != nil {
		return &entity.Author{}, err
	}
	author.ID = id.String()

	if err = tx.Commit(ctx); err != nil {
		return &entity.Author{}, err
	}

	return author, nil
}

func (p *postgresRepository) GetAuthorInfo(ctx context.Context, authorID string) (*entity.Author, error) {
	const GetQueryAuthor = `
SELECT id, name
FROM author
WHERE id = $1;
`

	var author entity.Author
	err := p.db.QueryRow(ctx, GetQueryAuthor, authorID).
		Scan(&author.ID, &author.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return &entity.Author{}, entity.ErrAuthorNotFound
	}
	if err != nil {
		return &entity.Author{}, err
	}

	return &author, nil
}

func (p *postgresRepository) ChangeAuthor(ctx context.Context, authorID string, newAuthorName string) error {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer p.rollbackTx(ctx, tx)

	const UpdateQueryAuthor = `
UPDATE author SET name = $1 WHERE id = $2;
`

	_, err = tx.Exec(ctx, UpdateQueryAuthor, newAuthorName, authorID)
	if err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (p *postgresRepository) GetAuthorBooks(ctx context.Context, authorID string) ([]*entity.Book, error) {
	res := make([]*entity.Book, 0)

	const GetQueryBooks = `
SELECT
  book.id,
  book.name,
  book.created_at,
  book.updated_at
FROM author_book
INNER JOIN book
ON author_book.book_id = book.id
WHERE author_book.author_id = $1;
`
	booksRows, err := p.db.Query(ctx, GetQueryBooks, authorID)
	if err != nil {
		return res, err
	}
	defer booksRows.Close()

	for booksRows.Next() {
		var book entity.Book

		if err := booksRows.Scan(&book.ID, &book.Name, &book.CreatedAt, &book.UpdatedAt); err != nil {
			return res, err
		}
		res = append(res, &book)
	}

	for _, book := range res {
		authors, err := p.getBookAuthors(ctx, book.ID)
		if err != nil {
			return res, err
		}
		book.AuthorIDs = authors
	}

	return res, nil
}

func (p *postgresRepository) getBookAuthors(ctx context.Context, bookID string) ([]string, error) {
	const GetQueryAuthors = `
	SELECT array_agg(author_id)
	FROM author_book
	WHERE book_id = $1;
	`
	var authorIDs []uuid.UUID
	err := p.db.QueryRow(ctx, GetQueryAuthors, bookID).Scan(&authorIDs)
	if err != nil {
		return nil, err
	}

	authors := convertUUIDsToStrings(authorIDs)
	return authors, nil
}

func (p *postgresRepository) rollbackTx(ctx context.Context, tx pgx.Tx) {
	err := tx.Rollback(ctx)
	if err != nil {
		p.logger.Debug("failed to rollback transaction", zap.Error(err))
	}
}

func bulkInsertInAuthorBook(ctx context.Context, authorIDs []string, bookID string, tx pgx.Tx) error {
	rows := make([][]interface{}, len(authorIDs))
	for i, authorID := range authorIDs {
		rows[i] = []interface{}{authorID, bookID}
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
