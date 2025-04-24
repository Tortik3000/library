package library

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/library"
	"github.com/project/library/internal/usecase/repository"
	"github.com/project/library/internal/usecase/repository/mocks"
)

func TestAddBook(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	book := &entity.Book{
		ID:        "book-id",
		Name:      "Test Book",
		AuthorIDs: []string{"author1", "author2"},
	}
	serialized, _ := json.Marshal(book)
	idempotencyKey := repository.OutboxKindBook.String() + "_" + book.ID

	tests := []struct {
		name                string
		repositoryRerunBook *entity.Book
		returnBook          *entity.Book
		repositoryErr       error
		outboxErr           error
	}{
		{
			name:                "add book",
			repositoryRerunBook: book,
			returnBook:          book,
			repositoryErr:       nil,
			outboxErr:           nil,
		},
		{
			name:                "add book | repository error",
			repositoryRerunBook: nil,
			returnBook:          nil,
			repositoryErr:       entity.ErrAuthorNotFound,
			outboxErr:           nil,
		},
		{
			name:                "add book | outbox error",
			repositoryRerunBook: book,
			returnBook:          nil,
			repositoryErr:       nil,
			outboxErr:           errors.New("cannot send message"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockBooksRepo := mocks.NewMockBooksRepository(ctrl)
			mockOutboxRepo := mocks.NewMockOutboxRepository(ctrl)
			mockTransactor := mocks.NewMockTransactor(ctrl)
			logger, _ := zap.NewProduction()
			useCase := library.New(logger, nil,
				mockBooksRepo, mockOutboxRepo, mockTransactor)
			ctx := t.Context()

			mockTransactor.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(
				func(ctx context.Context, fn func(ctx context.Context) error) error {
					return fn(ctx)
				},
			)
			mockBooksRepo.EXPECT().AddBook(ctx, gomock.Any()).
				Return(tt.repositoryRerunBook, tt.repositoryErr)

			if tt.repositoryErr == nil {
				mockOutboxRepo.EXPECT().SendMessage(ctx, idempotencyKey,
					repository.OutboxKindBook, serialized).Return(tt.outboxErr)
			}

			resultBook, err := useCase.AddBook(ctx, book.Name, book.AuthorIDs)
			switch {
			case tt.outboxErr == nil && tt.repositoryErr == nil:
				require.NoError(t, err)
			case tt.outboxErr != nil:
				require.ErrorIs(t, err, tt.outboxErr)
			case tt.repositoryErr != nil:
				require.ErrorIs(t, err, tt.repositoryErr)
			}

			assert.Equal(t, tt.returnBook, resultBook)
		})
	}
}

func TestGetBook(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	tests := []struct {
		name        string
		returnBook  *entity.Book
		wantErr     error
		wantErrCode codes.Code
	}{
		{
			name: "get book",
			returnBook: &entity.Book{
				ID:        uuid.NewString(),
				Name:      "name",
				AuthorIDs: make([]string, 0),
			},
		},
		{
			name: "get book | with error",
			returnBook: &entity.Book{
				ID:        uuid.NewString(),
				Name:      "name",
				AuthorIDs: make([]string, 0),
			},
			wantErrCode: codes.Internal,
			wantErr:     status.Error(codes.Internal, "error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockBookRepo := mocks.NewMockBooksRepository(ctrl)
			logger, _ := zap.NewProduction()
			useCase := library.New(logger, nil,
				mockBookRepo, nil, nil)
			ctx := t.Context()

			mockBookRepo.EXPECT().GetBook(ctx, tt.returnBook.ID).
				Return(tt.returnBook, tt.wantErr)

			got, err := useCase.GetBook(ctx, tt.returnBook.ID)
			CheckError(t, err, tt.wantErrCode)
			assert.Equal(t, tt.returnBook, got)
		})
	}
}

func TestUpdateBook(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	tests := []struct {
		name        string
		returnBook  *entity.Book
		wantErr     error
		wantErrCode codes.Code
	}{
		{
			name: "update book",
			returnBook: &entity.Book{
				ID:        uuid.NewString(),
				Name:      "name",
				AuthorIDs: make([]string, 0),
			},
		},
		{
			name: "update book | with error",
			returnBook: &entity.Book{
				ID:        uuid.NewString(),
				Name:      "name",
				AuthorIDs: make([]string, 0),
			},
			wantErrCode: codes.NotFound,
			wantErr:     entity.ErrBookNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockBookRepo := mocks.NewMockBooksRepository(ctrl)
			logger, _ := zap.NewProduction()
			useCase := library.New(logger, nil,
				mockBookRepo, nil, nil)
			ctx := t.Context()

			mockBookRepo.EXPECT().UpdateBook(ctx,
				tt.returnBook.ID, tt.returnBook.Name, tt.returnBook.AuthorIDs).
				Return(tt.wantErr)

			err := useCase.UpdateBook(ctx,
				tt.returnBook.ID, tt.returnBook.Name, tt.returnBook.AuthorIDs)
			CheckError(t, err, tt.wantErrCode)
		})
	}
}
