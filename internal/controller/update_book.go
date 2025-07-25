package controller

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	codes2 "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/project/library/generated/api/library"
)

var (
	UpdateBookDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "library_update_book_duration_ms",
		Help:    "Duration of UpdateBook in ms",
		Buckets: prometheus.DefBuckets,
	})

	UpdateBookRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "library_update_book_requests_total",
		Help: "Total number of UpdateBook requests",
	})
)

func init() {
	prometheus.MustRegister(UpdateBookRequests)
	prometheus.MustRegister(UpdateBookDuration)
}

func (i *impl) UpdateBook(
	ctx context.Context,
	req *library.UpdateBookRequest,
) (*library.UpdateBookResponse, error) {
	UpdateBookRequests.Inc()

	start := time.Now()

	defer func() {
		UpdateBookDuration.Observe(float64(time.Since(start).Milliseconds()))
	}()

	span := trace.SpanFromContext(ctx)
	defer span.End()

	i.logger.Info("start to update book",
		zap.String("layer", "controller"),
		zap.String("book_name", req.GetName()),
		zap.Strings("author_ids", req.GetAuthorIds()),
		zap.String("book_id", req.GetId()),
		zap.String("trace_id", span.SpanContext().TraceID().String()),
	)

	if err := req.ValidateAll(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes2.Code(codes.InvalidArgument),
			"invalid update book request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err := i.booksUseCase.UpdateBook(ctx, req.GetId(),
		req.GetName(), req.GetAuthorIds())
	if err != nil {
		span.RecordError(err)
		return nil, i.ConvertErr(err)
	}

	i.logger.Info("finish to update book",
		zap.String("layer", "controller"),
		zap.String("trace_id", span.SpanContext().TraceID().String()),
	)

	return &library.UpdateBookResponse{}, nil
}
