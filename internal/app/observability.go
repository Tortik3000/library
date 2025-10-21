package app

import (
	"context"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/grafana/pyroscope-go"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"github.com/project/library/metrics"
)

func startTableMetricsCollector(ctx context.Context, db *pgxpool.Pool, tables []string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	lastRowCounts := map[string]int64{}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, table := range tables {
				var count int64
				row := db.QueryRow(ctx, "SELECT COUNT(*) FROM "+table)
				if err := row.Scan(&count); err != nil {
					continue
				}
				metrics.DBTableRowsCount.WithLabelValues(table).Set(float64(count))

				last, ok := lastRowCounts[table]
				if ok && count > last {
					metrics.DBTableInsertRate.WithLabelValues(table).Add(float64(count - last))
				}
				lastRowCounts[table] = count
			}
		}
	}
}

func runPyroscope(logger *zap.Logger, url string) {
	logger.Info("starting pyroscope server", zap.String("address", url))

	runtime.SetMutexProfileFraction(1)
	runtime.SetBlockProfileRate(1)
	_, err := pyroscope.Start(pyroscope.Config{
		ApplicationName: "leak.app",
		ServerAddress:   url,
		Logger:          pyroscope.StandardLogger,
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,

			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		},
	})

	if err != nil {
		logger.Fatal("can not set up pyroscope", zap.Error(err))
	}
}

func runMetricsServer(logger *zap.Logger, port string) {
	logger.Info("starting metrics server", zap.String("port", port))
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		logger.Fatal("can not start metrics server", zap.Error(err))
	}
}

func initTracer(logger *zap.Logger, url string) func(context.Context) error {
	logger.Info("starting tracer server", zap.String("address", url))

	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))

	if err != nil {
		logger.Fatal("can not create jaeger collector", zap.Error(err))
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("library-service"),
		)),
	)

	otel.SetTracerProvider(tp)

	return tp.Shutdown
}

func grpcMetricsInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	duration := time.Since(start).Seconds()

	parts := strings.Split(info.FullMethod, "/")
	serviceName := ""
	methodName := ""
	if len(parts) == 3 {
		serviceName = parts[1]
		methodName = parts[2]
	}

	codeStr := "OK"
	if err != nil {
		st, _ := status.FromError(err)
		codeStr = st.Code().String()
	}

	metrics.GRPCRequestsTotal.WithLabelValues(serviceName, methodName, codeStr).Inc()
	metrics.GRPCRequestDuration.WithLabelValues(serviceName, methodName).Observe(duration)

	return resp, err
}
