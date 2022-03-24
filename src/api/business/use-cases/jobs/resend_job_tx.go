package jobs

import (
	"context"

	"github.com/Shopify/sarama"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	pkgsarama "github.com/consensys/orchestrate/src/infra/broker/sarama"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/api/store"
)

const resendJobTxComponent = "use-cases.resend-job-tx"

type resendJobTxUseCase struct {
	db            store.DB
	kafkaProducer sarama.SyncProducer
	topicsCfg     *pkgsarama.KafkaTopicConfig
	logger        *log.Logger
}

func NewResendJobTxUseCase(db store.DB, kafkaProducer sarama.SyncProducer, topicsCfg *pkgsarama.KafkaTopicConfig) usecases.ResendJobTxUseCase {
	return &resendJobTxUseCase{
		db:            db,
		kafkaProducer: kafkaProducer,
		topicsCfg:     topicsCfg,
		logger:        log.NewLogger().SetComponent(resendJobTxComponent),
	}
}

// Execute sends a job to the Kafka topic
func (uc *resendJobTxUseCase) Execute(ctx context.Context, jobUUID string, userInfo *multitenancy.UserInfo) error {
	ctx = log.WithFields(ctx, log.Field("job", jobUUID))
	logger := uc.logger.WithContext(ctx)
	logger.Debug("resending job transaction")

	job, err := uc.db.Job().FindOneByUUID(ctx, jobUUID, userInfo.AllowedTenants, userInfo.Username, false)
	if err != nil {
		return errors.FromError(err).ExtendComponent(resendJobTxComponent)
	}

	job.InternalData.ParentJobUUID = jobUUID
	err = sendEnvelope(uc.kafkaProducer, uc.topicsCfg.Sender, job)
	if err != nil {
		logger.WithError(err).Error("failed to send resend job envelope")
		return err
	}

	logger.Info("transaction resent successfully")
	return nil
}
