package app

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/grafana/pyroscope-go"
	grpcruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/project/library/db"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/exporters/jaeger"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/project/library/config"
	generated "github.com/project/library/generated/api/library"
	"github.com/project/library/internal/controller"
	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/library"
	"github.com/project/library/internal/usecase/outbox"
	"github.com/project/library/internal/usecase/repository"
)

const (
	timeSleepForTerminate = time.Second * 3
	MaxIdleConnections    = 100
	MaxConnectionsPerHost = 100
	IdleConnectionTimeout = 90 * time.Second
	TLSHandshakeTimeout   = 15 * time.Second
	ExpectContinueTimeout = 2 * time.Second
	Timeout               = 30 * time.Second
	KeepAlive             = 180 * time.Second
)

var (
	tableRowCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "db_table_row_count",
			Help: "Number of rows per table",
		},
		[]string{"table"},
	)

	tableInsertCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_table_insert_total",
			Help: "Total inserts per table",
		},
		[]string{"table"},
	)
)

func init() {
	prometheus.MustRegister(tableRowCount)
	prometheus.MustRegister(tableInsertCount)
}

func Run(logger *zap.Logger, cfg *config.Config) {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	shutdown := initTracer(logger, cfg.Observability.JaegerURL)
	defer func() {
		err := shutdown(ctx)

		if err != nil {
			logger.Error("can not shutdown jaeger collector", zap.Error(err))
		}
	}()

	go runMetricsServer(logger, cfg.Observability.MetricsPort)
	runPyroscope(logger, cfg.Observability.PyroscopeUrl)

	dbPool, err := pgxpool.New(ctx, cfg.PG.URL)
	if err != nil {
		logger.Error("can not create pgxpool", zap.Error(err))
		return
	}

	defer dbPool.Close()

	db.SetupPostgres(dbPool, logger)

	repo := repository.NewPostgresRepository(dbPool, logger)
	outboxRepository := repository.NewOutbox(dbPool, logger)
	transactor := repository.NewTransactor(dbPool, logger)
	runOutbox(ctx, cfg, logger, outboxRepository, transactor)

	useCases := library.New(logger, repo, repo, outboxRepository, transactor)
	ctrl := controller.New(logger, useCases, useCases)

	go runRest(ctx, cfg, logger)
	go runGrpc(cfg, logger, ctrl)
	go startTableMetricsCollector(ctx, dbPool, logger)

	<-ctx.Done()
	time.Sleep(timeSleepForTerminate)
}

func startTableMetricsCollector(ctx context.Context, db *pgxpool.Pool, logger *zap.Logger) {

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	lastRowCounts := map[string]int64{}

	tables := []string{"author", "book", "author_book"}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, table := range tables {
				var count int64
				err := db.QueryRow(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count)
				if err != nil {
					logger.Error("failed to count rows", zap.String("table", table), zap.Error(err))
					continue
				}
				tableRowCount.WithLabelValues(table).Set(float64(count))

				last, ok := lastRowCounts[table]
				if ok && count > last {
					tableInsertCount.WithLabelValues(table).Add(float64(count - last))
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

func runRest(ctx context.Context,
	cfg *config.Config,
	logger *zap.Logger,
) {
	mux := grpcruntime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	address := "localhost:" + cfg.GRPC.Port
	err := generated.RegisterLibraryHandlerFromEndpoint(ctx, mux, address, opts)
	if err != nil {
		logger.Error("can not register grpc gateway", zap.Error(err))
		return
	}

	gatewayPort := ":" + cfg.GatewayPort
	logger.Info("gateway listening at port",
		zap.String("port", gatewayPort))
	if err = http.ListenAndServe(gatewayPort, mux); err != nil {
		logger.Error("gateway listen error", zap.Error(err))
	}
}

func runGrpc(cfg *config.Config,
	logger *zap.Logger,
	libraryService generated.LibraryServer,
) {
	port := ":" + cfg.GRPC.Port
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logger.Error("can not open tcp socket", zap.Error(err))
		os.Exit(1)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(
			otelgrpc.UnaryServerInterceptor(
				otelgrpc.WithTracerProvider(otel.GetTracerProvider()),
			),
		),
		grpc.StreamInterceptor(
			otelgrpc.StreamServerInterceptor(
				otelgrpc.WithTracerProvider(otel.GetTracerProvider()),
			),
		),
	)
	reflection.Register(s)

	generated.RegisterLibraryServer(s, libraryService)

	logger.Info("grpc server listening at port", zap.String("port", port))

	if err = s.Serve(lis); err != nil {
		logger.Error("grpc server listen error", zap.Error(err))
	}
}

func runOutbox(
	ctx context.Context,
	cfg *config.Config,
	logger *zap.Logger,
	outboxRepository repository.OutboxRepository,
	transactor repository.Transactor,
) {
	dialer := &net.Dialer{
		Timeout:   Timeout,
		KeepAlive: KeepAlive,
	}

	transport := &http.Transport{
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          MaxIdleConnections,
		MaxConnsPerHost:       MaxConnectionsPerHost,
		IdleConnTimeout:       IdleConnectionTimeout,
		TLSHandshakeTimeout:   TLSHandshakeTimeout,
		ExpectContinueTimeout: ExpectContinueTimeout,
	}

	client := &http.Client{Transport: transport}

	globalHandler := globalOutboxHandler(
		client, cfg.Outbox.BookSendURL, cfg.Outbox.AuthorSendURL)
	outboxService := outbox.New(
		logger, outboxRepository, globalHandler, cfg, transactor)

	outboxService.Start(
		ctx,
		cfg.Outbox.Workers,
		cfg.Outbox.BatchSize,
		cfg.Outbox.WaitTimeMS,
		cfg.Outbox.InProgressTTLMS,
	)
}

func globalOutboxHandler(
	client *http.Client,
	bookURL string,
	authorURL string,
) outbox.GlobalHandler {
	return func(kind repository.OutboxKind) (outbox.KindHandler, error) {
		switch kind {
		case repository.OutboxKindBook:
			return bookOutboxHandler(client, bookURL), nil
		case repository.OutboxKindAuthor:
			return authorOutboxHandler(client, authorURL), nil
		default:
			return nil, fmt.Errorf("unsupported outbox kind: %d", kind)
		}
	}
}

func outboxHandler(
	client *http.Client,
	url string,
	unmarshalFunc func(data []byte) (string, error),
) outbox.KindHandler {
	return func(_ context.Context, data []byte) error {
		id, err := unmarshalFunc(data)
		if err != nil {
			return fmt.Errorf("can not deserialize data in outbox handler: %w", err)
		}

		resp, err := client.Post(url, "application/json", strings.NewReader(id))
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode < http.StatusOK || http.StatusMultipleChoices <= resp.StatusCode {
			return fmt.Errorf("request failed with status: %d", resp.StatusCode)
		}
		return nil
	}
}

func bookOutboxHandler(
	client *http.Client,
	url string,
) outbox.KindHandler {
	return outboxHandler(client, url, func(data []byte) (string, error) {
		book := entity.Book{}
		if err := json.Unmarshal(data, &book); err != nil {
			return "", err
		}
		return book.ID, nil
	})
}

func authorOutboxHandler(
	client *http.Client,
	url string,
) outbox.KindHandler {
	return outboxHandler(client, url, func(data []byte) (string, error) {
		author := entity.Author{}
		if err := json.Unmarshal(data, &author); err != nil {
			return "", err
		}
		return author.ID, nil
	})
}
