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

func (i *impl) RegisterAuthor(
	ctx context.Context,
	req *library.RegisterAuthorRequest,
) (*library.RegisterAuthorResponse, error) {
	span := trace.SpanFromContext(ctx)
	spanCtx := span.SpanContext()
	defer span.End()

	log := i.logger.With(
		zap.String("trace_id", spanCtx.TraceID().String()),
		zap.String("span_id", spanCtx.SpanID().String()),
		zap.String("layer", "controller"),
	)
	log.Info("start RegisterAuthor")

	if err := req.ValidateAll(); err != nil {
		log.Warn("invalid data", zap.Error(err))
		span.RecordError(err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	author, err := i.authorUseCase.RegisterAuthor(ctx, req.GetName())
	if err != nil {
		log.Warn("failed RegisterAuthor", zap.Error(err))
		span.RecordError(err)
		return nil, i.ConvertErr(err)
	}

	log.Info("successfully finished RegisterAuthor", zap.String("author_id", author.ID))
	span.SetAttributes(attribute.String("author_id", author.ID))

	return &library.RegisterAuthorResponse{
		Id: author.ID,
	}, nil
}
