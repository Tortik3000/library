package outbox

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/project/library/config"
	"github.com/project/library/internal/usecase/repository"
)

var (
	outboxTasksFailedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "outbox_tasks_failed_total",
			Help: "Total number of failed outbox message processing attempts",
		},
		[]string{"kind"},
	)

	outboxTasksDurationTotal = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "outbox_tasks_duration_ms",
			Help:    "Duration of process outbox tasks in ms",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"kind"},
	)
)

func init() {
	prometheus.MustRegister(outboxTasksDurationTotal)
	prometheus.MustRegister(outboxTasksFailedTotal)
}

type GlobalHandler = func(kind repository.OutboxKind) (KindHandler, error)
type KindHandler = func(ctx context.Context, data []byte) error

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
	for {
		time.Sleep(waitTIme)

		if !o.cfg.Outbox.Enabled {
			continue
		}

		err := o.transactor.WithTx(ctx, func(ctx context.Context) error {
			messages, err := o.outboxRepository.GetMessages(
				ctx, batchSize, inProgressTTL)

			if err != nil {
				o.logger.Error("can not fetch messages from outbox",
					zap.Error(err))
				return err
			}

			successKeys := make([]string, 0, len(messages))

			for i := 0; i < len(messages); i++ {
				start := time.Now()
				message := messages[i]
				key := message.IdempotencyKey

				var kindHandler KindHandler
				kindHandler, err = o.globalHandler(message.Kind)

				if err != nil {
					outboxTasksFailedTotal.WithLabelValues(message.Kind.String()).Inc()
					o.logger.Error("unexpected kind", zap.Error(err))
					continue
				}

				err = kindHandler(ctx, message.RawData)

				if err != nil {
					outboxTasksFailedTotal.WithLabelValues(message.Kind.String()).Inc()
					o.logger.Error("kind error", zap.Error(err))
					continue
				}

				successKeys = append(successKeys, key)
				outboxTasksDurationTotal.WithLabelValues(message.Kind.String()).
					Observe(float64(time.Since(start).Milliseconds()))
			}

			err = o.outboxRepository.MarkAsProcessed(ctx, successKeys)
			if err != nil {
				o.logger.Error("mark as processed outbox error",
					zap.Error(err))
				return err
			}

			o.logger.Info("mark as processed", zap.Int("entities", len(successKeys)))

			return nil
		})

		if err != nil {
			o.logger.Error("worker stage error",
				zap.Error(err))
		}
	}
}
