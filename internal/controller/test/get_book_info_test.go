package controller

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/project/library/generated/api/library"
	"github.com/project/library/internal/controller"
	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/library/mocks"
	testutils "github.com/project/library/internal/usecase/library/test"
)

func Test_GetBookInfo(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	ctx := t.Context()

	bookID := uuid.NewString()
	tests := []struct {
		name        string
		req         *library.GetBookInfoRequest
		want        *entity.Book
		wantErrCode codes.Code
		wantErr     error
		mocksUsed   bool
	}{
		{
			name: "get book info",
			req: &library.GetBookInfoRequest{
				Id: bookID,
			},
			want: &entity.Book{
				ID:        bookID,
				Name:      "Book Name",
				AuthorIDs: []string{uuid.NewString()},
			},
			wantErrCode: codes.OK,
			mocksUsed:   true,
		},
		{
			name: "get book info | not found",
			req: &library.GetBookInfoRequest{
				Id: uuid.NewString(),
			},
			wantErrCode: codes.NotFound,
			wantErr:     entity.ErrBookNotFound,
			mocksUsed:   true,
		},
		{
			name: "get book info | invalid argument",
			req: &library.GetBookInfoRequest{
				Id: "invalid-id",
			},
			wantErrCode: codes.InvalidArgument,
			wantErr:     status.Error(codes.InvalidArgument, "err"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger, _ := zap.NewProduction()
			authorUseCase := mocks.NewMockAuthorUseCase(ctrl)
			bookUseCase := mocks.NewMockBooksUseCase(ctrl)
			service := controller.New(logger, bookUseCase, authorUseCase)

			if tt.mocksUsed {
				bookUseCase.EXPECT().GetBook(ctx, tt.req.GetId()).Return(tt.want, tt.wantErr)
			}

			got, err := service.GetBookInfo(ctx, tt.req)
			testutils.CheckError(t, err, tt.wantErrCode)
			if err == nil && tt.want != nil {
				assert.Equal(t, tt.want.ID, got.GetBook().GetId())
				assert.Equal(t, tt.want.Name, got.GetBook().GetName())
				assert.Equal(t, tt.want.AuthorIDs, got.GetBook().GetAuthorId())
			}
		})
	}
}
