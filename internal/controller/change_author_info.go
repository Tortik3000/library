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
	ChangeAuthorDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "library_change_author_duration_ms",
		Help:    "Duration of ChangeAuthorInfo in ms",
		Buckets: prometheus.DefBuckets,
	})

	ChangeAuthorRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "library_change_author_requests_total",
		Help: "Total number of ChangeAuthor requests",
	})
)

func init() {
	prometheus.MustRegister(ChangeAuthorDuration)
	prometheus.MustRegister(ChangeAuthorRequests)
}

func (i *impl) ChangeAuthorInfo(
	ctx context.Context,
	req *library.ChangeAuthorInfoRequest,
) (*library.ChangeAuthorInfoResponse, error) {
	ChangeAuthorRequests.Inc()

	start := time.Now()

	defer func() {
		ChangeAuthorDuration.Observe(float64(time.Since(start).Milliseconds()))
	}()

	span := trace.SpanFromContext(ctx)
	defer span.End()

	i.logger.Info("start to change author info",
		zap.String("layer", "controller"),
		zap.String("author_id", req.GetId()),
		zap.String("author_name", req.GetName()),
		zap.String("trace_id", span.SpanContext().TraceID().String()),
	)

	if err := req.ValidateAll(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes2.Code(codes.InvalidArgument),
			"invalid change author request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err := i.authorUseCase.ChangeAuthor(ctx, req.GetId(), req.GetName())
	if err != nil {
		span.RecordError(err)
		return nil, i.ConvertErr(err)
	}

	i.logger.Info("finish to change author info",
		zap.String("layer", "controller"),
		zap.String("trace_id", span.SpanContext().TraceID().String()),
	)

	return &library.ChangeAuthorInfoResponse{}, nil
}
