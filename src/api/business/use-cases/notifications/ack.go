package notifications

import (
	"context"

	"github.com/consensys/orchestrate/src/entities"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/store"
)

const ackNotifComponent = "use-cases.ack-notification"

type ackUseCase struct {
	db     store.NotificationAgent
	logger *log.Logger
}

func NewAckUseCase(db store.NotificationAgent) usecases.AckNotificationUseCase {
	return &ackUseCase{
		db:     db,
		logger: log.NewLogger().SetComponent(ackNotifComponent),
	}
}

func (uc *ackUseCase) Execute(ctx context.Context, uuid string) error {
	ctx = log.WithFields(ctx, log.Field("notification", uuid))
	logger := uc.logger.WithContext(ctx)
	logger.Debug("acknowledging notification")

	_, err := uc.db.Update(ctx, &entities.Notification{UUID: uuid, Status: entities.NotificationStatusSent})
	if err != nil {
		return errors.FromError(err).ExtendComponent(ackNotifComponent)
	}

	logger.Debug("notification acknowledged successfully")
	return nil
}
