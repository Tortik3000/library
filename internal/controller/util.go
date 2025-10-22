package controller

import (
	"errors"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/project/library/internal/entity"
)

func (i *impl) ConvertErr(err error) error {
	switch {
	case errors.Is(err, entity.ErrAuthorNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, entity.ErrBookNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func (i *impl) handleError(
	span trace.Span,
	err error,
	operation string,
) error {
	if err == nil {
		return nil
	}

	log := i.logger.With(
		zap.String("operation", operation),
		zap.Error(err),
	)

	if s, ok := status.FromError(err); ok {
		log = log.With(zap.String("code", s.Code().String()))
		if isClientError(s.Code()) {
			log.Warn("client error")
		} else {
			log.Error("server error")
			span.RecordError(err)
		}
		return err
	}

	switch {
	case errors.Is(err, entity.ErrAuthorNotFound):
		log.Warn("author not found")
		return status.Error(codes.NotFound, "author not found")

	case errors.Is(err, entity.ErrBookNotFound):
		log.Warn("book not found")
		return status.Error(codes.NotFound, "book not found")

	default:
		log.Error("internal error")
		span.RecordError(err)
		return status.Error(codes.Internal, "internal server error")
	}
}

// isClientError проверяет, является ли код ошибки клиентской
func isClientError(code codes.Code) bool {
	switch code {
	case codes.InvalidArgument,
		codes.NotFound,
		codes.AlreadyExists,
		codes.PermissionDenied,
		codes.Unauthenticated,
		codes.FailedPrecondition,
		codes.OutOfRange,
		codes.Unimplemented,
		codes.ResourceExhausted:
		return true
	default:
		return false
	}
}
