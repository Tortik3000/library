package controller

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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

	log.Info("start GetBookInfo")

	if err := req.ValidateAll(); err != nil {
		log.Warn("invalid data", zap.Error(err))
		span.RecordError(err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	book, err := i.booksUseCase.GetBook(ctx, req.GetId())
	if err != nil {
		return nil, i.handleError(span, err, "GetBookInfo")
	}

	log.Info("successfully finished GetBookInfo")

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
