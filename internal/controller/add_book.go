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

func (i *impl) AddBook(
	ctx context.Context,
	req *library.AddBookRequest,
) (*library.AddBookResponse, error) {
	span := trace.SpanFromContext(ctx)
	spanCtx := span.SpanContext()
	defer span.End()

	log := i.logger.With(
		zap.String("layer", "controller"),
		zap.String("span_id", spanCtx.SpanID().String()),
		zap.String("trace_id", spanCtx.TraceID().String()),
	)

	log.Info("start addBook",
		zap.String("book_name", req.GetName()),
	)

	if err := req.ValidateAll(); err != nil {
		log.Warn("invalid data", zap.Error(err))
		span.RecordError(err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	book, err := i.booksUseCase.AddBook(ctx, req.GetName(), req.GetAuthorId())
	if err != nil {
		log.Warn("failed AddBook", zap.Error(err))
		span.RecordError(err)
		return nil, i.ConvertErr(err)
	}

	log.Info("successfully finished AddBook", zap.String("book_id", book.ID))
	span.SetAttributes(attribute.String("book_id", book.ID))

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
