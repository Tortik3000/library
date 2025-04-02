package controller

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/project/library/generated/api/library"
)

func (i *impl) UpdateBook(ctx context.Context, req *library.UpdateBookRequest) (*library.UpdateBookResponse, error) {
	i.logger.Info("Received UpdateBook request",
		zap.String("book name", req.GetName()),
		zap.Strings("author IDs", req.GetAuthorIds()),
		zap.String("book Id", req.GetId()))

	if err := req.ValidateAll(); err != nil {
		i.logger.Error("Invalid UpdateBook request", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err := i.booksUseCase.UpdateBook(ctx, req.GetId(), req.GetName(), req.GetAuthorIds())
	if err != nil {
		i.logger.Error("Failed to update book", zap.Error(err))
		return nil, i.ConvertErr(err)
	}

	return &library.UpdateBookResponse{}, nil
}
