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

func (i *impl) ChangeAuthorInfo(
	ctx context.Context,
	req *library.ChangeAuthorInfoRequest,
) (*library.ChangeAuthorInfoResponse, error) {
	span := trace.SpanFromContext(ctx)
	spanCtx := span.SpanContext()
	span.SetAttributes(attribute.String("author_id", req.GetId()))

	defer span.End()

	log := i.logger.With(
		zap.String("trace_id", spanCtx.TraceID().String()),
		zap.String("span_id", spanCtx.SpanID().String()),
		zap.String("layer", "controller"),
		zap.String("author_id", req.GetId()),
	)

	log.Info("start ChangeAuthorInfo")

	if err := req.ValidateAll(); err != nil {
		log.Warn("invalid data", zap.Error(err))
		span.RecordError(err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err := i.authorUseCase.ChangeAuthor(ctx, req.GetId(), req.GetName())
	if err != nil {
		log.Warn("failed ChangeAuthorInfo", zap.Error(err))
		span.RecordError(err)
		return nil, i.ConvertErr(err)
	}

	log.Info("successfully finished ChangeAuthorInfo")

	return &library.ChangeAuthorInfoResponse{}, nil
}
