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

	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/library"
	"github.com/project/library/internal/usecase/repository"
	"github.com/project/library/internal/usecase/repository/mocks"
)

var defaultAuthor = &entity.Author{
	ID:   uuid.NewString(),
	Name: "name",
}

func TestRegisterAuthor(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	serialized, _ := json.Marshal(defaultAuthor)
	idempotencyKey := repository.OutboxKindAuthor.String() + "_" + defaultAuthor.ID

	tests := []struct {
		name                  string
		repositoryRerunAuthor *entity.Author
		returnAuthor          *entity.Author
		repositoryErr         error
		outboxErr             error
	}{
		{
			name:                  "register author",
			repositoryRerunAuthor: defaultAuthor,
			returnAuthor:          defaultAuthor,
			repositoryErr:         nil,
			outboxErr:             nil,
		},
		{
			name:                  "register author | repository error",
			repositoryRerunAuthor: nil,
			returnAuthor:          nil,
			repositoryErr:         errors.New("error register author"),
			outboxErr:             nil,
		},
		{
			name:                  "register author | outbox error",
			repositoryRerunAuthor: defaultAuthor,
			returnAuthor:          nil,
			repositoryErr:         nil,
			outboxErr:             errors.New("outbox error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockAuthorRepo := mocks.NewMockAuthorRepository(ctrl)
			mockOutboxRepo := mocks.NewMockOutboxRepository(ctrl)
			mockTransactor := mocks.NewMockTransactor(ctrl)
			logger, _ := zap.NewProduction()
			useCase := library.New(logger, mockAuthorRepo,
				nil, mockOutboxRepo, mockTransactor)
			ctx := t.Context()

			mockTransactor.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(
				func(ctx context.Context, fn func(ctx context.Context) error) error {
					return fn(ctx)
				})

			mockAuthorRepo.EXPECT().RegisterAuthor(ctx, gomock.Any()).
				Return(tt.repositoryRerunAuthor, tt.repositoryErr)

			if tt.repositoryErr == nil {
				mockOutboxRepo.EXPECT().SendMessage(
					ctx,
					idempotencyKey,
					repository.OutboxKindAuthor,
					serialized,
					gomock.Any(),
				).Return(tt.outboxErr)
			}
			var (
				resultAuthor *entity.Author
				err          error
			)

			if tt.repositoryRerunAuthor == nil {
				resultAuthor, err = useCase.RegisterAuthor(ctx, defaultAuthor.Name)
			} else {
				resultAuthor, err = useCase.RegisterAuthor(ctx, tt.repositoryRerunAuthor.Name)
			}

			switch {
			case tt.outboxErr == nil && tt.repositoryErr == nil:
				require.NoError(t, err)
			case tt.outboxErr != nil:
				require.ErrorIs(t, err, tt.outboxErr)
			case tt.repositoryErr != nil:
				require.ErrorIs(t, err, tt.repositoryErr)
			}

			assert.Equal(t, tt.returnAuthor, resultAuthor)
		})
	}
}

func TestGetAuthorInfo(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	tests := []struct {
		name                  string
		repositoryRerunAuthor *entity.Author
		wantErr               error
		wantErrCode           codes.Code
	}{
		{
			name:                  "get author info",
			repositoryRerunAuthor: defaultAuthor,
		},
		{
			name:                  "get author info | with error",
			repositoryRerunAuthor: defaultAuthor,
			wantErr:               entity.ErrAuthorNotFound,
			wantErrCode:           codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockAuthorRepo := mocks.NewMockAuthorRepository(ctrl)
			logger, _ := zap.NewProduction()
			useCase := library.New(logger, mockAuthorRepo,
				nil, nil, nil)
			ctx := t.Context()

			mockAuthorRepo.EXPECT().GetAuthorInfo(ctx, tt.repositoryRerunAuthor.ID).Return(tt.repositoryRerunAuthor, tt.wantErr)

			got, wantErr := useCase.GetAuthorInfo(ctx, tt.repositoryRerunAuthor.ID)
			CheckError(t, wantErr, tt.wantErrCode)
			assert.Equal(t, tt.repositoryRerunAuthor, got)
		})
	}
}

func TestChangeAuthor(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	tests := []struct {
		name                  string
		repositoryRerunAuthor *entity.Author
		wantErr               error
		wantErrCode           codes.Code
	}{
		{
			name:                  "change author",
			repositoryRerunAuthor: defaultAuthor,
		},
		{
			name:                  "change author | with error",
			repositoryRerunAuthor: defaultAuthor,
			wantErr:               entity.ErrAuthorNotFound,
			wantErrCode:           codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockAuthorRepo := mocks.NewMockAuthorRepository(ctrl)
			logger, _ := zap.NewProduction()
			useCase := library.New(logger, mockAuthorRepo,
				nil, nil, nil)
			ctx := t.Context()

			mockAuthorRepo.EXPECT().ChangeAuthor(ctx, tt.repositoryRerunAuthor.ID, tt.repositoryRerunAuthor.Name).Return(tt.wantErr)

			wantErr := useCase.ChangeAuthor(ctx, tt.repositoryRerunAuthor.ID, tt.repositoryRerunAuthor.Name)
			CheckError(t, wantErr, tt.wantErrCode)
		})
	}
}

func TestGetAuthorBooks(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	tests := []struct {
		name                  string
		repositoryRerunAuthor *entity.Author
		returnBooks           []*entity.Book
		wantErr               error
		wantErrCode           codes.Code
	}{
		{
			name:                  "get author books",
			repositoryRerunAuthor: defaultAuthor,
			wantErr:               entity.ErrAuthorNotFound,
			wantErrCode:           codes.NotFound,
		},
		{
			name:                  "get author books | with error",
			repositoryRerunAuthor: defaultAuthor,
			returnBooks: []*entity.Book{
				{Name: "first book"},
				{Name: "second book"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockAuthorRepo := mocks.NewMockAuthorRepository(ctrl)
			logger, _ := zap.NewProduction()
			useCase := library.New(logger, mockAuthorRepo,
				nil, nil, nil)
			ctx := t.Context()

			mockAuthorRepo.EXPECT().GetAuthorBooks(ctx, tt.repositoryRerunAuthor.ID).Return(tt.returnBooks, tt.wantErr)

			books, wantErr := useCase.GetAuthorBooks(ctx, tt.repositoryRerunAuthor.ID)
			CheckError(t, wantErr, tt.wantErrCode)
			assert.Equal(t, tt.returnBooks, books)
		})
	}
}
