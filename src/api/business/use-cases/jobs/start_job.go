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

	job, err := uc.db.Job().FindOneByUUID(ctx, jobUUID, userInfo.AllowedTenants, userInfo.Username, false)
	if err != nil {
		return errors.FromError(err).ExtendComponent(startJobComponent)
	}

	if !canUpdateStatus(entities.StatusStarted, job.Status) {
		errMessage := "cannot start job at the current status"
		logger.WithField("status", job.Status).WithField("next_status", entities.StatusStarted).Error(errMessage)
		return errors.InvalidStateError(errMessage)
	}

	err = uc.updateStatus(ctx, job, entities.StatusStarted, "")
	if err != nil {
		return errors.FromError(err).ExtendComponent(startJobComponent)
	}

	partition, offset, err := SendJobMessage(job, uc.kafkaProducer, uc.topicsCfg.Sender)
	if err != nil {
		errMsg := "failed to send job message"
		_ = uc.updateStatus(ctx, job, entities.StatusFailed, errMsg)
		logger.WithError(err).Error(errMsg)
		return errors.FromError(err).ExtendComponent(startJobComponent)
	}

	logger.WithField("partition", partition).WithField("offset", offset).Info("job started successfully")
	return nil
}

func (uc *startJobUseCase) updateStatus(ctx context.Context, job *entities.Job, status entities.JobStatus, msg string) error {
	prevUpdatedAt := job.UpdatedAt
	prevStatus := job.Status

	job.Status = status
	jobLog := &entities.Log{
		Status:  status,
		Message: msg,
	}

	err := uc.db.RunInTransaction(ctx, func(dbtx store.DB) error {
		if der := dbtx.Job().Update(ctx, job); der != nil {
			return der
		}

		_, der := dbtx.Log().Insert(ctx, jobLog, job.UUID)
		if der != nil {
			return errors.FromError(der).ExtendComponent(startJobComponent)
		}

		return nil
	})
	if err != nil {
		return err
	}

	uc.addMetrics(job.UpdatedAt.Sub(prevUpdatedAt), prevStatus, status, job.ChainUUID)
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