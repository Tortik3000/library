package controller

import (
	"context"
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

func Test_AddBook(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	ctx := t.Context()

	type args struct {
		ctx context.Context
		req *library.AddBookRequest
	}

	tests := []struct {
		name        string
		args        args
		want        *entity.Book
		wantErrCode codes.Code
		wantErr     error
		mocksUsed   bool
	}{
		{
			name: "add book",
			args: args{ctx,
				&library.AddBookRequest{
					Name:     "book",
					AuthorId: make([]string, 0),
				},
			},
			want: &entity.Book{
				ID:        uuid.NewString(),
				Name:      "book",
				AuthorIDs: make([]string, 0),
			},
			wantErrCode: codes.OK,
			mocksUsed:   true,
		},

		{
			name: "add book | with err",
			args: args{ctx,
				&library.AddBookRequest{
					Name:     "book",
					AuthorId: make([]string, 0),
				},
			},
			wantErrCode: codes.NotFound,
			wantErr:     entity.ErrAuthorNotFound,
			mocksUsed:   true,
		},

		{
			name: "add book | with invalid authors",
			args: args{
				ctx,
				&library.AddBookRequest{
					Name:     "book",
					AuthorId: []string{"1"},
				},
			},

			wantErrCode: codes.InvalidArgument,
			wantErr:     status.Error(codes.InvalidArgument, "error"),
		},

		{
			name: "add book | with invalid name",
			args: args{
				ctx,
				&library.AddBookRequest{
					Name:     "",
					AuthorId: make([]string, 0),
				},
			},
			wantErrCode: codes.InvalidArgument,
			wantErr:     status.Error(codes.InvalidArgument, "error"),
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
				bookUseCase.EXPECT().AddBook(ctx, tt.args.req.GetName(),
					tt.args.req.GetAuthorId()).
					Return(tt.want, tt.wantErr)
			}

			got, err := service.AddBook(tt.args.ctx, tt.args.req)

			testutils.CheckError(t, err, tt.wantErrCode)
			if err == nil && tt.want != nil {
				assert.Equal(t, tt.want.ID, got.GetBook().GetId())
				assert.Equal(t, tt.want.Name, got.GetBook().GetName())
				assert.Equal(t, tt.want.AuthorIDs, got.GetBook().GetAuthorId())
			}
		})
	}
}
