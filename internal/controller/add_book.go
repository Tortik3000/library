package controller

import (
	"context"
	codes2 "go.opentelemetry.io/otel/codes"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/project/library/generated/api/library"
)

var (
	AddBookDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "library_add_book_duration_ms",
		Help:    "Duration of AddBook in ms",
		Buckets: prometheus.DefBuckets,
	})

	AddBookRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "library_add_book_requests_total",
		Help: "Total number of AddBook requests",
	})
)

func init() {
	prometheus.MustRegister(AddBookRequests)
	prometheus.MustRegister(AddBookDuration)
}

func (i *impl) AddBook(
	ctx context.Context,
	req *library.AddBookRequest,
) (*library.AddBookResponse, error) {
	AddBookRequests.Inc()
	start := time.Now()

	defer func() {
		AddBookDuration.Observe(float64(time.Since(start).Milliseconds()))
	}()

	span := trace.SpanFromContext(ctx)
	defer span.End()

	i.logger.Info("start to add book",
		zap.String("layer", "controller"),
		zap.String("book_name", req.GetName()),
		zap.String("trace_id", span.SpanContext().TraceID().String()),
	)

	if err := req.ValidateAll(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes2.Code(codes.InvalidArgument),
			"invalid add book request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	book, err := i.booksUseCase.AddBook(ctx, req.GetName(), req.GetAuthorId())
	if err != nil {
		span.RecordError(err)
		return nil, i.ConvertErr(err)
	}

	i.logger.Info("finish to add book",
		zap.String("layer", "controller"),
		zap.String("trace_id", span.SpanContext().TraceID().String()),
	)

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
