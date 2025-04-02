package controller

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/project/library/generated/api/library"
)

func (i *impl) GetAuthorInfo(ctx context.Context, req *library.GetAuthorInfoRequest) (*library.GetAuthorInfoResponse, error) {
	i.logger.Info("Received GetAuthor request",
		zap.String("authorI d", req.GetId()))

	if err := req.ValidateAll(); err != nil {
		i.logger.Error("Invalid GetAuthor request", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	author, err := i.authorUseCase.GetAuthorInfo(ctx, req.GetId())
	if err != nil {
		i.logger.Error("Failed to get author info", zap.Error(err))
		return nil, i.ConvertErr(err)
	}

	return &library.GetAuthorInfoResponse{
		Id:   author.ID,
		Name: author.Name,
	}, nil
}
