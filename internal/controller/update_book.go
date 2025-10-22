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

func (i *impl) UpdateBook(
	ctx context.Context,
	req *library.UpdateBookRequest,
) (*library.UpdateBookResponse, error) {
	span := trace.SpanFromContext(ctx)
	spanCtx := span.SpanContext()
	span.SetAttributes(attribute.String("book.id", req.GetId()))

	defer span.End()

	log := i.logger.With(
		zap.String("trace_id", spanCtx.TraceID().String()),
		zap.String("span_id", spanCtx.SpanID().String()),
		zap.String("layer", "controller"),
		zap.String("book_id", req.GetId()),
	)

	if err := req.ValidateAll(); err != nil {
		log.Warn("invalid data", zap.Error(err))
		span.RecordError(err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err := i.booksUseCase.UpdateBook(ctx, req.GetId(),
		req.GetName(), req.GetAuthorIds())
	if err != nil {
		return nil, i.handleError(span, err, "UpdateBook")
	}

	log.Info("successfully finished UpdateBook")

	return &library.UpdateBookResponse{}, nil
}
