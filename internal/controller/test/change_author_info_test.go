package controller

import (
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

	type args struct {
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
			name: "change author info",
			args: args{
				req: &library.ChangeAuthorInfoRequest{
					Id:   uuid.NewString(),
					Name: "New Name",
				},
			},
			wantErrCode: codes.OK,
			wantErr:     nil,
			mocksUsed:   true,
		},
		{
			name: "change author info | with err",
			args: args{
				req: &library.ChangeAuthorInfoRequest{
					Id:   uuid.NewString(),
					Name: "New Name",
				},
			},
			wantErrCode: codes.NotFound,
			wantErr:     entity.ErrAuthorNotFound,
			mocksUsed:   true,
		},
		{
			name: "change author info | with invalid name",
			args: args{
				req: &library.ChangeAuthorInfoRequest{
					Id:   uuid.NewString(),
					Name: "",
				},
			},
			wantErrCode: codes.InvalidArgument,
			wantErr:     status.Error(codes.InvalidArgument, "error"),
			mocksUsed:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			t.Cleanup(ctrl.Finish)

			logger, _ := zap.NewProduction()
			authorUseCase := mocks.NewMockAuthorUseCase(ctrl)
			bookUseCase := mocks.NewMockBooksUseCase(ctrl)
			service := controller.New(logger, bookUseCase, authorUseCase)
			ctx := t.Context()

			if tt.mocksUsed {
				authorUseCase.EXPECT().ChangeAuthor(ctx, tt.args.req.GetId(), tt.args.req.GetName()).
					Return(tt.wantErr)
			}

			_, err := service.ChangeAuthorInfo(ctx, tt.args.req)
			testutils.CheckError(t, err, tt.wantErrCode)
		})
	}
}
