package controller

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/project/library/generated/api/library"
)

func (i *impl) AddBook(ctx context.Context, req *library.AddBookRequest) (*library.AddBookResponse, error) {
	i.logger.Info("Received AddBook request",
		zap.String("book name", req.GetName()),
		zap.Strings("author IDs", req.GetAuthorId()))

	if err := req.ValidateAll(); err != nil {
		i.logger.Error("Invalid AddBook request", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	book, err := i.booksUseCase.AddBook(ctx, req.GetName(), req.GetAuthorId())
	if err != nil {
		i.logger.Error("Failed to add book", zap.Error(err))
		return nil, i.ConvertErr(err)
	}

	return &library.AddBookResponse{
		Book: &library.Book{
			Id:        book.ID,
			Name:      book.Name,
			AuthorId:  book.AuthorIDs,
			CreatedAt: timestamppb.New(book.CreatedAt),
			UpdatedAt: timestamppb.New(book.UpdatedAt),
		},
	}, nil
}
