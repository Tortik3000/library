package controller

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/project/library/generated/api/library"
)

func (i *impl) GetAuthorBooks(
	req *library.GetAuthorBooksRequest,
	server library.Library_GetAuthorBooksServer,
) error {
	i.logger.Info("Received GetAuthorBooks request",
		zap.String("author Id", req.GetAuthorId()))

	if err := req.ValidateAll(); err != nil {
		i.logger.Error("Invalid GetAuthorBooks request", zap.Error(err))
		return status.Error(codes.InvalidArgument, err.Error())
	}

	books, err := i.authorUseCase.GetAuthorBooks(
		context.Background(), req.GetAuthorId())
	if err != nil {
		i.logger.Error("Failed to get author books", zap.Error(err))
		return i.ConvertErr(err)
	}

	for _, book := range books {
		err := server.Send(&library.Book{
			Id:        book.ID,
			Name:      book.Name,
			AuthorId:  book.AuthorIDs,
			CreatedAt: timestamppb.New(book.CreatedAt),
			UpdatedAt: timestamppb.New(book.UpdatedAt),
		})
		if err != nil {
			return err
		}
	}

	return nil
}
