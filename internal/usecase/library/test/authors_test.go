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

func TestRegisterAuthor(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockAuthorRepo := mocks.NewMockAuthorRepository(ctrl)
	logger, _ := zap.NewProduction()
	useCase := library.New(logger, mockAuthorRepo, nil)
	ctx := t.Context()

	tests := []struct {
		name        string
		want        *entity.Author
		wantErr     error
		wantErrCode codes.Code
	}{
		{
			name: "register author",
			want: &entity.Author{
				ID:   uuid.NewString(),
				Name: "name",
			},
		},
		{
			name: "register author with error",
			want: &entity.Author{
				Name: "name",
			},
			wantErrCode: codes.Internal,
			wantErr:     status.Error(codes.Internal, "error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockAuthorRepo.EXPECT().RegisterAuthor(ctx, gomock.Any()).Return(tt.want, tt.wantErr)

			got, err := useCase.RegisterAuthor(ctx, tt.want.Name)
			CheckError(t, err, tt.wantErrCode)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetAuthorInfo(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockAuthorRepo := mocks.NewMockAuthorRepository(ctrl)
	logger, _ := zap.NewProduction()
	useCase := library.New(logger, mockAuthorRepo, nil)
	ctx := t.Context()

	tests := []struct {
		name        string
		want        *entity.Author
		wantErr     error
		wantErrCode codes.Code
	}{
		{
			name: "get author info",
			want: &entity.Author{
				ID:   uuid.NewString(),
				Name: "name",
			},
		},
		{
			name: "get author info with error",
			want: &entity.Author{
				ID:   uuid.NewString(),
				Name: "name",
			},
			wantErr:     entity.ErrAuthorNotFound,
			wantErrCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockAuthorRepo.EXPECT().GetAuthorInfo(ctx, tt.want.ID).Return(tt.want, tt.wantErr)

			got, err := useCase.GetAuthorInfo(ctx, tt.want.ID)
			CheckError(t, err, tt.wantErrCode)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestChangeAuthor(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockAuthorRepo := mocks.NewMockAuthorRepository(ctrl)
	logger, _ := zap.NewProduction()
	useCase := library.New(logger, mockAuthorRepo, nil)
	ctx := t.Context()

	tests := []struct {
		name        string
		want        *entity.Author
		wantErr     error
		wantErrCode codes.Code
	}{
		{
			name: "change author",
			want: &entity.Author{
				ID:   uuid.NewString(),
				Name: "new name",
			},
		},
		{
			name: "change author",
			want: &entity.Author{
				ID:   uuid.NewString(),
				Name: "new name",
			},
			wantErr:     entity.ErrAuthorNotFound,
			wantErrCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockAuthorRepo.EXPECT().ChangeAuthor(ctx, tt.want.ID, tt.want.Name).Return(tt.wantErr)

			err := useCase.ChangeAuthor(ctx, tt.want.ID, tt.want.Name)
			CheckError(t, err, tt.wantErrCode)
		})
	}
}

func TestGetAuthorBooks(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockAuthorRepo := mocks.NewMockAuthorRepository(ctrl)
	logger, _ := zap.NewProduction()
	useCase := library.New(logger, mockAuthorRepo, nil)
	ctx := t.Context()

	tests := []struct {
		name        string
		want        *entity.Author
		wantBooks   []*entity.Book
		wantErr     error
		wantErrCode codes.Code
	}{
		{
			name: "get author books",
			want: &entity.Author{
				ID:   uuid.NewString(),
				Name: "new name",
			},
			wantErr:     entity.ErrAuthorNotFound,
			wantErrCode: codes.NotFound,
		},
		{
			name: "get author books with err",
			want: &entity.Author{
				ID:   uuid.NewString(),
				Name: "new name",
			},
			wantBooks: []*entity.Book{
				{Name: "first book"},
				{Name: "second book"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockAuthorRepo.EXPECT().GetAuthorBooks(ctx, tt.want.ID).Return(tt.wantBooks, tt.wantErr)

			books, err := useCase.GetAuthorBooks(ctx, tt.want.ID)
			CheckError(t, err, tt.wantErrCode)
			assert.Equal(t, tt.wantBooks, books)
		})
	}
}
