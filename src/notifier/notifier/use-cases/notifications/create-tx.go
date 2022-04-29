package notifications

import (
	"context"
	"time"

	"github.com/gofrs/uuid"

	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/entities"
	usecases "github.com/consensys/orchestrate/src/notifier/notifier/use-cases"
)

const createTxComponent = "use-cases.notifier.create_tx"

type createTxUseCase struct {
	logger *log.Logger
}

func NewCreateTransactionUseCase() usecases.CreateTxNotificationUseCase {
	return &createTxUseCase{
		logger: log.NewLogger().SetComponent(createTxComponent),
	}
}

func (uc *createTxUseCase) Execute(ctx context.Context, job *entities.Job, errStr string) *entities.Notification {

	notif := &entities.Notification{
		SourceUUID: job.ScheduleUUID,
		Status:     entities.NotificationStatusPending,
		UUID:       uuid.Must(uuid.NewV4()).String(),
		Type:       jobStatusToNotificationType(job.Status),
		APIVersion: "v1",
		Job:        job,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
		Error:      errStr,
	}

	// TODO(dario): Save notification in DB

	uc.logger.WithField("notification", notif.UUID).Debug("transaction notification created successfully")
	return notif
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
