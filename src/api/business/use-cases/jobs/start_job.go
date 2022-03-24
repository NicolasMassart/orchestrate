package jobs

import (
	"context"
	"time"

	"github.com/Shopify/sarama"
	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/metrics"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/entities"
	pkgsarama "github.com/consensys/orchestrate/src/infra/broker/sarama"
)

const startJobComponent = "use-cases.start-job"

// startJobUseCase is a use case to start a transaction job
type startJobUseCase struct {
	db            store.DB
	kafkaProducer sarama.SyncProducer
	topicsCfg     *pkgsarama.KafkaTopicConfig
	metrics       metrics.TransactionSchedulerMetrics
	logger        *log.Logger
}

// NewStartJobUseCase creates a new StartJobUseCase
func NewStartJobUseCase(
	db store.DB,
	kafkaProducer sarama.SyncProducer,
	topicsCfg *pkgsarama.KafkaTopicConfig,
	m metrics.TransactionSchedulerMetrics,
) usecases.StartJobUseCase {
	return &startJobUseCase{
		db:            db,
		kafkaProducer: kafkaProducer,
		topicsCfg:     topicsCfg,
		metrics:       m,
		logger:        log.NewLogger().SetComponent(startJobComponent),
	}
}

// Execute sends a job to the Kafka topic
func (uc *startJobUseCase) Execute(ctx context.Context, jobUUID string, userInfo *multitenancy.UserInfo) error {
	logger := uc.logger.WithContext(ctx).WithField("job", jobUUID)
	logger.Debug("starting job")

	curJob, err := uc.db.Job().FindOneByUUID(ctx, jobUUID, userInfo.AllowedTenants, userInfo.Username, false)
	if err != nil {
		return errors.FromError(err).ExtendComponent(startJobComponent)
	}

	prevJobUpdateAt := curJob.UpdatedAt
	nextJob := &entities.Job{
		UUID:   jobUUID,
		Status: entities.StatusStarted,
	}
	jobLog := &entities.Log{
		Status: entities.StatusStarted,
	}

	err = uc.db.Job().Update(ctx, nextJob, jobLog)
	if err != nil {
		return err
	}

	uc.addMetrics(time.Since(prevJobUpdateAt), curJob.Status, jobLog.Status, curJob.ChainUUID)

	err = sendEnvelope(uc.kafkaProducer, uc.topicsCfg.Sender, curJob)
	if err != nil {
		logger.WithError(err).Error("failed to send start job envelope")
		return err
	}

	logger.Info("job started successfully")
	return nil
}

func (uc *startJobUseCase) addMetrics(elapseTime time.Duration, previousStatus, nextStatus entities.JobStatus, chainUUID string) {
	baseLabels := []string{
		"chain_uuid", chainUUID,
	}

	uc.metrics.JobsLatencyHistogram().With(append(baseLabels,
		"prev_status", string(previousStatus),
		"status", string(nextStatus),
	)...).Observe(elapseTime.Seconds())
}
