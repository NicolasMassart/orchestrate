package jobs

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/tx-listener/store"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
)

const commitJobUseCaseComponent = "tx-listener.use-case.completed-job"

type completedJob struct {
	pendingJobState store.PendingJob
	messengerState  store.Message
	logger          *log.Logger
}

func CompletedUseCase(pendingJobState store.PendingJob, messengerState store.Message, logger *log.Logger) usecases.CompletedJob {
	consumer := &completedJob{
		pendingJobState: pendingJobState,
		messengerState:  messengerState,
		logger:          logger.SetComponent(commitJobUseCaseComponent),
	}

	return consumer
}

func (c *completedJob) Execute(ctx context.Context, job *entities.Job) error {
	childrenUUIDs := c.pendingJobState.GetChildrenJobUUIDs(ctx, job.UUID)
	for _, childrenUUID := range childrenUUIDs {
		err := c.execute(ctx, childrenUUID)
		if err != nil {
			return err
		}
	}

	return c.execute(ctx, job.UUID)
}

func (c *completedJob) execute(ctx context.Context, jobUUID string) error {
	logger := c.logger.WithField("job", jobUUID)
	idx, err := c.messengerState.MarkJobMessage(ctx, jobUUID)
	if err != nil {
		logger.WithError(err).Error("failed to mark job message")
		return err
	}

	logger.WithField("position", idx).Debug("job was appended in queue")
	if idx == 0 {
		var msg *entities.Message
		msg, err = c.messengerState.GetJobMessage(ctx, jobUUID)
		if err != nil {
			logger.WithError(err).Error("failed to retrieve job message")
			return errors.FromError(err).ExtendComponent(failedSessionJobComponent)
		}

		err = c.commitJobMessage(ctx, jobUUID, msg)
		if err != nil {
			return err
		}
	}

	err = c.pendingJobState.Remove(ctx, jobUUID)
	if err != nil {
		logger.WithError(err).Error("failed to remove pending job")
		return err
	}

	return nil
}

func (c *completedJob) commitJobMessage(ctx context.Context, jobUUID string, msg *entities.Message) error {
	logger := c.logger.WithField("offset", jobUUID)
	err := msg.Commit()
	if err != nil {
		logger.WithField("offset", msg.Offset+1).WithError(err).Error("failed to commit job message")
		return errors.FromError(err).ExtendComponent(failedSessionJobComponent)
	}

	err = c.messengerState.RemoveJobMessage(ctx, jobUUID)
	if err != nil {
		logger.WithError(err).Error("failed to remove job message")
		return errors.FromError(err).ExtendComponent(failedSessionJobComponent)
	}

	nextJobUUID, nextMsg, err := c.messengerState.GetMarkedJobMessageByOffset(ctx, msg.Offset+1)
	if err != nil {
		logger.WithError(err).Error("failed to get job message by offset")
		return errors.FromError(err).ExtendComponent(failedSessionJobComponent)
	}

	if nextMsg != nil {
		return c.commitJobMessage(ctx, nextJobUUID, nextMsg)
	}

	logger.WithField("offset", msg.Offset).Debug("job message committed successfully")
	return nil
}
