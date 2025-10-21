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

func (i *impl) GetAuthorBooks(
	req *library.GetAuthorBooksRequest,
	server library.Library_GetAuthorBooksServer,
) error {
	span := trace.SpanFromContext(server.Context())
	spanCtx := span.SpanContext()
	span.SetAttributes(attribute.String("author_id", req.GetAuthorId()))

	defer span.End()

	log := i.logger.With(
		zap.String("trace_id", spanCtx.TraceID().String()),
		zap.String("span_id", spanCtx.SpanID().String()),
		zap.String("layer", "controller"),
		zap.String("author_id", req.GetAuthorId()),
	)

	log.Info("start GetAuthorBooks")

	if err := req.ValidateAll(); err != nil {
		log.Warn("invalid data", zap.Error(err))
		span.RecordError(err)
		return status.Error(codes.InvalidArgument, err.Error())
	}

	books, err := i.authorUseCase.GetAuthorBooks(
		context.Background(), req.GetAuthorId())
	if err != nil {
		log.Warn("failed GetAuthorBooks", zap.Error(err))
		span.RecordError(err)
		return i.ConvertErr(err)
	}

	for _, book := range books {
		err = server.Send(&library.Book{
			Id:        book.ID,
			Name:      book.Name,
			AuthorId:  book.AuthorIDs,
			CreatedAt: timestamppb.New(book.CreatedAt),
			UpdatedAt: timestamppb.New(book.UpdatedAt),
		})
		if err != nil {
			log.Warn("failed to insert record in stream", zap.Error(err))
			span.RecordError(err)
			return err
		}
	}

	log.Info("successfully finished GetAuthorBooks")

	return nil
}
