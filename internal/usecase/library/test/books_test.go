package library

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/library"
	"github.com/project/library/internal/usecase/repository/mocks"
)

func TestAddBook(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockBookRepo := mocks.NewMockBooksRepository(ctrl)
	logger, _ := zap.NewProduction()
	useCase := library.New(logger, nil, mockBookRepo)
	ctx := t.Context()

	tests := []struct {
		name        string
		want        *entity.Book
		wantErr     error
		wantErrCode codes.Code
	}{
		{
			name: "add book",
			want: &entity.Book{
				Name:      "name",
				AuthorIDs: make([]string, 0),
			},
		},
		{
			name: "add book with error",
			want: &entity.Book{
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
			mockBookRepo.EXPECT().AddBook(ctx, tt.want).Return(tt.want, tt.wantErr)

			got, err := useCase.AddBook(ctx, tt.want.Name, tt.want.AuthorIDs)
			CheckError(t, err, tt.wantErrCode)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetBook(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockBookRepo := mocks.NewMockBooksRepository(ctrl)
	logger, _ := zap.NewProduction()
	useCase := library.New(logger, nil, mockBookRepo)
	ctx := t.Context()

	tests := []struct {
		name        string
		want        *entity.Book
		wantErr     error
		wantErrCode codes.Code
	}{
		{
			name: "get book",
			want: &entity.Book{
				ID:        uuid.NewString(),
				Name:      "name",
				AuthorIDs: make([]string, 0),
			},
		},
		{
			name: "get book with error",
			want: &entity.Book{
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
			mockBookRepo.EXPECT().GetBook(ctx, tt.want.ID).Return(tt.want, tt.wantErr)

			got, err := useCase.GetBook(ctx, tt.want.ID)
			CheckError(t, err, tt.wantErrCode)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUpdateBook(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockBookRepo := mocks.NewMockBooksRepository(ctrl)
	logger, _ := zap.NewProduction()
	useCase := library.New(logger, nil, mockBookRepo)
	ctx := t.Context()

	tests := []struct {
		name        string
		want        *entity.Book
		wantErr     error
		wantErrCode codes.Code
	}{
		{
			name: "update book",
			want: &entity.Book{
				ID:        uuid.NewString(),
				Name:      "new name",
				AuthorIDs: make([]string, 0),
			},
		},
		{
			name: "update book with error",
			want: &entity.Book{
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
			mockBookRepo.EXPECT().UpdateBook(
				ctx, tt.want.ID, tt.want.Name, tt.want.AuthorIDs).
				Return(tt.wantErr)

			err := useCase.UpdateBook(ctx, tt.want.ID, tt.want.Name, tt.want.AuthorIDs)
			CheckError(t, err, tt.wantErrCode)
		})
	}
}
