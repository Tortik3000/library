package controller

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/project/library/generated/api/library"
	"github.com/project/library/internal/controller"
	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/library/mocks"
	testutils "github.com/project/library/internal/usecase/library/test"
)

type mockLibraryGetAuthorBooksServer struct {
	grpc.ServerStream
	books []*library.Book
	ctx   context.Context
}

func (m *mockLibraryGetAuthorBooksServer) Context() context.Context {
	if m.ctx == nil {
		return context.Background()
	}
	return m.ctx
}

func (m *mockLibraryGetAuthorBooksServer) Send(book *library.Book) error {
	m.books = append(m.books, book)
	return nil
}

type mockLibraryGetAuthorBooksServerErr struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockLibraryGetAuthorBooksServerErr) Context() context.Context {
	if m.ctx == nil {
		return context.Background()
	}
	return m.ctx
}

func (m *mockLibraryGetAuthorBooksServerErr) Send(book *library.Book) error {
	_ = book
	return status.Error(codes.Internal, "some err")
}

func Test_GetAuthorBooks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		req         *library.GetAuthorBooksRequest
		wantErrCode codes.Code
		wantErr     error
		mocksUsed   bool
		server      library.Library_GetAuthorBooksServer
	}{
		{
			name: "get author books | with err not found",
			req: &library.GetAuthorBooksRequest{
				AuthorId: uuid.NewString(),
			},
			wantErrCode: codes.NotFound,
			wantErr:     entity.ErrAuthorNotFound,
			mocksUsed:   true,
			server: &mockLibraryGetAuthorBooksServer{
				ctx: context.Background(),
			},
		},
		{
			name: "get author books | with err invalid args",
			req: &library.GetAuthorBooksRequest{
				AuthorId: "1",
			},
			wantErrCode: codes.InvalidArgument,
			wantErr:     status.Error(codes.InvalidArgument, "error"),
			mocksUsed:   false,
			server: &mockLibraryGetAuthorBooksServer{
				ctx: context.Background(),
			},
		},
		{
			name: "get author books | with err internal",
			req: &library.GetAuthorBooksRequest{
				AuthorId: uuid.NewString(),
			},
			wantErrCode: codes.Internal,
			wantErr:     status.Error(codes.Internal, "error"),
			mocksUsed:   true,
			server: &mockLibraryGetAuthorBooksServerErr{
				ctx: context.Background(),
			},
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

			if tt.mocksUsed {
				authorUseCase.EXPECT().GetAuthorBooks(gomock.Any(), tt.req.GetAuthorId()).Return(nil, tt.wantErr)
			}

			err := service.GetAuthorBooks(tt.req, tt.server)
			testutils.CheckError(t, err, tt.wantErrCode)
		})
	}
}
