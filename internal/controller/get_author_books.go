package controller

import (
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
	span.SetAttributes(attribute.String("author.id", req.GetAuthorId()))

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

	books, err := i.authorUseCase.GetAuthorBooks(server.Context(), req.GetAuthorId())
	if err != nil {
		return i.handleError(span, err, "GetAuthorBooks")
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
			log.Error("failed to send book in stream",
				zap.Error(err),
				zap.String("book_id", book.ID),
			)
			span.RecordError(err)
			return err
		}
	}

	log.Info("successfully finished GetAuthorBooks")

	return nil
}
