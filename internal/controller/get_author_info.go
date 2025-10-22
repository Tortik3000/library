package controller

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/project/library/generated/api/library"
)

func (i *impl) GetAuthorInfo(
	ctx context.Context,
	req *library.GetAuthorInfoRequest,
) (*library.GetAuthorInfoResponse, error) {
	span := trace.SpanFromContext(ctx)
	spanCtx := span.SpanContext()
	span.SetAttributes(attribute.String("author.id", req.GetId()))
	defer span.End()

	log := i.logger.With(
		zap.String("trace_id", spanCtx.TraceID().String()),
		zap.String("span_id", spanCtx.SpanID().String()),
		zap.String("layer", "controller"),
		zap.String("author_id", req.GetId()),
	)

	log.Info("start GetAuthorInfo")

	if err := req.ValidateAll(); err != nil {
		log.Warn("invalid data", zap.Error(err))
		span.RecordError(err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	author, err := i.authorUseCase.GetAuthorInfo(ctx, req.GetId())
	if err != nil {
		return nil, i.handleError(span, err, "GetAuthorInfo")
	}

	log.Info("successfully finished GetAuthorInfo")

	return &library.GetAuthorInfoResponse{
		Id:   author.ID,
		Name: author.Name,
	}, nil
}
