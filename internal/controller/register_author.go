package controller

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/project/library/generated/api/library"
)

func (i *impl) RegisterAuthor(
	ctx context.Context,
	req *library.RegisterAuthorRequest,
) (*library.RegisterAuthorResponse, error) {
	i.logger.Info("Received RegisterAuthor request",
		zap.String("author name", req.GetName()))

	if err := req.ValidateAll(); err != nil {
		i.logger.Error("Invalid RegisterAuthor request", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	author, err := i.authorUseCase.RegisterAuthor(ctx, req.GetName())
	if err != nil {
		i.logger.Error("Failed to register author", zap.Error(err))
		return nil, i.ConvertErr(err)
	}

	return &library.RegisterAuthorResponse{
		Id: author.ID,
	}, nil
}
