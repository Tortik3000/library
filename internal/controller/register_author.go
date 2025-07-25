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
	RegisterAuthorDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "library_register_author_duration_ms",
		Help:    "Duration of RegisterAuthor in ms",
		Buckets: prometheus.DefBuckets,
	})

	RegisterAuthorRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "library_register_author_requests_total",
		Help: "Total number of RegisterAuthor requests",
	})
)

func init() {
	prometheus.MustRegister(RegisterAuthorDuration)
	prometheus.MustRegister(RegisterAuthorRequests)
}

func (i *impl) RegisterAuthor(
	ctx context.Context,
	req *library.RegisterAuthorRequest,
) (*library.RegisterAuthorResponse, error) {
	RegisterAuthorRequests.Inc()
	start := time.Now()

	defer func() {
		AddBookDuration.Observe(float64(time.Since(start).Milliseconds()))
	}()

	span := trace.SpanFromContext(ctx)
	defer span.End()

	i.logger.Info("start to register author",
		zap.String("layer", "controller"),
		zap.String("author_name", req.GetName()),
		zap.String("trace_id", span.SpanContext().TraceID().String()))

	if err := req.ValidateAll(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes2.Code(codes.InvalidArgument), "invalid register author request")

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	author, err := i.authorUseCase.RegisterAuthor(ctx, req.GetName())
	if err != nil {
		span.RecordError(err)
		return nil, i.ConvertErr(err)
	}

	return &library.RegisterAuthorResponse{
		Id: author.ID,
	}, nil
}
