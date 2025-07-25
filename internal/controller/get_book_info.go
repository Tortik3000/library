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
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/project/library/generated/api/library"
)

var (
	GetBookInfoDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "library_get_book_duration_ms",
		Help:    "Duration of GetBookInfo in ms",
		Buckets: prometheus.DefBuckets,
	})

	GetBookInfoRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "library_get_book_requests_total",
		Help: "Total number of GetBookInfo requests",
	})
)

func init() {
	prometheus.MustRegister(GetBookInfoRequests)
	prometheus.MustRegister(GetBookInfoDuration)
}

func (i *impl) GetBookInfo(
	ctx context.Context,
	req *library.GetBookInfoRequest,
) (*library.GetBookInfoResponse, error) {
	GetBookInfoRequests.Inc()
	start := time.Now()

	defer func() {
		GetBookInfoDuration.Observe(time.Since(start).Seconds())
	}()

	span := trace.SpanFromContext(ctx)
	defer span.End()

	i.logger.Info("start to get book info",
		zap.String("layer", "controller"),
		zap.String("book_id", req.Id),
		zap.String("trace_id", span.SpanContext().TraceID().String()),
	)

	if err := req.ValidateAll(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes2.Code(codes.InvalidArgument),
			"invalid get book request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	book, err := i.booksUseCase.GetBook(ctx, req.GetId())
	if err != nil {
		span.RecordError(err)
		return nil, i.ConvertErr(err)
	}

	i.logger.Info("finish to get book info",
		zap.String("layer", "controller"),
		zap.String("trace_id", span.SpanContext().TraceID().String()),
	)

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
