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

func Test_GetAuthorInfo(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	ctx := t.Context()

	authorID := uuid.NewString()
	tests := []struct {
		name        string
		req         *library.GetAuthorInfoRequest
		want        *entity.Author
		wantErrCode codes.Code
		wantErr     error
		mocksUsed   bool
	}{
		{
			name: "get author info",
			req: &library.GetAuthorInfoRequest{
				Id: authorID,
			},
			want: &entity.Author{
				Name: "name",
				ID:   authorID,
			},
			wantErrCode: codes.OK,
			mocksUsed:   true,
		},
		{
			name: "get author info | not found",
			req: &library.GetAuthorInfoRequest{
				Id: uuid.NewString(),
			},
			wantErrCode: codes.NotFound,
			wantErr:     entity.ErrAuthorNotFound,
			mocksUsed:   true,
		},
		{
			name: "get author info | invalid argument",
			req: &library.GetAuthorInfoRequest{
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
				authorUseCase.EXPECT().GetAuthorInfo(ctx, tt.req.GetId()).Return(tt.want, tt.wantErr)
			}

			got, err := service.GetAuthorInfo(ctx, tt.req)
			testutils.CheckError(t, err, tt.wantErrCode)
			if err == nil && tt.want != nil {
				assert.Equal(t, tt.want.Name, got.GetName())
				assert.Equal(t, tt.want.ID, got.GetId())
			}
		})
	}
}
