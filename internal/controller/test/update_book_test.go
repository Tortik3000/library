package controller

import (
	"context"

	"testing"

	"github.com/google/uuid"
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

func Test_UpdateBook(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	ctx := t.Context()

	type args struct {
		ctx context.Context
		req *library.UpdateBookRequest
	}
	tests := []struct {
		name        string
		args        args
		wantErrCode codes.Code
		wantErr     error
		mocksUsed   bool
	}{
		{
			name: "successful update",
			args: args{
				ctx: ctx,
				req: &library.UpdateBookRequest{
					Id:        uuid.NewString(),
					Name:      "New Book Name",
					AuthorIds: []string{uuid.NewString()},
				},
			},
			wantErrCode: codes.OK,
			wantErr:     nil,
			mocksUsed:   true,
		},
		{
			name: "invalid request",
			args: args{
				ctx: ctx,
				req: &library.UpdateBookRequest{
					Id:        "1",
					Name:      "name",
					AuthorIds: []string{},
				},
			},
			wantErrCode: codes.InvalidArgument,
			wantErr:     status.Error(codes.InvalidArgument, "error"),
			mocksUsed:   false,
		},
		{
			name: "book not found",
			args: args{
				ctx: ctx,
				req: &library.UpdateBookRequest{
					Id:        uuid.NewString(),
					Name:      "New Book Name",
					AuthorIds: []string{uuid.NewString()},
				},
			},
			wantErrCode: codes.NotFound,
			wantErr:     entity.ErrBookNotFound,
			mocksUsed:   true,
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
				bookUseCase.EXPECT().UpdateBook(ctx, tt.args.req.GetId(),
					tt.args.req.GetName(), tt.args.req.GetAuthorIds()).
					Return(tt.wantErr)
			}

			_, err := service.UpdateBook(ctx, tt.args.req)
			testutils.CheckError(t, err, tt.wantErrCode)
		})
	}
}
