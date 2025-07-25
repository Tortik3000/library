package controller

import (
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
}

func (m *mockLibraryGetAuthorBooksServer) Send(book *library.Book) error {
	m.books = append(m.books, book)
	return nil
}

type mockLibraryGetAuthorBooksServerErr struct {
	grpc.ServerStream
}

func (m *mockLibraryGetAuthorBooksServerErr) Send(book *library.Book) error {
	_ = book
	return status.Error(codes.Internal, "some err")
}

func Test_GetAuthorBooks(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	logger, _ := zap.NewProduction()
	authorUseCase := mocks.NewMockAuthorUseCase(ctrl)
	bookUseCase := mocks.NewMockBooksUseCase(ctrl)
	service := controller.New(logger, bookUseCase, authorUseCase)

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
			server:      &mockLibraryGetAuthorBooksServer{},
		},
		{
			name: "get author books | with err invalid args",
			req: &library.GetAuthorBooksRequest{
				AuthorId: "1",
			},
			wantErrCode: codes.InvalidArgument,
			wantErr:     status.Error(codes.InvalidArgument, "error"),
			server:      &mockLibraryGetAuthorBooksServer{},
		},
		{
			"get author books | with err internal",
			&library.GetAuthorBooksRequest{
				AuthorId: uuid.NewString(),
			},
			codes.Internal,
			status.Error(codes.Internal, "error"),
			true,
			&mockLibraryGetAuthorBooksServerErr{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.mocksUsed {
				authorUseCase.EXPECT().GetAuthorBooks(gomock.Any(), tt.req.GetAuthorId()).Return(nil, tt.wantErr)
			}

			err := service.GetAuthorBooks(tt.req, tt.server)
			testutils.CheckError(t, err, tt.wantErrCode)
		})
	}
}
