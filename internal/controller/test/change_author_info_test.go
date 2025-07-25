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

func Test_ChangeAuthorInfo(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	logger, _ := zap.NewProduction()
	authorUseCase := mocks.NewMockAuthorUseCase(ctrl)
	bookUseCase := mocks.NewMockBooksUseCase(ctrl)
	service := controller.New(logger, bookUseCase, authorUseCase)
	ctx := t.Context()

	type args struct {
		ctx context.Context
		req *library.ChangeAuthorInfoRequest
	}

	tests := []struct {
		name        string
		args        args
		wantErrCode codes.Code
		wantErr     error
		mocksUsed   bool
	}{
		{
			"change author info",
			args{ctx,
				&library.ChangeAuthorInfoRequest{
					Id:   uuid.NewString(),
					Name: "New Name",
				},
			},
			codes.OK,
			nil,
			true,
		},

		{
			"change author info | with err",
			args{ctx,
				&library.ChangeAuthorInfoRequest{
					Id:   uuid.NewString(),
					Name: "New Name",
				},
			},
			codes.NotFound,
			entity.ErrAuthorNotFound,
			true,
		},

		{
			"change author info | with invalid name",
			args{
				ctx,
				&library.ChangeAuthorInfoRequest{
					Id:   uuid.NewString(),
					Name: "",
				},
			},
			codes.InvalidArgument,
			status.Error(codes.InvalidArgument, "error"),
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.mocksUsed {
				authorUseCase.EXPECT().ChangeAuthor(ctx, tt.args.req.GetId(), tt.args.req.GetName()).
					Return(tt.wantErr)
			}

			_, err := service.ChangeAuthorInfo(tt.args.ctx, tt.args.req)

			testutils.CheckError(t, err, tt.wantErrCode)
		})
	}
}
