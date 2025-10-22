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

func Test_RegisterAuthor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		req         *library.RegisterAuthorRequest
		want        *entity.Author
		wantErrCode codes.Code
		wantErr     error
		mocksUsed   bool
	}{
		{
			name: "register author",
			req: &library.RegisterAuthorRequest{
				Name: "Author Name",
			},
			want: &entity.Author{
				ID:   uuid.NewString(),
				Name: "Author Name",
			},
			wantErrCode: codes.OK,
			wantErr:     nil,
			mocksUsed:   true,
		},
		{
			name: "register author | invalid argument",
			req: &library.RegisterAuthorRequest{
				Name: "",
			},
			wantErrCode: codes.InvalidArgument,
			wantErr:     status.Error(codes.InvalidArgument, "error"),
			mocksUsed:   false,
		},
		{
			name: "register author | internal error",
			req: &library.RegisterAuthorRequest{
				Name: "name",
			},
			wantErrCode: codes.Internal,
			wantErr:     status.Error(codes.Internal, "error"),
			mocksUsed:   true,
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
				authorUseCase.EXPECT().RegisterAuthor(ctx, tt.req.GetName()).Return(tt.want, tt.wantErr)
			}

			got, err := service.RegisterAuthor(ctx, tt.req)
			testutils.CheckError(t, err, tt.wantErrCode)
			if err == nil && tt.want != nil {
				assert.Equal(t, tt.want.ID, got.GetId())
			}
		})
	}
}
