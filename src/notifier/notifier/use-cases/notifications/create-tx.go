package notifications

import (
	"context"
	"time"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/notifier/store"

	"github.com/gofrs/uuid"

	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/entities"
	usecases "github.com/consensys/orchestrate/src/notifier/notifier/use-cases"
)

const createTxComponent = "use-cases.notifier.create_tx"

type createTxUseCase struct {
	logger *log.Logger
	db     store.NotificationAgent
}

func NewCreateTransactionUseCase(db store.NotificationAgent) usecases.CreateTxNotificationUseCase {
	return &createTxUseCase{
		db:     db,
		logger: log.NewLogger().SetComponent(createTxComponent),
	}
}

func (uc *createTxUseCase) Execute(ctx context.Context, job *entities.Job, errStr string) (*entities.Notification, error) {
	logger := uc.logger.WithContext(ctx)

	notif := &entities.Notification{
		SourceUUID: job.ScheduleUUID,
		SourceType: entities.NotificationSourceTypeJob,
		Status:     entities.NotificationStatusPending,
		UUID:       uuid.Must(uuid.NewV4()).String(),
		Type:       jobStatusToNotificationType(job.Status),
		APIVersion: "v1",
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
		Error:      errStr,
	}

	notif, err := uc.db.Insert(ctx, notif)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(createTxComponent)
	}
	notif.Job = job

	logger.WithField("notification", notif.UUID).Debug("transaction notification created successfully")
	return notif, nil
}

func jobStatusToNotificationType(jobStatus entities.JobStatus) entities.NotificationType {
	switch jobStatus {
	case entities.StatusFailed:
		return entities.NotificationTypeTxFailed
	case entities.StatusMined:
		return entities.NotificationTypeTxMined
	}

	return ""
}
