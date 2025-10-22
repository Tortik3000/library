package controller

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	service_ "github.com/project/library/internal/controller"
	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/library/mocks"
)

func TestConvertErr(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	logger, _ := zap.NewProduction()
	authorUseCase := mocks.NewMockAuthorUseCase(ctrl)
	bookUseCase := mocks.NewMockBooksUseCase(ctrl)
	service := service_.New(logger, bookUseCase, authorUseCase)

	tests := []struct {
		name     string
		inputErr error
		wantCode codes.Code
	}{
		{
			name:     "AuthorNotFound",
			inputErr: entity.ErrAuthorNotFound,
			wantCode: codes.NotFound,
		},
		{
			name:     "BookNotFound",
			inputErr: entity.ErrBookNotFound,
			wantCode: codes.NotFound,
		},
		{
			name:     "InternalError",
			inputErr: errors.New("some internal error"),
			wantCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := service.ConvertErr(tt.inputErr)
			s, ok := status.FromError(err)
			assert.True(t, ok)
			assert.Equal(t, tt.wantCode, s.Code())
		})
	}
}
