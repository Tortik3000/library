package controller

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/project/library/generated/api/library"
)

func (i *impl) GetBookInfo(
	ctx context.Context,
	req *library.GetBookInfoRequest,
) (*library.GetBookInfoResponse, error) {
	i.logger.Info("Received GetBook request",
		zap.String("book ID", req.GetId()))

	if err := req.ValidateAll(); err != nil {
		i.logger.Error("Invalid GetBook request", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	book, err := i.booksUseCase.GetBook(ctx, req.GetId())
	if err != nil {
		i.logger.Error("Failed to get book", zap.Error(err))
		return nil, i.ConvertErr(err)
	}

	return &library.GetBookInfoResponse{
		Book: &library.Book{
			Id:        book.ID,
			Name:      book.Name,
			AuthorId:  book.AuthorIDs,
			CreatedAt: timestamppb.New(book.CreatedAt),
			UpdatedAt: timestamppb.New(book.UpdatedAt),
		},
	}, nil
}
