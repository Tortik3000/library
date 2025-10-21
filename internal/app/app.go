package app

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpcruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/project/library/config"
	"github.com/project/library/db"
	generated "github.com/project/library/generated/api/library"
	"github.com/project/library/internal/controller"
	"github.com/project/library/internal/usecase/library"
	"github.com/project/library/internal/usecase/repository"
)

const (
	gracefulShutdownTimeout = 5 * time.Second
	tableMetricsInterval    = time.Minute
)

func Run(
	logger *zap.Logger,
	cfg *config.Config,
) {
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

	tables := []string{"author", "book", "author_book"}
	go startTableMetricsCollector(ctx, dbPool, tables, tableMetricsInterval)

	<-ctx.Done()
	time.Sleep(gracefulShutdownTimeout)
}

func runRest(
	ctx context.Context,
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

func runGrpc(
	cfg *config.Config,
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
		grpc.StatsHandler(
			otelgrpc.NewServerHandler(
				otelgrpc.WithTracerProvider(otel.GetTracerProvider()),
			),
		),
		grpc.UnaryInterceptor(grpcMetricsInterceptor),
	)
	reflection.Register(s)

	generated.RegisterLibraryServer(s, libraryService)
	logger.Info("grpc server listening at port", zap.String("port", port))

	if err = s.Serve(lis); err != nil {
		logger.Error("grpc server listen error", zap.Error(err))
	}
}
