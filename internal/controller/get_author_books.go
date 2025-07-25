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
	GerAuthorBooksDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "library_get_author_books_duration_ms",
		Help:    "Duration of GetAuthorBooks in ms",
		Buckets: prometheus.DefBuckets,
	})

	GerAuthorBooksRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "library_get_author_books_requests_total",
		Help: "Total number of GetAuthorBooks requests",
	})
)

func init() {
	prometheus.MustRegister(GerAuthorBooksRequests)
	prometheus.MustRegister(GerAuthorBooksDuration)
}

func (i *impl) GetAuthorBooks(
	req *library.GetAuthorBooksRequest,
	server library.Library_GetAuthorBooksServer,
) error {
	GerAuthorBooksRequests.Inc()
	start := time.Now()

	defer func() {
		GerAuthorBooksDuration.Observe(float64(time.Since(start).Milliseconds()))
	}()

	span := trace.SpanFromContext(server.Context())
	defer span.End()

	i.logger.Info("start to get author's books",
		zap.String("layer", "controller"),
		zap.String("author_id", req.GetAuthorId()),
		zap.String("trace_id", span.SpanContext().TraceID().String()),
	)

	if err := req.ValidateAll(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes2.Code(codes.InvalidArgument),
			"invalid get author's books request")
		return status.Error(codes.InvalidArgument, err.Error())
	}

	books, err := i.authorUseCase.GetAuthorBooks(
		context.Background(), req.GetAuthorId())
	if err != nil {
		span.RecordError(err)
		return i.ConvertErr(err)
	}

	for _, book := range books {
		err := server.Send(&library.Book{
			Id:        book.ID,
			Name:      book.Name,
			AuthorId:  book.AuthorIDs,
			CreatedAt: timestamppb.New(book.CreatedAt),
			UpdatedAt: timestamppb.New(book.UpdatedAt),
		})
		if err != nil {
			return err
		}
	}

	i.logger.Info("finish to get author's books",
		zap.String("layer", "controller"),
		zap.String("trace_id", span.SpanContext().TraceID().String()),
	)

	return nil
}
