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
	GetAuthorInfoDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "library_get_author_duration_ms",
		Help:    "Duration of GetAuthorInfo in ms",
		Buckets: prometheus.DefBuckets,
	})

	GetAuthorInfoRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "library_get_author_requests_total",
		Help: "Total number of GetAuthorInfo requests",
	})
)

func init() {
	prometheus.MustRegister(GetAuthorInfoDuration)
	prometheus.MustRegister(GetAuthorInfoRequests)
}

func (i *impl) GetAuthorInfo(
	ctx context.Context,
	req *library.GetAuthorInfoRequest,
) (*library.GetAuthorInfoResponse, error) {
	GetAuthorInfoRequests.Inc()
	start := time.Now()

	defer func() {
		GetAuthorInfoDuration.Observe(float64(time.Since(start).Milliseconds()))
	}()

	span := trace.SpanFromContext(ctx)
	defer span.End()

	i.logger.Info("start to get author info",
		zap.String("layer", "controller"),
		zap.String("author_id", req.Id),
		zap.String("trace_id", span.SpanContext().TraceID().String()),
	)

	if err := req.ValidateAll(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes2.Code(codes.InvalidArgument),
			"invalid get author request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	author, err := i.authorUseCase.GetAuthorInfo(ctx, req.GetId())
	if err != nil {
		span.RecordError(err)
		return nil, i.ConvertErr(err)
	}

	i.logger.Info("finish to get author info",
		zap.String("layer", "controller"),
		zap.String("trace_id", span.SpanContext().TraceID().String()),
	)

	return &library.GetAuthorInfoResponse{
		Id:   author.ID,
		Name: author.Name,
	}, nil
}
