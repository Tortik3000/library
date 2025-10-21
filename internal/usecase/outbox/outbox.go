package outbox

import (
	"context"
	"time"

	"github.com/project/library/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/project/library/config"
	"github.com/project/library/internal/usecase/repository"
)

type GlobalHandler = func(kind repository.OutboxKind) (KindHandler, error)
type KindHandler = func(ctx context.Context, data []byte) error

var tracer = otel.Tracer("library-service")

type Outbox interface {
	Start(ctx context.Context, workers int, batchSize int,
		waitTime time.Duration, inProgressTTL time.Duration)
}

var _ Outbox = (*outboxImpl)(nil)

type outboxImpl struct {
	logger           *zap.Logger
	outboxRepository repository.OutboxRepository
	globalHandler    GlobalHandler
	cfg              *config.Config
	transactor       repository.Transactor
}

func New(
	logger *zap.Logger,
	outboxRepository repository.OutboxRepository,
	globalHandler GlobalHandler,
	cfg *config.Config,
	transactor repository.Transactor,
) *outboxImpl {
	return &outboxImpl{
		logger:           logger,
		outboxRepository: outboxRepository,
		globalHandler:    globalHandler,
		cfg:              cfg,
		transactor:       transactor,
	}
}

func (o *outboxImpl) Start(
	ctx context.Context,
	workers int, batchSize int,
	waitTime time.Duration,
	inProgressTTL time.Duration,
) {
	for workerID := 1; workerID <= workers; workerID++ {
		go o.worker(ctx, batchSize, waitTime, inProgressTTL)
	}
}

func (o *outboxImpl) worker(
	ctx context.Context,
	batchSize int,
	waitTIme time.Duration,
	inProgressTTL time.Duration,
) {
	log := o.logger.With(
		zap.String("layer", "outbox"))
	for {
		time.Sleep(waitTIme)

		if !o.cfg.Outbox.Enabled {
			continue
		}

		err := o.transactor.WithTx(ctx, func(ctx context.Context) error {
			messages, err := o.outboxRepository.GetMessages(
				ctx, batchSize, inProgressTTL)

			if err != nil {
				log.Error("can not fetch messages from outbox",
					zap.Error(err))
				return err
			}

			successKeys := make([]string, 0, len(messages))

			for _, message := range messages {
				start := time.Now()
				key := message.IdempotencyKey
				kind := message.Kind.String()

				traceID, parseErr := trace.TraceIDFromHex(message.TraceID)
				if parseErr != nil {
					log.Warn("invalid trace_id",
						zap.String("trace_id", message.TraceID),
						zap.Error(parseErr))
				}

				var eventCtx context.Context
				if parseErr == nil {
					parentSC := trace.NewSpanContext(trace.SpanContextConfig{
						TraceID:    traceID,
						SpanID:     trace.SpanID{},
						TraceFlags: trace.FlagsSampled,
						Remote:     true,
					})
					eventCtx = trace.ContextWithRemoteSpanContext(context.Background(), parentSC)
				} else {
					eventCtx = context.Background()
				}

				eventCtx, span := tracer.Start(eventCtx, "ProcessOutboxEvent")
				span.SetAttributes(
					attribute.String("outbox.id", message.IdempotencyKey),
					attribute.String("outbox.kind", message.Kind.String()),
				)

				var kindHandler KindHandler
				kindHandler, err = o.globalHandler(message.Kind)

				if err != nil {
					log.Error("unexpected kind",
						zap.String("trace_id", traceID.String()),
						zap.String("span_id", span.SpanContext().SpanID().String()),
						zap.Error(err))
					metrics.OutboxTasksFailed.WithLabelValues(kind).Inc()
					span.End()
					continue
				}

				err = kindHandler(ctx, message.RawData)

				duration := time.Since(start).Seconds()
				metrics.OutboxTaskProcessingDuration.WithLabelValues(kind).Observe(duration)

				if err != nil {
					log.Error("kind error",
						zap.String("trace_id", traceID.String()),
						zap.String("span_id", span.SpanContext().SpanID().String()),
						zap.Error(err))
					metrics.OutboxTasksFailed.WithLabelValues(kind).Inc()
					span.End()
					continue
				}

				o.logger.Info("outbox worker executing",
					zap.String("trace_id", traceID.String()),
					zap.String("span_id", span.SpanContext().SpanID().String()),
					zap.String("kind", message.Kind.String()),
					zap.String("idempotency_key", message.IdempotencyKey))

				successKeys = append(successKeys, key)
				metrics.OutboxTasksProcessed.WithLabelValues(kind).Inc()
				span.End()
			}

			err = o.outboxRepository.MarkAsProcessed(ctx, successKeys)
			if err != nil {
				o.logger.Error("mark as processed outbox error", zap.Error(err))
				return err
			}

			return nil
		})

		if err != nil {
			o.logger.Error("worker transaction error", zap.Error(err))
			return
		}
	}
}
